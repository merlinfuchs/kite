package flow

import (
	"fmt"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
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

	err := n.autoDeferInteraction(ctx)
	if err != nil {
		return traceError(n, err)
	}

	err = n.ExecuteChildren(ctx)
	if err != nil {
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
