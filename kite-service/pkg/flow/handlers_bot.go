package flow

import (
	"fmt"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
)

func init() {
	registerHandlers(map[FlowNodeType]nodeHandler{
		FlowNodeTypeActionVoiceChannelJoin:  executeActionVoiceChannelJoin,
		FlowNodeTypeActionVoiceChannelLeave: executeActionVoiceChannelLeave,
		FlowNodeTypeActionStatusSet:         executeActionStatusSet,
		FlowNodeTypeActionBotStatsGet:       executeActionBotStatsGet,
	})
}

func executeActionVoiceChannelJoin(n *CompiledFlowNode, ctx *FlowContext) error {
	channelTarget, err := ctx.EvalTemplate(n.Data.ChannelTarget)
	if err != nil {
		return traceError(n, err)
	}

	// The guild comes from the channel, so the block also works with voice
	// channels of servers other than the one the flow runs in.
	channel, err := ctx.Discord.Channel(ctx, discord.ChannelID(channelTarget.Snowflake()))
	if err != nil {
		return traceError(n, err)
	}
	if channel.Type != discord.GuildVoice && channel.Type != discord.GuildStageVoice {
		return traceError(n, fmt.Errorf("channel %s is not a voice channel", channel.ID))
	}

	err = ctx.Discord.UpdateVoiceState(
		ctx,
		channel.GuildID,
		channel.ID,
		n.Data.VoiceSelfMute,
		n.Data.VoiceSelfDeaf,
	)
	if err != nil {
		return traceError(n, err)
	}

	return n.ExecuteChildren(ctx)
}

func executeActionVoiceChannelLeave(n *CompiledFlowNode, ctx *FlowContext) error {
	guildID, err := n.targetGuildID(ctx)
	if err != nil {
		return traceError(n, err)
	}
	if guildID == 0 {
		return traceError(n, fmt.Errorf("leaving a voice channel only works in servers"))
	}

	err = ctx.Discord.UpdateVoiceState(ctx, guildID, 0, false, false)
	if err != nil {
		return traceError(n, err)
	}

	return n.ExecuteChildren(ctx)
}

func executeActionStatusSet(n *CompiledFlowNode, ctx *FlowContext) error {
	if n.Data.StatusData == nil {
		return n.ExecuteChildren(ctx)
	}

	activityName, err := ctx.EvalTemplate(n.Data.StatusData.ActivityName)
	if err != nil {
		return traceError(n, err)
	}

	activityURL, err := ctx.EvalTemplate(n.Data.StatusData.ActivityURL)
	if err != nil {
		return traceError(n, err)
	}

	status := discord.Status(n.Data.StatusData.Status)
	if status == "" {
		status = discord.OnlineStatus
	}

	// Same shape as the statuses from the app settings, which also put the
	// name into State so it shows up for custom statuses.
	err = ctx.Discord.UpdatePresence(ctx, status, discord.Activity{
		Type:  discord.ActivityType(n.Data.StatusData.ActivityType),
		Name:  activityName.String(),
		State: activityName.String(),
		URL:   activityURL.String(),
	})
	if err != nil {
		return traceError(n, err)
	}

	return n.ExecuteChildren(ctx)
}

func executeActionBotStatsGet(n *CompiledFlowNode, ctx *FlowContext) error {
	stats, err := ctx.Discord.BotStats(ctx)
	if err != nil {
		return traceError(n, err)
	}

	// Both stay 0 while it isn't known when the bot connected.
	var uptime, connectedAt int64
	if !stats.ConnectedAt.IsZero() {
		uptime = int64(time.Since(stats.ConnectedAt).Seconds())
		connectedAt = stats.ConnectedAt.Unix()
	}

	ctx.StoreNodeResult(n, thing.NewObject(map[string]thing.Thing{
		"guild_count":  thing.NewInt(stats.GuildCount),
		"member_count": thing.NewInt(stats.MemberCount),
		"uptime":       thing.NewInt(uptime),
		"connected_at": thing.NewInt(connectedAt),
		"latency":      thing.NewInt(stats.Latency.Milliseconds()),
	}))

	return n.ExecuteChildren(ctx)
}
