package host

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/disgoorg/disgo/rest"

	"github.com/merlinfuchs/dismod/distype"
	"github.com/merlinfuchs/kite/kite-sdk-go/call"
	"github.com/merlinfuchs/kite/kite-sdk-go/fail"
	"github.com/merlinfuchs/kite/kite-sdk-go/kv"
	"github.com/merlinfuchs/kite/kite-service/pkg/store"
)

func modelError(err error) *fail.HostError {
	var derr rest.Error
	if errors.As(err, &derr) {
		if derr.Code == 0 {
			return &fail.HostError{
				Code:    fail.HostErrorTypeDiscordUnknown,
				Message: derr.Error(),
			}
		}

		// TODO: map all error codes
		code := fail.HostErrorTypeDiscordUnknown
		switch derr.Code {
		case 10004:
			code = fail.HostErrorTypeDiscordGuildNotFound
		case 10003:
			code = fail.HostErrorTypeDiscordChannelNotFound
		case 10008:
			code = fail.HostErrorTypeDiscordMessageNotFound
		case 10026:
			code = fail.HostErrorTypeDiscordBanNotFound
		}

		return &fail.HostError{
			Code:    code,
			Message: derr.Message,
		}
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return &fail.HostError{
			Code:    fail.HostErrorTypeTimeout,
			Message: err.Error(),
		}
	}

	if errors.Is(err, context.Canceled) {
		return &fail.HostError{
			Code:    fail.HostErrorTypeCanceled,
			Message: err.Error(),
		}
	}

	return &fail.HostError{
		Message: err.Error(),
	}
}

