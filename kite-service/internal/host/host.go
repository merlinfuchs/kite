package host

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/merlinfuchs/dismod/disrest"
	"github.com/merlinfuchs/dismod/distype"
	"github.com/merlinfuchs/kite/kite-sdk-go/call"
	"github.com/merlinfuchs/kite/kite-sdk-go/kv"
	"github.com/merlinfuchs/kite/kite-sdk-go/manifest"
	"github.com/merlinfuchs/kite/kite-service/internal/logging/logattr"
	"github.com/merlinfuchs/kite/kite-service/pkg/model"
	"github.com/merlinfuchs/kite/kite-service/pkg/store"
)

type HostEnvironmentStores struct {
	deployments       store.DeploymentStore
	deploymentLogs    store.DeploymentLogStore
	deploymentMetrics store.DeploymentMetricStore
	kvStorage         store.KVStorageStore
	discordState      store.DiscordStateStore
	discordClient     *disrest.Client
}

func NewHostEnvironmentStores(
	deployments store.DeploymentStore,
	deploymentLogs store.DeploymentLogStore,
	deploymentMetrics store.DeploymentMetricStore,
	kvStorage store.KVStorageStore,
	discordState store.DiscordStateStore,
	discordClient *disrest.Client,
) HostEnvironmentStores {
	return HostEnvironmentStores{
		deployments:       deployments,
		deploymentLogs:    deploymentLogs,
		deploymentMetrics: deploymentMetrics,
		kvStorage:         kvStorage,
		discordState:      discordState,
		discordClient:     discordClient,
	}
}

type HostEnvironment struct {
	HostEnvironmentStores

	DeploymentID string
	GuildID      string
	Manifest     manifest.Manifest
	Config       map[string]interface{}
}

func NewEnv(stores HostEnvironmentStores) HostEnvironment {
	return HostEnvironment{
		HostEnvironmentStores: stores,
	}
}

