package flow

import (
	"fmt"

	"github.com/diamondburned/arikawa/v3/discord"
)

func init() {
	registerHandlers(map[FlowNodeType]nodeHandler{
		FlowNodeTypeActionVoiceChannelJoin:  executeActionVoiceChannelJoin,
		FlowNodeTypeActionVoiceChannelLeave: executeActionVoiceChannelLeave,
		FlowNodeTypeActionStatusSet:         executeActionStatusSet,
		FlowNodeTypeActionServerLeave:       executeActionServerLeave,
	})
}

func executeActionVoiceChannelJoin(n *CompiledFlowNode, ctx *FlowContext) error {
	channelTarget, err := ctx.EvalTemplate(n.Data.ChannelTarget)
	if err != nil {
		return err
	}

	// The guild comes from the channel, so the block also works with voice
	// channels of servers other than the one the flow runs in.
	channel, err := ctx.Discord.Channel(ctx, discord.ChannelID(channelTarget.Snowflake()))
	if err != nil {
		return err
	}
	if channel.Type != discord.GuildVoice && channel.Type != discord.GuildStageVoice {
		return fmt.Errorf("channel %s is not a voice channel", channel.ID)
	}

	err = ctx.Discord.UpdateVoiceState(
		ctx,
		channel.GuildID,
		channel.ID,
		n.Data.VoiceSelfMute,
		n.Data.VoiceSelfDeaf,
	)
	if err != nil {
		return err
	}

	return n.ExecuteChildren(ctx)
}

func executeActionVoiceChannelLeave(n *CompiledFlowNode, ctx *FlowContext) error {
	guildID, err := n.targetGuildID(ctx)
	if err != nil {
		return err
	}
	if guildID == 0 {
		return fmt.Errorf("leaving a voice channel only works in servers")
	}

	err = ctx.Discord.UpdateVoiceState(ctx, guildID, 0, false, false)
	if err != nil {
		return err
	}

	return n.ExecuteChildren(ctx)
}

func executeActionStatusSet(n *CompiledFlowNode, ctx *FlowContext) error {
	if n.Data.StatusData == nil {
		return n.ExecuteChildren(ctx)
	}

	activityName, err := ctx.EvalTemplate(n.Data.StatusData.ActivityName)
	if err != nil {
		return err
	}

	activityState, err := ctx.EvalTemplate(n.Data.StatusData.ActivityState)
	if err != nil {
		return err
	}

	activityURL, err := ctx.EvalTemplate(n.Data.StatusData.ActivityURL)
	if err != nil {
		return err
	}

	status := discord.Status(n.Data.StatusData.Status)
	if status == "" {
		status = discord.OnlineStatus
	}

	activityType := discord.ActivityType(n.Data.StatusData.ActivityType)
	if activityType == discord.StreamingActivity && !IsStreamURL(activityURL.String()) {
		return fmt.Errorf("the streaming activity needs a Twitch or YouTube URL")
	}

	// Same shape as the statuses from the app settings
	err = ctx.Discord.UpdatePresence(ctx, status, StatusActivity(
		activityType,
		activityName.String(),
		activityState.String(),
		activityURL.String(),
	))
	if err != nil {
		return err
	}

	return n.ExecuteChildren(ctx)
}

func executeActionServerLeave(n *CompiledFlowNode, ctx *FlowContext) error {
	// Always the server the flow runs in, never a configurable one, so the
	// block can't be aimed at another server.
	guildID := ctx.Data.GuildID()
	if guildID == 0 {
		return fmt.Errorf("leaving a server only works in flows that run in a server")
	}

	err := ctx.Discord.LeaveGuild(ctx, guildID)
	if err != nil {
		return err
	}

	return n.ExecuteChildren(ctx)
}
