package flow

import (
	"errors"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
)

func init() {
	registerHandlers(map[FlowNodeType]nodeHandler{
		FlowNodeTypeActionGuildGet: executeActionGuildGet,
		FlowNodeTypeActionRoleGet:  executeActionRoleGet,
	})
}

func executeActionRoleGet(n *CompiledFlowNode, ctx *FlowContext) error {
	guildID, err := n.targetGuildID(ctx)
	if err != nil {
		return err
	}

	roleID, err := ctx.EvalTemplate(n.Data.RoleTarget)
	if err != nil {
		return err
	}

	role, err := ctx.Discord.Role(ctx, guildID, discord.RoleID(roleID.Snowflake()))
	if err != nil && !errors.Is(err, provider.ErrNotFound) {
		return err
	}

	if role != nil {
		ctx.StoreNodeResult(n, thing.NewDiscordRole(*role))
	} else {
		ctx.StoreNodeResult(n, thing.Null)
	}

	return n.ExecuteChildren(ctx)
}

func executeActionGuildGet(n *CompiledFlowNode, ctx *FlowContext) error {
	guildID, err := ctx.EvalTemplate(n.Data.GuildTarget)
	if err != nil {
		return err
	}

	guild, err := ctx.Discord.Guild(ctx, discord.GuildID(guildID.Snowflake()))
	if err != nil && !errors.Is(err, provider.ErrNotFound) {
		return err
	}

	if guild != nil {
		ctx.StoreNodeResult(n, thing.NewDiscordGuild(*guild))
	} else {
		ctx.StoreNodeResult(n, thing.Null)
	}

	return n.ExecuteChildren(ctx)
}
