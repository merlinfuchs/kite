package flow

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
)

func init() {
	registerHandlers(map[FlowNodeType]nodeHandler{
		FlowNodeTypeActionVoiceChannelJoin:  executeActionVoiceChannelJoin,
		FlowNodeTypeActionVoiceChannelLeave: executeActionVoiceChannelLeave,
		FlowNodeTypeActionStatusSet:         executeActionStatusSet,
		FlowNodeTypeActionServerLeave:       executeActionServerLeave,
		FlowNodeTypeActionServerProfileEdit: executeActionServerProfileEdit,
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

func executeActionServerProfileEdit(n *CompiledFlowNode, ctx *FlowContext) error {
	guildID, err := n.targetGuildID(ctx)
	if err != nil {
		return err
	}
	if guildID == 0 {
		return fmt.Errorf("editing the server profile only works in servers")
	}

	profile := n.Data.ServerProfileData
	if profile == nil {
		return n.ExecuteChildren(ctx)
	}

	if profile.Reset {
		err = ctx.Discord.EditServerProfile(ctx, guildID, provider.ServerProfileData{
			Nick:   option.NullString,
			Avatar: option.NullString,
			Banner: option.NullString,
			Bio:    option.NullString,
		})
		if err != nil {
			return err
		}
		return n.ExecuteChildren(ctx)
	}

	var data provider.ServerProfileData
	if data.Nick, err = evalProfileText(ctx, profile.Nick); err != nil {
		return err
	}
	if data.Bio, err = evalProfileText(ctx, profile.Bio); err != nil {
		return err
	}
	if data.Avatar, err = evalProfileImage(ctx, profile.AvatarURL); err != nil {
		return err
	}
	if data.Banner, err = evalProfileImage(ctx, profile.BannerURL); err != nil {
		return err
	}

	err = ctx.Discord.EditServerProfile(ctx, guildID, data)
	if err != nil {
		return err
	}

	return n.ExecuteChildren(ctx)
}

func evalProfileText(ctx *FlowContext, tmpl *string) (option.NullableString, error) {
	if tmpl == nil {
		return nil, nil
	}

	value, err := ctx.EvalTemplate(*tmpl)
	if err != nil {
		return nil, err
	}

	if value.String() == "" {
		return option.NullString, nil
	}
	return option.NewNullableString(value.String()), nil
}

// Discord's limit for avatars and banners.
const maxProfileImageSize = 10 * 1024 * 1024

// evalProfileImage downloads the image at the URL the template evaluates to,
// as Discord only accepts images as data URIs.
func evalProfileImage(ctx *FlowContext, tmpl *string) (option.NullableString, error) {
	if tmpl == nil {
		return nil, nil
	}

	value, err := ctx.EvalTemplate(*tmpl)
	if err != nil {
		return nil, err
	}

	rawURL := strings.TrimSpace(value.String())
	if rawURL == "" {
		return option.NullString, nil
	}

	if !isHTTPURL(rawURL) {
		return nil, fmt.Errorf("image URL must start with http:// or https://")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create image request: %w", err)
	}

	resp, err := ctx.HTTP.HTTPRequest(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download image: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxProfileImageSize+1))
	if err != nil {
		return nil, fmt.Errorf("failed to download image: %w", err)
	}
	if len(body) > maxProfileImageSize {
		return nil, fmt.Errorf("image is larger than 10 MB")
	}

	// From the content, as the server's content type can be anything.
	contentType := http.DetectContentType(body)
	switch contentType {
	case "image/png", "image/jpeg", "image/gif", "image/webp":
	default:
		return nil, fmt.Errorf("image must be a PNG, JPEG, GIF or WebP, not %s", contentType)
	}

	return option.NewNullableString(
		"data:" + contentType + ";base64," + base64.StdEncoding.EncodeToString(body),
	), nil
}

func isHTTPURL(raw string) bool {
	u, err := url.Parse(raw)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https")
}
