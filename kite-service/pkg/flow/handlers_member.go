package flow

import (
	"errors"
	"fmt"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
)

func init() {
	registerHandlers(map[FlowNodeType]nodeHandler{
		FlowNodeTypeActionMemberVoiceEdit: executeActionMemberVoiceEdit,
		FlowNodeTypeActionMemberEdit:      executeActionMemberEdit,
		FlowNodeTypeActionMemberGet:       executeActionMemberGet,
		FlowNodeTypeActionUserGet:         executeActionUserGet,
	})
}

func executeActionMemberVoiceEdit(n *CompiledFlowNode, ctx *FlowContext) error {
	guildID, err := n.targetGuildID(ctx)
	if err != nil {
		return traceError(n, err)
	}

	userID, err := ctx.EvalTemplate(n.Data.UserTarget)
	if err != nil {
		return traceError(n, err)
	}

	auditLogReason, err := ctx.EvalTemplate(n.Data.AuditLogReason)
	if err != nil {
		return traceError(n, err)
	}

	mute, err := n.Data.MemberVoiceMute.Option()
	if err != nil {
		return traceError(n, err)
	}

	deaf, err := n.Data.MemberVoiceDeaf.Option()
	if err != nil {
		return traceError(n, err)
	}

	data := api.ModifyMemberData{
		Mute:           mute,
		Deaf:           deaf,
		AuditLogReason: api.AuditLogReason(auditLogReason.String()),
	}

	// An empty channel target leaves the member where they are. A set one
	// that doesn't resolve to an ID is an error, as a zero ID would be left
	// out of the request and the block would report a move that never
	// happened.
	if n.Data.ChannelTarget != "" {
		channelTarget, err := ctx.EvalTemplate(n.Data.ChannelTarget)
		if err != nil {
			return traceError(n, err)
		}

		channelID := discord.ChannelID(channelTarget.Snowflake())
		if !channelID.IsValid() {
			return traceError(n, fmt.Errorf("channel %q is not a valid channel ID", channelTarget.String()))
		}
		data.VoiceChannel = channelID
	}

	if data.Mute == nil && data.Deaf == nil && data.VoiceChannel == 0 {
		return traceError(n, errors.New("nothing to change: set mute, deafen or a channel to move the member to"))
	}

	err = ctx.Discord.EditMember(
		ctx,
		guildID,
		discord.UserID(userID.Snowflake()),
		data,
	)
	if err != nil {
		return traceError(n, err)
	}

	return n.ExecuteChildren(ctx)
}

func executeActionMemberEdit(n *CompiledFlowNode, ctx *FlowContext) error {
	guildID, err := n.targetGuildID(ctx)
	if err != nil {
		return traceError(n, err)
	}

	userID, err := ctx.EvalTemplate(n.Data.UserTarget)
	if err != nil {
		return traceError(n, err)
	}

	auditLogReason, err := ctx.EvalTemplate(n.Data.AuditLogReason)
	if err != nil {
		return traceError(n, err)
	}

	data := api.ModifyMemberData{
		AuditLogReason: api.AuditLogReason(auditLogReason.String()),
	}

	if n.Data.MemberData != nil {
		if n.Data.MemberData.Nick != nil {
			nick, err := eval.EvalTemplateToString(ctx, *n.Data.MemberData.Nick, ctx.EvalCtx)
			if err != nil {
				return traceError(n, err)
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
		return traceError(n, err)
	}

	return n.ExecuteChildren(ctx)
}

func executeActionMemberGet(n *CompiledFlowNode, ctx *FlowContext) error {
	guildID, err := n.targetGuildID(ctx)
	if err != nil {
		return traceError(n, err)
	}

	memberID, err := ctx.EvalTemplate(n.Data.UserTarget)
	if err != nil {
		return traceError(n, err)
	}

	member, err := ctx.Discord.Member(ctx, guildID, discord.UserID(memberID.Snowflake()))
	if err != nil && !errors.Is(err, provider.ErrNotFound) {
		return traceError(n, err)
	}

	if member != nil {
		ctx.StoreNodeResult(n, thing.NewDiscordMember(*member))
	} else {
		ctx.StoreNodeResult(n, thing.Null)
	}

	return n.ExecuteChildren(ctx)
}

func executeActionUserGet(n *CompiledFlowNode, ctx *FlowContext) error {
	userID, err := ctx.EvalTemplate(n.Data.UserTarget)
	if err != nil {
		return traceError(n, err)
	}

	user, err := ctx.Discord.User(ctx, discord.UserID(userID.Snowflake()))
	if err != nil && !errors.Is(err, provider.ErrNotFound) {
		return traceError(n, err)
	}

	if user != nil {
		ctx.StoreNodeResult(n, thing.NewDiscordUser(*user))
	} else {
		ctx.StoreNodeResult(n, thing.Null)
	}

	return n.ExecuteChildren(ctx)
}