func (h HostEnvironment) Call(ctx context.Context, req call.Call) (res interface{}, err error) {
	timeout := time.Duration(time.Duration(req.Config.Timeout) * time.Millisecond)
	if timeout == 0 || timeout > 3*time.Second {
		timeout = 3 * time.Second
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	startTime := time.Now()

	switch req.Type {
	case call.Sleep:
		data := req.Data.(call.SleepCall)

		duration := time.Duration(data.Duration) * time.Millisecond
		if duration > 3*time.Second {
			duration = 3 * time.Second
		}

		time.Sleep(duration)
		return call.SleepResponse{}, nil
	case call.KVKeyGet:
		res, err = h.callKVKeyGet(ctx, req.Data.(kv.KVKeyGetCall))
	case call.KVKeySet:
		res, err = h.callKVKeySet(ctx, req.Data.(kv.KVKeySetCall))
	case call.KVKeyDelete:
		res, err = h.callKVKeyDelete(ctx, req.Data.(kv.KVKeyDeleteCall))
	case call.KVKeyIncrease:
		res, err = h.callKVKeyIncrease(ctx, req.Data.(kv.KVKeyIncreaseCall))
	case call.DiscordBanList:
		res, err = h.callDiscordBanList(ctx, req.Data.(distype.BanListRequest))
	case call.DiscordBanGet:
		res, err = h.callDiscordBanGet(ctx, req.Data.(distype.BanGetRequest))
	case call.DiscordBanCreate:
		res, err = h.callDiscordBanCreate(ctx, req.Data.(distype.BanCreateRequest))
	case call.DiscordBanRemove:
		res, err = h.callDiscordBanRemove(ctx, req.Data.(distype.BanRemoveRequest))
	case call.DiscordChannelGet:
		res, err = h.callDiscordChannelGet(ctx, req.Data.(distype.ChannelGetRequest))
	case call.DiscordChannelList:
		res, err = h.callDiscordChannelList(ctx, req.Data.(distype.GuildChannelListRequest))
	case call.DiscordChannelCreate:
		res, err = h.callDiscordChannelCreate(ctx, req.Data.(distype.GuildChannelCreateRequest))
	case call.DiscordChannelUpdate:
		res, err = h.callDiscordChannelUpdate(ctx, req.Data.(distype.ChannelModifyRequest))
	case call.DiscordChannelUpdatePositions:
		res, err = h.callDiscordChannelUpdatePositions(ctx, req.Data.(distype.GuildChannelModifyPositionsRequest))
	case call.DiscordChannelDelete:
		res, err = h.callDiscordChannelDelete(ctx, req.Data.(distype.ChannelDeleteRequest))
	case call.DiscordChannelUpdatePermissions:
		res, err = h.callDiscordChannelUpdatePermissions(ctx, req.Data.(distype.ChannelEditPermissionsRequest))
	case call.DiscordChannelDeletePermissions:
		res, err = h.callDiscordChannelDeletePermissions(ctx, req.Data.(distype.ChannelDeletePermissionsRequest))
	case call.DiscordThreadStartFromMessage:
		res, err = h.callDiscordThreadStartFromMessage(ctx, req.Data.(distype.ThreadStartFromMessageRequest))
	case call.DiscordThreadStart:
		res, err = h.callDiscordThreadStart(ctx, req.Data.(distype.ThreadStartWithoutMessageRequest))
	case call.DiscordThreadStartInForum:
		res, err = h.callDiscordThreadStartInForum(ctx, req.Data.(distype.ThreadStartInForumRequest))
	case call.DiscordThreadJoin:
		res, err = h.callDiscordThreadJoin(ctx, req.Data.(distype.ThreadJoinRequest))
	case call.DiscordThreadMemberAdd:
		res, err = h.callDiscordThreadMemberAdd(ctx, req.Data.(distype.ThreadMemberAddRequest))
	case call.DiscordThreadLeave:
		res, err = h.callDiscordThreadLeave(ctx, req.Data.(distype.ThreadLeaveRequest))
	case call.DiscordThreadMemberRemove:
		res, err = h.callDiscordThreadMemberRemove(ctx, req.Data.(distype.ThreadMemberRemoveRequest))
	case call.DiscordThreadMemberGet:
		res, err = h.callDiscordThreadMemberGet(ctx, req.Data.(distype.ThreadMemberGetRequest))
	case call.DiscordThreadMemberList:
		res, err = h.callDiscordThreadMemberList(ctx, req.Data.(distype.ThreadMemberListRequest))
	case call.DiscordThreadListPublicArchived:
		res, err = h.callDiscordThreadListPublicArchived(ctx, req.Data.(distype.ThreadListPublicArchivedRequest))
	case call.DiscordThreadListPrivateArchived:
		res, err = h.callDiscordThreadListPrivateArchived(ctx, req.Data.(distype.ThreadListPrivateArchivedRequest))
	case call.DiscordThreadListJoinedPrivateArchived:
		res, err = h.callDiscordThreadListJoinedPrivateArchived(ctx, req.Data.(distype.ThreadListJoinedPrivateArchivedRequest))
	case call.DiscordThreadListActive:
		res, err = h.callDiscordThreadListActive(ctx, req.Data.(distype.GuildThreadListActiveRequest))
	case call.DiscordEmojiList:
		res, err = h.callDiscordEmojiList(ctx, req.Data.(distype.GuildEmojiListRequest))
	case call.DiscordEmojiGet:
		res, err = h.callDiscordEmojiGet(ctx, req.Data.(distype.EmojiGetRequest))
	case call.DiscordEmojiCreate:
		res, err = h.callDiscordEmojiCreate(ctx, req.Data.(distype.EmojiCreateRequest))
	case call.DiscordEmojiUpdate:
		res, err = h.callDiscordEmojiUpdate(ctx, req.Data.(distype.EmojiModifyRequest))
	case call.DiscordEmojiDelete:
		res, err = h.callDiscordEmojiDelete(ctx, req.Data.(distype.EmojiDeleteRequest))
	case call.DiscordGuildGet:
		res, err = h.callDiscordGuildGet(ctx, req.Data.(distype.GuildGetRequest))
	case call.DiscordGuildUpdate:
		res, err = h.callDiscordGuildUpdate(ctx, req.Data.(distype.GuildUpdateRequest))
	case call.DiscordInteractionResponseCreate:
		res, err = h.callDiscordInteractionResponseCreate(ctx, req.Data.(distype.InteractionResponseCreateRequest))
	case call.DiscordInteractionResponseUpdate:
		res, err = h.callDiscordInteractionResponseUpdate(ctx, req.Data.(distype.InteractionResponseEditRequest))
	case call.DiscordInteractionResponseDelete:
		res, err = h.callDiscordInteractionResponseDelete(ctx, req.Data.(distype.InteractionResponseDeleteRequest))
	case call.DiscordInteractionResponseGet:
		res, err = h.callDiscordInteractionResponseGet(ctx, req.Data.(distype.InteractionResponseGetRequest))
	case call.DiscordInteractionFollowupCreate:
		res, err = h.callDiscordInteractionFollowupCreate(ctx, req.Data.(distype.InteractionFollowupCreateRequest))
	case call.DiscordInteractionFollowupUpdate:
		res, err = h.callDiscordInteractionFollowupUpdate(ctx, req.Data.(distype.InteractionFollowupEditRequest))
	case call.DiscordInteractionFollowupDelete:
		res, err = h.callDiscordInteractionFollowupDelete(ctx, req.Data.(distype.InteractionFollowupDeleteRequest))
	case call.DiscordInteractionFollowupGet:
		res, err = h.callDiscordInteractionFollowupGet(ctx, req.Data.(distype.InteractionFollowupGetRequest))
	case call.DiscordInviteListForChannel:
		res, err = h.callDiscordInviteListForChannel(ctx, req.Data.(distype.ChannelInviteListRequest))
	case call.DiscordInviteListForGuild:
		res, err = h.callDiscordInviteListForGuild(ctx, req.Data.(distype.GuildInviteListRequest))
	case call.DiscordInviteCreate:
		res, err = h.callDiscordInviteCreate(ctx, req.Data.(distype.ChannelInviteCreateRequest))
	case call.DiscordInviteGet:
		res, err = h.callDiscordInviteGet(ctx, req.Data.(distype.InviteGetRequest))
	case call.DiscordInviteDelete:
		res, err = h.callDiscordInviteDelete(ctx, req.Data.(distype.InviteDeleteRequest))
	case call.DiscordMemberGet:
		res, err = h.callDiscordMemberGet(ctx, req.Data.(distype.MemberGetRequest))
	case call.DiscordMemberList:
		res, err = h.callDiscordMemberList(ctx, req.Data.(distype.GuildMemberListRequest))
	case call.DiscordMemberSearch:
		res, err = h.callDiscordMemberSearch(ctx, req.Data.(distype.GuildMemberSearchRequest))
	case call.DiscordMemberUpdate:
		res, err = h.callDiscordMemberUpdate(ctx, req.Data.(distype.MemberModifyRequest))
	case call.DiscordMemberUpdateOwn:
		res, err = h.callDiscordMemberUpdateOwn(ctx, req.Data.(distype.MemberModifyCurrentRequest))
	case call.DiscordMemberRoleAdd:
		res, err = h.callDiscordMemberRoleAdd(ctx, req.Data.(distype.MemberRoleAddRequest))
	case call.DiscordMemberRoleRemove:
		res, err = h.callDiscordMemberRoleRemove(ctx, req.Data.(distype.MemberRoleRemoveRequest))
	case call.DiscordMemberRemove:
		res, err = h.callDiscordMemberRemove(ctx, req.Data.(distype.MemberRemoveRequest))
	case call.DiscordMemberPruneCount:
		res, err = h.callDiscordMemberPruneCount(ctx, req.Data.(distype.MemberPruneCountRequest))
	case call.DiscordMemberPruneBegin:
		res, err = h.callDiscordMemberPruneBegin(ctx, req.Data.(distype.MemberPruneRequest))
	case call.DiscordMessageList:
		res, err = h.callDiscordMessageList(ctx, req.Data.(distype.ChannelMessageListRequest))
	case call.DiscordMessageGet:
		res, err = h.callDiscordMessageGet(ctx, req.Data.(distype.MessageGetRequest))
	case call.DiscordMessageCreate:
		res, err = h.callDiscordMessageCreate(ctx, req.Data.(distype.MessageCreateRequest))
	case call.DiscordMessageUpdate:
		res, err = h.callDiscordMessageUpdate(ctx, req.Data.(distype.MessageEditRequest))
	case call.DiscordMessageDelete:
		res, err = h.callDiscordMessageDelete(ctx, req.Data.(distype.MessageDeleteRequest))
	case call.DiscordMessageDeleteBulk:
		res, err = h.callDiscordMessageDeleteBulk(ctx, req.Data.(distype.MessageBulkDeleteRequest))
	case call.DiscordMessageReactionCreate:
		res, err = h.callDiscordMessageReactionCreate(ctx, req.Data.(distype.MessageReactionCreateRequest))
	case call.DiscordMessageReactionDeleteOwn:
		res, err = h.callDiscordMessageReactionDeleteOwn(ctx, req.Data.(distype.MessageReactionDeleteOwnRequest))
	case call.DiscordMessageReactionDeleteUser:
		res, err = h.callDiscordMessageReactionDeleteUser(ctx, req.Data.(distype.MessageReactionDeleteRequest))
	case call.DiscordMessageReactionList:
		res, err = h.callDiscordMessageReactionList(ctx, req.Data.(distype.MessageReactionListRequest))
	case call.DiscordMessageReactionDeleteAll:
		res, err = h.callDiscordMessageReactionDeleteAll(ctx, req.Data.(distype.MessageReactionDeleteAllRequest))
	case call.DiscordMessageReactionDeleteEmoji:
		res, err = h.callDiscordMessageReactionDeleteEmoji(ctx, req.Data.(distype.MessageReactionDeleteEmojiRequest))
	case call.DiscordMessageGetPinned:
		res, err = h.callDiscordMessageGetPinned(ctx, req.Data.(distype.ChannelPinnedMessageListRequest))
	case call.DiscordMessagePin:
		res, err = h.callDiscordMessagePin(ctx, req.Data.(distype.MessagePinRequest))
	case call.DiscordMessageUnpin:
		res, err = h.callDiscordMessageUnpin(ctx, req.Data.(distype.MessageUnpinRequest))
	case call.DiscordRoleList:
		res, err = h.callDiscordRoleList(ctx, req.Data.(distype.GuildRoleListRequest))
	case call.DiscordRoleCreate:
		res, err = h.callDiscordRoleCreate(ctx, req.Data.(distype.RoleCreateRequest))
	case call.DiscordRoleUpdate:
		res, err = h.callDiscordRoleUpdate(ctx, req.Data.(distype.RoleModifyRequest))
	case call.DiscordRoleUpdatePositions:
		res, err = h.callDiscordRoleUpdatePositions(ctx, req.Data.(distype.RolePositionsModifyRequest))
	case call.DiscordRoleDelete:
		res, err = h.callDiscordRoleDelete(ctx, req.Data.(distype.RoleDeleteRequest))
	case call.DiscordScheduledEventList:
		res, err = h.callDiscordScheduledEventList(ctx, req.Data.(distype.GuildScheduledEventListRequest))
	case call.DiscordScheduledEventCreate:
		res, err = h.callDiscordScheduledEventCreate(ctx, req.Data.(distype.ScheduledEventCreateRequest))
	case call.DiscordScheduledEventGet:
		res, err = h.callDiscordScheduledEventGet(ctx, req.Data.(distype.ScheduledEventGetRequest))
	case call.DiscordScheduledEventUpdate:
		res, err = h.callDiscordScheduledEventUpdate(ctx, req.Data.(distype.ScheduledEventModifyRequest))
	case call.DiscordScheduledEventDelete:
		res, err = h.callDiscordScheduledEventDelete(ctx, req.Data.(distype.ScheduledEventDeleteRequest))
	case call.DiscordScheduledEventUserList:
		res, err = h.callDiscordScheduledEventUserList(ctx, req.Data.(distype.ScheduledEventUserListRequest))
	case call.DiscordStageInstanceCreate:
		res, err = h.callDiscordStageInstanceCreate(ctx, req.Data.(distype.StageInstanceCreateRequest))
	case call.DiscordStageInstanceGet:
		res, err = h.callDiscordStageInstanceGet(ctx, req.Data.(distype.StageInstanceGetRequest))
	case call.DiscordStageInstanceUpdate:
		res, err = h.callDiscordStageInstanceUpdate(ctx, req.Data.(distype.StageInstanceModifyRequest))
	case call.DiscordStageInstanceDelete:
		res, err = h.callDiscordStageInstanceDelete(ctx, req.Data.(distype.StageInstanceDeleteRequest))
	case call.DiscordStickerList:
		res, err = h.callDiscordStickerList(ctx, req.Data.(distype.GuildStickerListRequest))
	case call.DiscordStickerGet:
		res, err = h.callDiscordStickerGet(ctx, req.Data.(distype.StickerGetRequest))
	case call.DiscordStickerCreate:
		res, err = h.callDiscordStickerCreate(ctx, req.Data.(distype.StickerCreateRequest))
	case call.DiscordStickerUpdate:
		res, err = h.callDiscordStickerUpdate(ctx, req.Data.(distype.StickerModifyRequest))
	case call.DiscordStickerDelete:
		res, err = h.callDiscordStickerDelete(ctx, req.Data.(distype.StickerDeleteRequest))
	case call.DiscordUserGet:
		res, err = h.callDiscordUserGet(ctx, req.Data.(distype.UserGetRequest))
	case call.DiscordWebhookGet:
		res, err = h.callDiscordWebhookGet(ctx, req.Data.(distype.WebhookGetRequest))
	case call.DiscordWebhookListForChannel:
		res, err = h.callDiscordWebhookListForChannel(ctx, req.Data.(distype.ChannelWebhookListRequest))
	case call.DiscordWebhookListForGuild:
		res, err = h.callDiscordWebhookListForGuild(ctx, req.Data.(distype.GuildWebhookListRequest))
	case call.DiscordWebhookCreate:
		res, err = h.callDiscordWebhookCreate(ctx, req.Data.(distype.WebhookCreateRequest))
	case call.DiscordWebhookUpdate:
		res, err = h.callDiscordWebhookUpdate(ctx, req.Data.(distype.WebhookModifyRequest))
	case call.DiscordWebhookDelete:
		res, err = h.callDiscordWebhookDelete(ctx, req.Data.(distype.WebhookDeleteRequest))
	case call.DiscordWebhookGetWithToken:
		res, err = h.callDiscordWebhookGetWithToken(ctx, req.Data.(distype.WebhookGetWithTokenRequest))
	case call.DiscordWebhookUpdateWithToken:
		res, err = h.callDiscordWebhookUpdateWithToken(ctx, req.Data.(distype.WebhookModifyWithTokenRequest))
	case call.DiscordWebhookDeleteWithToken:
		res, err = h.callDiscordWebhookDeleteWithToken(ctx, req.Data.(distype.WebhookDeleteWithTokenRequest))
	case call.DiscordWebhookExecute:
		res, err = h.callDiscordWebhookExecute(ctx, req.Data.(distype.WebhookExecuteRequest))
	default:
		return nil, fmt.Errorf("unknown call type: %s", req.Type)
	}

	mCtx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	mErr := h.deploymentMetrics.CreateDeploymentMetricEntry(mCtx, model.DeploymentMetricEntry{
		DeploymentID:  h.DeploymentID,
		Type:          model.DeploymentMetricEntryTypeCallExecuted,
		CallType:      string(req.Type),
		CallSuccess:   err == nil,
		CallTotalTime: time.Since(startTime),
		Timestamp:     time.Now().UTC(),
	})
	if mErr != nil {
		slog.With(logattr.Error(mErr)).Error("Error creating deployment metric entry for action")
	}

	if err != nil {
		return nil, modelError(err)
	}

	return res, nil
}
