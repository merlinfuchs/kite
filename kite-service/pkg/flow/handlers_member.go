package flow

import (
	"errors"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/kitecloud/kite/kite-service/pkg/guildevent"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
)

func init() {
	registerHandlers(map[FlowNodeType]nodeHandler{
		FlowNodeTypeActionMemberEdit: executeActionMemberEdit,
		FlowNodeTypeActionMemberGet:  executeActionMemberGet,
		FlowNodeTypeActionUserGet:    executeActionUserGet,

		FlowNodeTypeActionMemberPresenceGet: executeActionMemberPresenceGet,
	})
}

func executeActionMemberEdit(n *CompiledFlowNode, ctx *FlowContext) error {
	guildID, err := n.targetGuildID(ctx)
	if err != nil {
		return err
	}

	userID, err := ctx.EvalTemplate(n.Data.UserTarget)
	if err != nil {
		return err
	}

	auditLogReason, err := ctx.EvalTemplate(n.Data.AuditLogReason)
	if err != nil {
		return err
	}

	data := api.ModifyMemberData{
		AuditLogReason: api.AuditLogReason(auditLogReason.String()),
	}

	if n.Data.MemberData != nil {
		if n.Data.MemberData.Nick != nil {
			nick, err := eval.EvalTemplateToString(ctx, *n.Data.MemberData.Nick, ctx.EvalCtx)
			if err != nil {
				return err
			}

			data.Nick = &nick
		}
	}

	err = ctx.Discord.EditMember(
		ctx,
		guildID,
		discord.UserID(userID.Snowflake()),
		data,
	)
	if err != nil {
		return err
	}

	return n.ExecuteChildren(ctx)
}

func executeActionMemberGet(n *CompiledFlowNode, ctx *FlowContext) error {
	guildID, err := n.targetGuildID(ctx)
	if err != nil {
		return err
	}

	memberID, err := ctx.EvalTemplate(n.Data.UserTarget)
	if err != nil {
		return err
	}

	member, err := ctx.Discord.Member(ctx, guildID, discord.UserID(memberID.Snowflake()))
	if err != nil && !errors.Is(err, provider.ErrNotFound) {
		return err
	}

	if member != nil {
		ctx.StoreNodeResult(n, thing.NewDiscordMember(*member))
	} else {
		ctx.StoreNodeResult(n, thing.Null)
	}

	return n.ExecuteChildren(ctx)
}

// executeActionMemberPresenceGet stores the status and custom status of a
// member. Members Discord sent no presence for are offline.
func executeActionMemberPresenceGet(n *CompiledFlowNode, ctx *FlowContext) error {
	guildID, err := n.targetGuildID(ctx)
	if err != nil {
		return err
	}
	if !guildID.IsValid() {
		return errors.New("a server is required to get the status of a member")
	}

	userID, err := ctx.EvalTemplate(n.Data.UserTarget)
	if err != nil {
		return err
	}

	presence, err := ctx.Discord.MemberPresence(ctx, guildID, discord.UserID(userID.Snowflake()))
	if err != nil && !errors.Is(err, provider.ErrNotFound) {
		return err
	}

	p := discord.Presence{Status: discord.OfflineStatus}
	if presence != nil {
		p = *presence
	}

	customStatus, customStatusEmoji := guildevent.CustomStatus(p)
	ctx.StoreNodeResult(n, thing.NewObject(map[string]thing.Thing{
		"status":              thing.NewString(presenceStatus(p.Status)),
		"custom_status":       thing.NewString(customStatus),
		"custom_status_emoji": thing.NewString(customStatusEmoji),
		"activity":            thing.NewString(presenceActivity(p)),
	}))
	return n.ExecuteChildren(ctx)
}

// presenceActivity is the name of what a member is doing, like the game they
// play, empty if nothing. The custom status is its own field.
func presenceActivity(p discord.Presence) string {
	for _, a := range p.Activities {
		if a.Type != discord.CustomActivity {
			return a.Name
		}
	}
	return ""
}

func executeActionUserGet(n *CompiledFlowNode, ctx *FlowContext) error {
	userID, err := ctx.EvalTemplate(n.Data.UserTarget)
	if err != nil {
		return err
	}

	user, err := ctx.Discord.User(ctx, discord.UserID(userID.Snowflake()))
	if err != nil && !errors.Is(err, provider.ErrNotFound) {
		return err
	}

	if user != nil {
		ctx.StoreNodeResult(n, thing.NewDiscordUser(*user))
	} else {
		ctx.StoreNodeResult(n, thing.Null)
	}

	return n.ExecuteChildren(ctx)
}
