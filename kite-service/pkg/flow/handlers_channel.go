package flow

import (
	"errors"
	"fmt"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
)

func init() {
	registerHandlers(map[FlowNodeType]nodeHandler{
		FlowNodeTypeActionChannelGet:      executeActionChannelGet,
		FlowNodeTypeActionChannelCreate:   executeActionChannelCreate,
		FlowNodeTypeActionChannelEdit:     executeActionChannelEdit,
		FlowNodeTypeActionThreadCreate:    executeActionThreadCreate,
		FlowNodeTypeActionForumPostCreate: executeActionForumPostCreate,
	})
}

func executeActionChannelGet(n *CompiledFlowNode, ctx *FlowContext) error {
	channelID, err := ctx.EvalTemplate(n.Data.ChannelTarget)
	if err != nil {
		return traceError(n, err)
	}

	channel, err := ctx.Discord.Channel(ctx, discord.ChannelID(channelID.Snowflake()))
	if err != nil && !errors.Is(err, provider.ErrNotFound) {
		return traceError(n, err)
	}

	if channel != nil {
		ctx.StoreNodeResult(n, thing.NewDiscordChannel(*channel))
	} else {
		ctx.StoreNodeResult(n, thing.Null)
	}

	return n.ExecuteChildren(ctx)
}

func executeActionChannelCreate(n *CompiledFlowNode, ctx *FlowContext) error {
	if n.Data.ChannelData == nil {
		return traceError(n, fmt.Errorf("channel data is required"))
	}

	guildID, err := n.targetGuildID(ctx)
	if err != nil {
		return traceError(n, err)
	}

	channelData, err := n.Data.ChannelData.ToCreateChannelData(ctx, ctx.EvalCtx)
	if err != nil {
		return traceError(n, err)
	}

	channel, err := ctx.Discord.CreateChannel(ctx, guildID, channelData)
	if err != nil {
		return traceError(n, err)
	}

	ctx.StoreNodeResult(n, thing.NewDiscordChannel(*channel))
	return n.ExecuteChildren(ctx)
}

func executeActionChannelEdit(n *CompiledFlowNode, ctx *FlowContext) error {
	if n.Data.ChannelData == nil {
		return traceError(n, fmt.Errorf("channel data is required"))
	}

	channelTarget, err := ctx.EvalTemplate(n.Data.ChannelTarget)
	if err != nil {
		return traceError(n, err)
	}

	createData, err := n.Data.ChannelData.ToCreateChannelData(ctx, ctx.EvalCtx)
	if err != nil {
		return traceError(n, err)
	}

	editData := api.ModifyChannelData{
		Name: createData.Name,
		NSFW: &option.NullableBoolData{
			Val:  createData.NSFW,
			Init: true,
		},
	}
	if createData.Topic != "" {
		editData.Topic = option.NewNullableString(createData.Topic)
	}
	if createData.CategoryID != discord.NullChannelID && createData.CategoryID != 0 {
		parentTarget, err := ctx.EvalTemplate(n.Data.ChannelData.ParentID)
		if err != nil {
			return traceError(n, err)
		}

		editData.CategoryID = discord.ChannelID(parentTarget.Snowflake())
	}
	if createData.VoiceBitrate != 0 {
		editData.VoiceBitrate = option.NewNullableUint(createData.VoiceBitrate)
	}
	if createData.VoiceUserLimit != 0 {
		editData.VoiceUserLimit = option.NewNullableUint(createData.VoiceUserLimit)
	}
	if createData.Position != nil {
		editData.Position = option.NewNullableInt(*createData.Position)
	}
	if len(createData.Overwrites) > 0 {
		editData.Overwrites = &createData.Overwrites
	}

	// Unlike the other fields, 0 is meaningful here (it turns slowmode off),
	// so only an empty setting leaves the current slowmode alone.
	slowmode, ok, err := n.Data.ChannelData.EvalSlowmode(ctx, ctx.EvalCtx)
	if err != nil {
		return traceError(n, err)
	}
	if ok {
		editData.UserRateLimit = option.NewNullableUint(uint(slowmode))
	}

	err = ctx.Discord.EditChannel(ctx, discord.ChannelID(channelTarget.Snowflake()), editData)
	if err != nil {
		return traceError(n, err)
	}

	// TODO: ctx.StoreNodeResult(n, thing.NewDiscordChannel(*channel))
	return n.ExecuteChildren(ctx)
}

func executeActionThreadCreate(n *CompiledFlowNode, ctx *FlowContext) error {
	if n.Data.ChannelData == nil {
		return traceError(n, fmt.Errorf("channel data is required"))
	}

	parentTarget, err := ctx.EvalTemplate(n.Data.ChannelData.ParentID)
	if err != nil {
		return traceError(n, err)
	}

	messageTarget, err := ctx.EvalTemplate(n.Data.MessageTarget)
	if err != nil {
		return traceError(n, err)
	}

	channelName, err := ctx.EvalTemplate(n.Data.ChannelData.Name)
	if err != nil {
		return traceError(n, err)
	}

	threadData := api.StartThreadData{
		Name:      channelName.String(),
		Type:      discord.ChannelType(n.Data.ChannelData.Type),
		Invitable: n.Data.ChannelData.Invitable,
	}
	if threadData.Type == 0 {
		threadData.Type = discord.GuildPublicThread
	}

	var thread *discord.Channel
	if messageTarget.IsEmpty() || messageTarget.IsNil() {
		thread, err = ctx.Discord.StartThreadWithoutMessage(
			ctx,
			discord.ChannelID(parentTarget.Snowflake()),
			threadData,
		)
		if err != nil {
			return traceError(n, err)
		}
	} else {
		thread, err = ctx.Discord.StartThreadWithMessage(
			ctx,
			discord.ChannelID(parentTarget.Snowflake()),
			discord.MessageID(messageTarget.Snowflake()),
			threadData,
		)
		if err != nil {
			return traceError(n, err)
		}
	}

	ctx.StoreNodeResult(n, thing.NewDiscordChannel(*thread))
	return n.ExecuteChildren(ctx)
}

func executeActionForumPostCreate(n *CompiledFlowNode, ctx *FlowContext) error {
	return n.ExecuteChildren(ctx)
}
