package flow

import (
	"errors"
	"fmt"
	"slices"
	"strings"

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
	// so only an empty setting leaves the current slowmode alone. The type
	// setting isn't the edited channel's type, so it's left to Discord to
	// reject channels without slowmode.
	seconds, ok, err := n.Data.ChannelData.EvalRateLimitPerUser(ctx, ctx.EvalCtx)
	if err != nil {
		return traceError(n, err)
	}
	if ok {
		editData.UserRateLimit = option.NewNullableUint(uint(seconds))
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

// maxForumPostTags is how many tags Discord allows on a forum post.
const maxForumPostTags = 5

func executeActionForumPostCreate(n *CompiledFlowNode, ctx *FlowContext) error {
	if ctx.IsEntry() {
		return n.resumeFromComponent(ctx)
	}

	if n.Data.ChannelData == nil {
		return traceError(n, fmt.Errorf("channel data is required"))
	}

	channelTarget, err := ctx.EvalTemplate(n.Data.ChannelTarget)
	if err != nil {
		return traceError(n, err)
	}
	channelID := discord.ChannelID(channelTarget.Snowflake())
	if !channelID.IsValid() {
		return traceError(n, fmt.Errorf("invalid forum channel: %q", channelTarget.String()))
	}

	name, err := ctx.EvalTemplate(n.Data.ChannelData.Name)
	if err != nil {
		return traceError(n, err)
	}
	if strings.TrimSpace(name.String()) == "" {
		return traceError(n, fmt.Errorf("post name is required"))
	}

	tags, err := n.forumPostTags(ctx, channelID)
	if err != nil {
		return traceError(n, err)
	}

	auditLogReason, err := ctx.EvalTemplate(n.Data.AuditLogReason)
	if err != nil {
		return traceError(n, err)
	}

	data, opts, resumePointID, err := n.prepareMessage(ctx)
	if err != nil {
		return traceError(n, err)
	}

	thread, err := ctx.Discord.StartThreadInForum(ctx, channelID, api.StartForumThreadData{
		Name:           name.String(),
		AppliedTags:    tags,
		Message:        data.ToSendMessageData(opts),
		AuditLogReason: api.AuditLogReason(auditLogReason.String()),
	})
	if err != nil {
		return traceError(n, err)
	}

	ctx.StoreNodeResult(n, thing.NewDiscordChannel(*thread))
	if resumePointID != "" {
		// We have to create the resume point after the post has been stored
		_, err = ctx.suspend(ResumePointTypeMessageComponents, resumePointID, n.ID)
		if err != nil {
			return traceError(n, err)
		}
	}

	if n.Data.MessageTemplateID != "" {
		// The first message of a post has the ID of the post.
		err := ctx.MessageTemplate.LinkMessageTemplateInstance(ctx, provider.MessageTemplateInstance{
			MessageTemplateID: n.Data.MessageTemplateID,
			MessageID:         discord.MessageID(thread.ID),
			ChannelID:         thread.ID,
			GuildID:           ctx.Data.GuildID(),
		})
		if err != nil {
			ctx.Log.CreateLogEntry(ctx, n.Data.LogLevel, fmt.Sprintf("failed to link message template instance: %s", err.Error()))
		}
	}

	return n.ExecuteChildren(ctx)
}

// forumPostTags returns the IDs of the forum's tags the node names, by name
// or ID, since Discord doesn't show tag IDs.
func (n *CompiledFlowNode) forumPostTags(ctx *FlowContext, channelID discord.ChannelID) ([]discord.TagID, error) {
	value, err := ctx.EvalTemplate(n.Data.ChannelData.AppliedTags)
	if err != nil {
		return nil, err
	}

	items, ok := apiListItems(value)
	if !ok {
		for _, s := range strings.Split(value.String(), ",") {
			items = append(items, thing.NewString(s))
		}
	}

	var names []string
	for _, item := range items {
		if name := strings.TrimSpace(item.String()); name != "" {
			names = append(names, name)
		}
	}
	if len(names) == 0 {
		return nil, nil
	}
	if len(names) > maxForumPostTags {
		return nil, fmt.Errorf("a post can have at most %d tags, got %d", maxForumPostTags, len(names))
	}

	channel, err := ctx.Discord.Channel(ctx, channelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get forum channel: %w", err)
	}
	if channel == nil || channel.Type != discord.GuildForum && channel.Type != guildMediaChannel {
		return nil, fmt.Errorf("channel %s is not a forum or media channel", channelID)
	}

	tags := make([]discord.TagID, 0, len(names))
	for _, name := range names {
		i := slices.IndexFunc(channel.AvailableTags, func(tag discord.Tag) bool {
			return tag.ID.String() == name || strings.EqualFold(tag.Name, name)
		})
		if i == -1 {
			return nil, fmt.Errorf("forum has no tag %q", name)
		}
		if !slices.Contains(tags, channel.AvailableTags[i].ID) {
			tags = append(tags, channel.AvailableTags[i].ID)
		}
	}
	return tags, nil
}
