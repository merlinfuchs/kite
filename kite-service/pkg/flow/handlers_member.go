package flow

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
)

func init() {
	registerHandlers(map[FlowNodeType]nodeHandler{
		FlowNodeTypeActionMemberEdit:  executeActionMemberEdit,
		FlowNodeTypeActionMemberGet:   executeActionMemberGet,
		FlowNodeTypeActionMemberPrune: executeActionMemberPrune,
		FlowNodeTypeActionUserGet:     executeActionUserGet,
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

func executeActionMemberPrune(n *CompiledFlowNode, ctx *FlowContext) error {
	guildID, err := n.targetGuildID(ctx)
	if err != nil {
		return traceError(n, err)
	}

	days, err := n.memberPruneDays(ctx)
	if err != nil {
		return traceError(n, err)
	}

	auditLogReason, err := ctx.EvalTemplate(n.Data.AuditLogReason)
	if err != nil {
		return traceError(n, err)
	}

	// Without included roles Discord only removes members that have no
	// roles, which keeps the block from clearing out active members.
	pruned, err := ctx.Discord.PruneMembers(ctx, guildID, api.PruneData{
		Days:           days,
		ReturnCount:    true,
		AuditLogReason: api.AuditLogReason(auditLogReason.String()),
	})
	if err != nil {
		return traceError(n, err)
	}

	ctx.StoreNodeResult(n, thing.NewInt(pruned))
	return n.ExecuteChildren(ctx)
}

// memberPruneMaxDays is the most days of inactivity Discord accepts for a prune.
const memberPruneMaxDays = 30

// memberPruneDays evaluates how many days members must have been inactive to
// be pruned. It has no default, so a prune never runs with a number the user
// didn't set.
func (n *CompiledFlowNode) memberPruneDays(ctx *FlowContext) (uint, error) {
	value, err := ctx.EvalTemplate(n.Data.MemberPruneDays)
	if err != nil {
		return 0, err
	}

	days, err := strconv.Atoi(strings.TrimSpace(value.String()))
	if err != nil {
		return 0, fmt.Errorf("prune days %q is not a whole number of days", value.String())
	}
	if days < 1 || days > memberPruneMaxDays {
		return 0, fmt.Errorf("prune days must be between 1 and %d, got %d", memberPruneMaxDays, days)
	}

	return uint(days), nil
}
