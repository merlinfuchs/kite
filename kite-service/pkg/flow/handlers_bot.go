package flow

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
)

func init() {
	registerHandlers(map[FlowNodeType]nodeHandler{
		FlowNodeTypeActionVoiceChannelJoin:  executeActionVoiceChannelJoin,
		FlowNodeTypeActionVoiceChannelLeave: executeActionVoiceChannelLeave,
		FlowNodeTypeActionVoiceAudioPlay:    executeActionVoiceAudioPlay,
		FlowNodeTypeActionVoiceAudioStop:    executeActionVoiceAudioStop,
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

// maxAudioSize is the size of the largest audio file Play audio downloads.
const maxAudioSize = 25 * 1024 * 1024

// Bounds of the volume of Play audio, in percent.
const (
	minAudioVolume = 1
	maxAudioVolume = 200
)

func executeActionVoiceAudioPlay(n *CompiledFlowNode, ctx *FlowContext) error {
	channelTarget, err := ctx.EvalTemplate(n.Data.ChannelTarget)
	if err != nil {
		return err
	}

	// Like Join voice channel, the guild comes from the channel.
	channel, err := ctx.Discord.Channel(ctx, discord.ChannelID(channelTarget.Snowflake()))
	if err != nil {
		return err
	}
	switch channel.Type {
	case discord.GuildVoice:
	case discord.GuildStageVoice:
		// The bot joins stages as a listener, nobody would hear it.
		return fmt.Errorf("playing audio in stage channels isn't supported")
	default:
		return fmt.Errorf("channel %s is not a voice channel", channel.ID)
	}

	volume, err := audioVolume(n, ctx)
	if err != nil {
		return err
	}

	audioURL, err := ctx.EvalTemplate(n.Data.AudioURL)
	if err != nil {
		return err
	}
	audio, err := downloadAudio(ctx, audioURL.String())
	if err != nil {
		return err
	}

	err = ctx.Discord.PlayAudio(ctx, provider.PlayAudioOpts{
		GuildID:    channel.GuildID,
		ChannelID:  channel.ID,
		Audio:      audio,
		Volume:     volume / 100,
		LeaveAfter: n.Data.AudioLeaveAfter,
	})
	if err != nil {
		return err
	}

	return n.ExecuteChildren(ctx)
}

// audioVolume is the volume of Play audio in percent.
func audioVolume(n *CompiledFlowNode, ctx *FlowContext) (float64, error) {
	if n.Data.AudioVolume == "" {
		return 100, nil
	}

	raw, err := ctx.EvalTemplate(n.Data.AudioVolume)
	if err != nil {
		return 0, err
	}

	volume, err := strconv.ParseFloat(strings.TrimSpace(raw.String()), 64)
	if err != nil || volume < minAudioVolume || volume > maxAudioVolume {
		return 0, fmt.Errorf("the volume must be a number from %d to %d", minAudioVolume, maxAudioVolume)
	}
	return volume, nil
}

// downloadAudio downloads the audio file at the URL, through the same client
// as HTTP requests.
func downloadAudio(ctx *FlowContext, rawURL string) ([]byte, error) {
	u, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return nil, fmt.Errorf("the audio URL must be a link starting with http:// or https://")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := ctx.HTTP.HTTPRequest(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to download the audio: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("failed to download the audio: status %d", resp.StatusCode)
	}

	tooLarge := fmt.Errorf("the audio file can't be larger than %d MB", maxAudioSize/1024/1024)
	if resp.ContentLength > maxAudioSize {
		return nil, tooLarge
	}

	audio, err := io.ReadAll(io.LimitReader(resp.Body, maxAudioSize+1))
	if err != nil {
		return nil, fmt.Errorf("failed to download the audio: %w", err)
	}
	if len(audio) > maxAudioSize {
		return nil, tooLarge
	}
	if len(audio) == 0 {
		return nil, fmt.Errorf("the audio file is empty")
	}

	return audio, nil
}

func executeActionVoiceAudioStop(n *CompiledFlowNode, ctx *FlowContext) error {
	guildID, err := n.targetGuildID(ctx)
	if err != nil {
		return err
	}
	if guildID == 0 {
		return fmt.Errorf("stopping audio only works in servers")
	}

	err = ctx.Discord.StopAudio(ctx, guildID)
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
