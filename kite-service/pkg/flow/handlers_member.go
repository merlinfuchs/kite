package flow

import (
	"errors"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
)

func init() {
	registerHandlers(map[FlowNodeType]nodeHandler{
		FlowNodeTypeActionMemberEdit: executeActionMemberEdit,
		FlowNodeTypeActionMemberGet:  executeActionMemberGet,
		FlowNodeTypeActionUserGet:    executeActionUserGet,
	})
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
