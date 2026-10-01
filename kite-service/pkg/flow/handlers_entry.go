package flow

import (
	"context"
	"fmt"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
)

func init() {
	registerHandlers(map[FlowNodeType]nodeHandler{
		FlowNodeTypeEntryCommand:         executeEntryCommand,
		FlowNodeTypeEntryComponentButton: executeEntryCommand,
		FlowNodeTypeEntryEvent:           executeEntryEvent,
	})
}

func executeEntryCommand(n *CompiledFlowNode, ctx *FlowContext) error {
	if !ctx.IsEntry() {
		return fmt.Errorf("command entry isn't the entry node")
	}

	cooldown, err := n.checkCommandCooldown(ctx)
	if err != nil {
		createDefaultErrorResponse(ctx, err)
		return traceError(n, err)
	}
	if cooldown.onCooldown {
		return nil
	}

	err = n.autoDeferInteraction(ctx)
	if err != nil {
		cooldown.reset(ctx)
		return traceError(n, err)
	}

	err = n.ExecuteChildren(ctx)
	if err != nil {
		// A failed run shouldn't use up the cooldown.
		cooldown.reset(ctx)
		createDefaultErrorResponse(ctx, err)
		return traceError(n, err)
	}

	acknowledgeUnansweredComponent(ctx)
	return nil
}

func executeEntryEvent(n *CompiledFlowNode, ctx *FlowContext) error {
	if !ctx.IsEntry() {
		return fmt.Errorf("event entry isn't the entry node")
	}

	err := n.ExecuteChildren(ctx)
	if err != nil {
		return traceError(n, err)
	}

	return nil
}

// acknowledgeUnansweredComponent acknowledges a component interaction the flow
// finished without responding to, e.g. a button that only sends a channel
// message. Discord shows "This interaction failed" otherwise, and the
// auto-defer doesn't fire for flows that finish quickly.
func acknowledgeUnansweredComponent(ctx *FlowContext) {
	interaction := ctx.Data.Interaction()
	if interaction == nil {
		return
	}
	if _, ok := interaction.Data.(discord.ComponentInteraction); !ok {
		return
	}

	hasCreatedResponse, err := ctx.Discord.HasCreatedInteractionResponse(ctx, interaction.ID)
	if err != nil || hasCreatedResponse {
		return
	}

	_, _ = ctx.Discord.CreateInteractionResponse(ctx, interaction.ID, interaction.Token, api.InteractionResponse{
		Type: api.DeferredMessageUpdate,
	})
}

// commandCooldown is the outcome of checkCommandCooldown.
type commandCooldown struct {
	// onCooldown reports whether the command was on cooldown. If so, the
	// interaction has already been answered with the cooldown message.
	onCooldown bool
	// key and expiresAt identify the cooldown this run started, if any.
	key       string
	expiresAt time.Time
}

// reset ends the cooldown this run started, so a failed run doesn't use it up.
func (c commandCooldown) reset(ctx *FlowContext) {
	if c.key == "" {
		return
	}

	// Best effort: the flow already failed, and the cooldown expires anyway.
	// The flow may have failed because its context was cancelled, which
	// mustn't stop the reset.
	_ = ctx.Cooldown.Reset(context.WithoutCancel(ctx), c.key, c.expiresAt)
}

// checkCommandCooldown checks the cooldown option attached to n, if any. If
// the command is currently on cooldown, it has already replied to the
// interaction with the cooldown message, and the flow must stop without
// running its children or auto-deferring. Otherwise it starts a new cooldown.
func (n *CompiledFlowNode) checkCommandCooldown(ctx *FlowContext) (commandCooldown, error) {
	cooldownNode := n.CommandCooldown()
	if cooldownNode == nil {
		return commandCooldown{}, nil
	}

	interaction := ctx.Data.Interaction()
	if interaction == nil {
		return commandCooldown{}, nil
	}

	durationValue, err := ctx.EvalTemplate(cooldownNode.Data.CooldownDurationSeconds)
	if err != nil {
		return commandCooldown{}, traceError(cooldownNode, err)
	}

	duration, err := parseCooldownDuration(durationValue.String())
	if err != nil {
		return commandCooldown{}, traceError(cooldownNode, err)
	}

	key := cooldownKey(ctx, interaction.AppID.String(), cooldownNode)

	remaining, expiresAt, err := ctx.Cooldown.CheckAndStart(ctx, key, duration)
	if err != nil {
		return commandCooldown{}, traceError(cooldownNode, err)
	}

	if remaining <= 0 {
		return commandCooldown{key: key, expiresAt: expiresAt}, nil
	}

	if err := cooldownNode.respondCooldown(ctx, remaining); err != nil {
		return commandCooldown{}, traceError(cooldownNode, err)
	}

	return commandCooldown{onCooldown: true}, nil
}

// cooldownKey builds a key that's unique per app, per cooldown block, and per
// scope target -- so the same cooldown block on different bots, or different
// cooldown blocks on the same bot, never collide.
func cooldownKey(ctx *FlowContext, appID string, cooldownNode *CompiledFlowNode) string {
	switch cooldownNode.Data.CooldownScope {
	case CooldownScopeUser, "": // Empty is the default, per user.
		return appID + ":" + cooldownNode.ID + ":user:" + ctx.Data.UserID().String()
	case CooldownScopeServer:
		if guildID := ctx.Data.GuildID(); guildID != 0 {
			return appID + ":" + cooldownNode.ID + ":server:" + guildID.String()
		}
		// No server to key by in DMs, so fall back to a per-user cooldown.
		return appID + ":" + cooldownNode.ID + ":user:" + ctx.Data.UserID().String()
	default: // CooldownScopeGlobal
		return appID + ":" + cooldownNode.ID + ":global"
	}
}

// respondCooldown replies to the interaction with the cooldown message. The
// remaining seconds are made available to the message as
// {{var('cooldown_remaining')}}, the same way other computed values are
// exposed to templates.
func (n *CompiledFlowNode) respondCooldown(ctx *FlowContext, remaining time.Duration) error {
	interaction := ctx.Data.Interaction()
	if interaction == nil {
		return nil
	}

	remainingSeconds := int64(remaining / time.Second)
	if remaining%time.Second != 0 {
		remainingSeconds++
	}
	ctx.SetTemporary("cooldown_remaining", thing.NewInt(remainingSeconds))

	rawMessage := n.Data.CooldownMessage
	if rawMessage == "" {
		rawMessage = "You're on cooldown. Try again in {{var('cooldown_remaining')}} seconds."
	}

	content, err := ctx.EvalTemplate(rawMessage)
	if err != nil {
		return err
	}

	respData := api.InteractionResponseData{
		Content: option.NewNullableString(content.String()),
		Flags:   discord.EphemeralMessage,
	}

	hasCreatedResponse, _ := ctx.Discord.HasCreatedInteractionResponse(ctx, interaction.ID)
	if hasCreatedResponse {
		_, err = ctx.Discord.CreateInteractionFollowup(ctx, interaction.AppID, interaction.Token, respData)
	} else {
		_, err = ctx.Discord.CreateInteractionResponse(ctx, interaction.ID, interaction.Token, api.InteractionResponse{
			Type: api.MessageInteractionWithSource,
			Data: &respData,
		})
	}

	return err
}