func (h HostEnvironment) guildIDForCall(ctx context.Context, c call.Call) (distype.Snowflake, error) {
	switch d := c.Data.(type) {
	case call.SleepCall, kv.KVKeyGetCall, kv.KVKeySetCall, kv.KVKeyDeleteCall, kv.KVKeyIncreaseCall:
		return "", nil
	case distype.BanListRequest:
		return d.GuildID, nil
	case distype.BanGetRequest:
		return d.GuildID, nil
	case distype.BanCreateRequest:
		return d.GuildID, nil
	case distype.BanRemoveRequest:
		return d.GuildID, nil
	case distype.ChannelGetRequest:
		return h.guildIDForChannelID(ctx, d.ChannelID)
	case distype.GuildChannelListRequest:
		return d.GuildID, nil
	case distype.GuildChannelCreateRequest:
		return d.GuildID, nil
	case distype.ChannelModifyRequest:
		return h.guildIDForChannelID(ctx, d.ChannelID)
	case distype.GuildChannelModifyPositionsRequest:
		return d.GuildID, nil
	case distype.ChannelDeleteRequest:
		return h.guildIDForChannelID(ctx, d.ChannelID)
	case distype.ChannelEditPermissionsRequest:
		return h.guildIDForChannelID(ctx, d.ChannelID)
	case distype.ChannelDeletePermissionsRequest:
		return h.guildIDForChannelID(ctx, d.ChannelID)
	case distype.ThreadStartFromMessageRequest:
		return h.guildIDForChannelID(ctx, d.ChannelID)
	case distype.ThreadStartWithoutMessageRequest:
		return h.guildIDForChannelID(ctx, d.ChannelID)
	case distype.ThreadStartInForumRequest:
		return h.guildIDForChannelID(ctx, d.ChannelID)
	case distype.ThreadJoinRequest:
		return h.guildIDForChannelID(ctx, d.ChannelID)
	case distype.ThreadMemberAddRequest:
		return h.guildIDForChannelID(ctx, d.ChannelID)
	case distype.ThreadLeaveRequest:
		return h.guildIDForChannelID(ctx, d.ChannelID)
	case distype.ThreadMemberRemoveRequest:
		return h.guildIDForChannelID(ctx, d.ChannelID)
	case distype.ThreadMemberGetRequest:
		return h.guildIDForChannelID(ctx, d.ChannelID)
	case distype.ThreadMemberListRequest:
		return h.guildIDForChannelID(ctx, d.ChannelID)
	case distype.ThreadListPublicArchivedRequest:
		return h.guildIDForChannelID(ctx, d.ChannelID)
	case distype.ThreadListPrivateArchivedRequest:
		return h.guildIDForChannelID(ctx, d.ChannelID)
	case distype.ThreadListJoinedPrivateArchivedRequest:
		return h.guildIDForChannelID(ctx, d.ChannelID)
	case distype.GuildThreadListActiveRequest:
		return d.GuildID, nil
	case distype.GuildEmojiListRequest:
		return d.GuildID, nil
	case distype.EmojiGetRequest:
		return d.GuildID, nil
	case distype.EmojiCreateRequest:
		return d.GuildID, nil
	case distype.EmojiModifyRequest:
		return d.GuildID, nil
	case distype.EmojiDeleteRequest:
		return d.GuildID, nil
	case distype.GuildGetRequest:
		return d.GuildID, nil
	case distype.GuildUpdateRequest:
		return d.GuildID, nil
	case distype.InteractionResponseCreateRequest:
		return "", nil
	case distype.InteractionResponseEditRequest:
		return "", nil
	case distype.InteractionResponseDeleteRequest:
		return "", nil
	case distype.InteractionResponseGetRequest:
		return "", nil
	case distype.InteractionFollowupCreateRequest:
		return "", nil
	case distype.InteractionFollowupEditRequest:
		return "", nil
	case distype.InteractionFollowupDeleteRequest:
		return "", nil
	case distype.InteractionFollowupGetRequest:
		return "", nil
	case distype.ChannelInviteListRequest:
		return h.guildIDForChannelID(ctx, d.ChannelID)
	case distype.GuildInviteListRequest:
		return d.GuildID, nil
	case distype.ChannelInviteCreateRequest:
		return h.guildIDForChannelID(ctx, d.ChannelID)
	case distype.InviteGetRequest:
		return "", nil
	case distype.InviteDeleteRequest:
		return "", nil
	case distype.MemberGetRequest:
		return d.GuildID, nil
	case distype.GuildMemberListRequest:
		return d.GuildID, nil
	case distype.GuildMemberSearchRequest:
		return d.GuildID, nil
	case distype.MemberModifyRequest:
		return d.GuildID, nil
	case distype.MemberModifyCurrentRequest:
		return d.GuildID, nil
	case distype.MemberRoleAddRequest:
		return d.GuildID, nil
	case distype.MemberRoleRemoveRequest:
		return d.GuildID, nil
	case distype.MemberRemoveRequest:
		return d.GuildID, nil
	case distype.MemberPruneCountRequest:
		return d.GuildID, nil
	case distype.MemberPruneRequest:
		return d.GuildID, nil
	case distype.ChannelMessageListRequest:
		return h.guildIDForChannelID(ctx, d.ChannelID)
	case distype.MessageGetRequest:
		return h.guildIDForChannelID(ctx, d.ChannelID)
	case distype.MessageCreateRequest:
		return h.guildIDForChannelID(ctx, d.ChannelID)
	case distype.MessageEditRequest:
		return h.guildIDForChannelID(ctx, d.ChannelID)
	case distype.MessageDeleteRequest:
		return h.guildIDForChannelID(ctx, d.ChannelID)
	case distype.MessageBulkDeleteRequest:
		return h.guildIDForChannelID(ctx, d.ChannelID)
	case distype.MessageReactionCreateRequest:
		return h.guildIDForChannelID(ctx, d.ChannelID)
	case distype.MessageReactionDeleteOwnRequest:
		return h.guildIDForChannelID(ctx, d.ChannelID)
	case distype.MessageReactionDeleteRequest:
		return h.guildIDForChannelID(ctx, d.ChannelID)
	case distype.MessageReactionListRequest:
		return h.guildIDForChannelID(ctx, d.ChannelID)
	case distype.MessageReactionDeleteAllRequest:
		return h.guildIDForChannelID(ctx, d.ChannelID)
	case distype.MessageReactionDeleteEmojiRequest:
		return h.guildIDForChannelID(ctx, d.ChannelID)
	case distype.ChannelPinnedMessageListRequest:
		return h.guildIDForChannelID(ctx, d.ChannelID)
	case distype.MessagePinRequest:
		return h.guildIDForChannelID(ctx, d.ChannelID)
	case distype.MessageUnpinRequest:
		return h.guildIDForChannelID(ctx, d.ChannelID)
	case distype.GuildRoleListRequest:
		return d.GuildID, nil
	case distype.RoleCreateRequest:
		return d.GuildID, nil
	case distype.RoleModifyRequest:
		return d.GuildID, nil
	case distype.RolePositionsModifyRequest:
		return d.GuildID, nil
	case distype.RoleDeleteRequest:
		return d.GuildID, nil
	case distype.GuildScheduledEventListRequest:
		return d.GuildID, nil
	case distype.ScheduledEventCreateRequest:
		return d.GuildID, nil
	case distype.ScheduledEventGetRequest:
		return d.GuildID, nil
	case distype.ScheduledEventModifyRequest:
		return d.GuildID, nil
	case distype.ScheduledEventDeleteRequest:
		return d.GuildID, nil
	case distype.ScheduledEventUserListRequest:
		return d.GuildID, nil
	case distype.StageInstanceCreateRequest:
		return h.guildIDForChannelID(ctx, d.ChannelID)
	case distype.StageInstanceGetRequest:
		return h.guildIDForChannelID(ctx, d.ChannelID)
	case distype.StageInstanceModifyRequest:
		return h.guildIDForChannelID(ctx, d.ChannelID)
	case distype.StageInstanceDeleteRequest:
		return h.guildIDForChannelID(ctx, d.ChannelID)
	case distype.GuildStickerListRequest:
		return d.GuildID, nil
	case distype.StickerGetRequest:
		return d.GuildID, nil
	case distype.StickerCreateRequest:
		return d.GuildID, nil
	case distype.StickerModifyRequest:
		return d.GuildID, nil
	case distype.StickerDeleteRequest:
		return d.GuildID, nil
	case distype.UserGetRequest:
		return "", nil
	case distype.WebhookGetRequest:
		return "", fmt.Errorf("cannot get guild id from WebhookGetRequest")
	case distype.ChannelWebhookListRequest:
		return h.guildIDForChannelID(ctx, d.ChannelID)
	case distype.GuildWebhookListRequest:
		return d.GuildID, nil
	case distype.WebhookCreateRequest:
		return h.guildIDForChannelID(ctx, d.ChannelID)
	case distype.WebhookModifyRequest:
		return "", fmt.Errorf("cannot get guild id from WebhookModifyRequest")
	case distype.WebhookDeleteRequest:
		return "", fmt.Errorf("cannot get guild id from WebhookDeleteRequest")
	case distype.WebhookGetWithTokenRequest:
		return "", nil
	case distype.WebhookModifyWithTokenRequest:
		return "", nil
	case distype.WebhookDeleteWithTokenRequest:
		return "", nil
	case distype.WebhookExecuteRequest:
		return "", nil
	}

	slog.With("type", fmt.Sprintf("%T", c.Data)).Error("call type not handled in guildIDForCall")
	return "", fmt.Errorf("cannot get guild id from call type: %T", c.Data)
}

func (h HostEnvironment) guildIDForChannelID(ctx context.Context, channelID distype.Snowflake) (distype.Snowflake, error) {
	channel, err := h.discordState.GetChannel(ctx, channelID)
	if err != nil {
		if err == store.ErrNotFound {
			return "", fmt.Errorf("channel %s not found", channelID)
		}
		return "", fmt.Errorf("error getting channel to get guild id: %w", err)
	}

	if channel.GuildID == nil {
		return "", fmt.Errorf("channel %s is not part of a guild", channelID)
	}

	return *channel.GuildID, nil
}
