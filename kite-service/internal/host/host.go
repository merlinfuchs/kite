package host

import (
	"context"
	"fmt"
	"log/slog"
	"time"

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
	apps              store.AppProvider
}

func NewHostEnvironmentStores(
	deployments store.DeploymentStore,
	deploymentLogs store.DeploymentLogStore,
	deploymentMetrics store.DeploymentMetricStore,
	kvStorage store.KVStorageStore,
	apps store.AppProvider,
) HostEnvironmentStores {
	return HostEnvironmentStores{
		deployments:       deployments,
		deploymentLogs:    deploymentLogs,
		deploymentMetrics: deploymentMetrics,
		kvStorage:         kvStorage,
		apps:              apps,
	}
}

type HostEnvironment struct {
	HostEnvironmentStores

	DeploymentID string
	AppID        string
	Manifest     manifest.Manifest
	Config       map[string]interface{}
	State        store.AppStateProvider
}

func NewEnv(stores HostEnvironmentStores, appID string, deploymentID string) HostEnvironment {
	return HostEnvironment{
		HostEnvironmentStores: stores,
		DeploymentID:          deploymentID,
		AppID:                 appID,
	}
}

func (h HostEnvironment) AppState() (store.AppStateProvider, error) {
	app, err := h.apps.AppState(distype.Snowflake(h.AppID))
	if err != nil {
		if err == store.ErrNotFound {
			return nil, fmt.Errorf("app state %s not found: %w", h.AppID, err)
		}
		return nil, fmt.Errorf("error getting app state: %w", err)
	}

	return app, nil
}

func (h HostEnvironment) Call(ctx context.Context, req call.Call) (res interface{}, err error) {
	timeout := time.Duration(time.Duration(req.Config.Timeout) * time.Millisecond)
	if timeout == 0 || timeout > 3*time.Second {
		timeout = 3 * time.Second
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	startTime := time.Now()
	defer func() {
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
	}()

	app, err := h.AppState()
	if err != nil {
		return nil, fmt.Errorf("error getting app state: %w", err)
	}

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
		res, err = h.callDiscordBanList(ctx, app, req.Data.(distype.BanListRequest))
	case call.DiscordBanGet:
		res, err = h.callDiscordBanGet(ctx, app, req.Data.(distype.BanGetRequest))
	case call.DiscordBanCreate:
		res, err = h.callDiscordBanCreate(ctx, app, req.Data.(distype.BanCreateRequest))
	case call.DiscordBanRemove:
		res, err = h.callDiscordBanRemove(ctx, app, req.Data.(distype.BanRemoveRequest))
	case call.DiscordChannelGet:
		res, err = h.callDiscordChannelGet(ctx, app, req.Data.(distype.ChannelGetRequest))
	case call.DiscordChannelList:
		res, err = h.callDiscordChannelList(ctx, app, req.Data.(distype.GuildChannelListRequest))
	case call.DiscordChannelCreate:
		res, err = h.callDiscordChannelCreate(ctx, app, req.Data.(distype.GuildChannelCreateRequest))
	case call.DiscordChannelUpdate:
		res, err = h.callDiscordChannelUpdate(ctx, app, req.Data.(distype.ChannelModifyRequest))
	case call.DiscordChannelUpdatePositions:
		res, err = h.callDiscordChannelUpdatePositions(ctx, app, req.Data.(distype.GuildChannelModifyPositionsRequest))
	case call.DiscordChannelDelete:
		res, err = h.callDiscordChannelDelete(ctx, app, req.Data.(distype.ChannelDeleteRequest))
	case call.DiscordChannelUpdatePermissions:
		res, err = h.callDiscordChannelUpdatePermissions(ctx, app, req.Data.(distype.ChannelEditPermissionsRequest))
	case call.DiscordChannelDeletePermissions:
		res, err = h.callDiscordChannelDeletePermissions(ctx, app, req.Data.(distype.ChannelDeletePermissionsRequest))
	case call.DiscordThreadStartFromMessage:
		res, err = h.callDiscordThreadStartFromMessage(ctx, app, req.Data.(distype.ThreadStartFromMessageRequest))
	case call.DiscordThreadStart:
		res, err = h.callDiscordThreadStart(ctx, app, req.Data.(distype.ThreadStartWithoutMessageRequest))
	case call.DiscordThreadStartInForum:
		res, err = h.callDiscordThreadStartInForum(ctx, app, req.Data.(distype.ThreadStartInForumRequest))
	case call.DiscordThreadJoin:
		res, err = h.callDiscordThreadJoin(ctx, app, req.Data.(distype.ThreadJoinRequest))
	case call.DiscordThreadMemberAdd:
		res, err = h.callDiscordThreadMemberAdd(ctx, app, req.Data.(distype.ThreadMemberAddRequest))
	case call.DiscordThreadLeave:
		res, err = h.callDiscordThreadLeave(ctx, app, req.Data.(distype.ThreadLeaveRequest))
	case call.DiscordThreadMemberRemove:
		res, err = h.callDiscordThreadMemberRemove(ctx, app, req.Data.(distype.ThreadMemberRemoveRequest))
	case call.DiscordThreadMemberGet:
		res, err = h.callDiscordThreadMemberGet(ctx, app, req.Data.(distype.ThreadMemberGetRequest))
	case call.DiscordThreadMemberList:
		res, err = h.callDiscordThreadMemberList(ctx, app, req.Data.(distype.ThreadMemberListRequest))
	case call.DiscordThreadListPublicArchived:
		res, err = h.callDiscordThreadListPublicArchived(ctx, app, req.Data.(distype.ThreadListPublicArchivedRequest))
	case call.DiscordThreadListPrivateArchived:
		res, err = h.callDiscordThreadListPrivateArchived(ctx, app, req.Data.(distype.ThreadListPrivateArchivedRequest))
	case call.DiscordThreadListJoinedPrivateArchived:
		res, err = h.callDiscordThreadListJoinedPrivateArchived(ctx, app, req.Data.(distype.ThreadListJoinedPrivateArchivedRequest))
	case call.DiscordThreadListActive:
		res, err = h.callDiscordThreadListActive(ctx, app, req.Data.(distype.GuildThreadListActiveRequest))
	case call.DiscordEmojiList:
		res, err = h.callDiscordEmojiList(ctx, app, req.Data.(distype.GuildEmojiListRequest))
	case call.DiscordEmojiGet:
		res, err = h.callDiscordEmojiGet(ctx, app, req.Data.(distype.EmojiGetRequest))
	case call.DiscordEmojiCreate:
		res, err = h.callDiscordEmojiCreate(ctx, app, req.Data.(distype.EmojiCreateRequest))
	case call.DiscordEmojiUpdate:
		res, err = h.callDiscordEmojiUpdate(ctx, app, req.Data.(distype.EmojiModifyRequest))
	case call.DiscordEmojiDelete:
		res, err = h.callDiscordEmojiDelete(ctx, app, req.Data.(distype.EmojiDeleteRequest))
	case call.DiscordGuildGet:
		res, err = h.callDiscordGuildGet(ctx, app, req.Data.(distype.GuildGetRequest))
	case call.DiscordGuildUpdate:
		res, err = h.callDiscordGuildUpdate(ctx, app, req.Data.(distype.GuildUpdateRequest))
	case call.DiscordInteractionResponseCreate:
		res, err = h.callDiscordInteractionResponseCreate(ctx, app, req.Data.(distype.InteractionResponseCreateRequest))
	case call.DiscordInteractionResponseUpdate:
		res, err = h.callDiscordInteractionResponseUpdate(ctx, app, req.Data.(distype.InteractionResponseEditRequest))
	case call.DiscordInteractionResponseDelete:
		res, err = h.callDiscordInteractionResponseDelete(ctx, app, req.Data.(distype.InteractionResponseDeleteRequest))
	case call.DiscordInteractionResponseGet:
		res, err = h.callDiscordInteractionResponseGet(ctx, app, req.Data.(distype.InteractionResponseGetRequest))
	case call.DiscordInteractionFollowupCreate:
		res, err = h.callDiscordInteractionFollowupCreate(ctx, app, req.Data.(distype.InteractionFollowupCreateRequest))
	case call.DiscordInteractionFollowupUpdate:
		res, err = h.callDiscordInteractionFollowupUpdate(ctx, app, req.Data.(distype.InteractionFollowupEditRequest))
	case call.DiscordInteractionFollowupDelete:
		res, err = h.callDiscordInteractionFollowupDelete(ctx, app, req.Data.(distype.InteractionFollowupDeleteRequest))
	case call.DiscordInteractionFollowupGet:
		res, err = h.callDiscordInteractionFollowupGet(ctx, app, req.Data.(distype.InteractionFollowupGetRequest))
	case call.DiscordInviteListForChannel:
		res, err = h.callDiscordInviteListForChannel(ctx, app, req.Data.(distype.ChannelInviteListRequest))
	case call.DiscordInviteListForGuild:
		res, err = h.callDiscordInviteListForGuild(ctx, app, req.Data.(distype.GuildInviteListRequest))
	case call.DiscordInviteCreate:
		res, err = h.callDiscordInviteCreate(ctx, app, req.Data.(distype.ChannelInviteCreateRequest))
	case call.DiscordInviteGet:
		res, err = h.callDiscordInviteGet(ctx, app, req.Data.(distype.InviteGetRequest))
	case call.DiscordInviteDelete:
		res, err = h.callDiscordInviteDelete(ctx, app, req.Data.(distype.InviteDeleteRequest))
	case call.DiscordMemberGet:
		res, err = h.callDiscordMemberGet(ctx, app, req.Data.(distype.MemberGetRequest))
	case call.DiscordMemberList:
		res, err = h.callDiscordMemberList(ctx, app, req.Data.(distype.GuildMemberListRequest))
	case call.DiscordMemberSearch:
		res, err = h.callDiscordMemberSearch(ctx, app, req.Data.(distype.GuildMemberSearchRequest))
	case call.DiscordMemberUpdate:
		res, err = h.callDiscordMemberUpdate(ctx, app, req.Data.(distype.MemberModifyRequest))
	case call.DiscordMemberUpdateOwn:
		res, err = h.callDiscordMemberUpdateOwn(ctx, app, req.Data.(distype.MemberModifyCurrentRequest))
	case call.DiscordMemberRoleAdd:
		res, err = h.callDiscordMemberRoleAdd(ctx, app, req.Data.(distype.MemberRoleAddRequest))
	case call.DiscordMemberRoleRemove:
		res, err = h.callDiscordMemberRoleRemove(ctx, app, req.Data.(distype.MemberRoleRemoveRequest))
	case call.DiscordMemberRemove:
		res, err = h.callDiscordMemberRemove(ctx, app, req.Data.(distype.MemberRemoveRequest))
	case call.DiscordMemberPruneCount:
		res, err = h.callDiscordMemberPruneCount(ctx, app, req.Data.(distype.MemberPruneCountRequest))
	case call.DiscordMemberPruneBegin:
		res, err = h.callDiscordMemberPruneBegin(ctx, app, req.Data.(distype.MemberPruneRequest))
	case call.DiscordMessageList:
		res, err = h.callDiscordMessageList(ctx, app, req.Data.(distype.ChannelMessageListRequest))
	case call.DiscordMessageGet:
		res, err = h.callDiscordMessageGet(ctx, app, req.Data.(distype.MessageGetRequest))
	case call.DiscordMessageCreate:
		res, err = h.callDiscordMessageCreate(ctx, app, req.Data.(distype.MessageCreateRequest))
	case call.DiscordMessageUpdate:
		res, err = h.callDiscordMessageUpdate(ctx, app, req.Data.(distype.MessageEditRequest))
	case call.DiscordMessageDelete:
		res, err = h.callDiscordMessageDelete(ctx, app, req.Data.(distype.MessageDeleteRequest))
	case call.DiscordMessageDeleteBulk:
		res, err = h.callDiscordMessageDeleteBulk(ctx, app, req.Data.(distype.MessageBulkDeleteRequest))
	case call.DiscordMessageReactionCreate:
		res, err = h.callDiscordMessageReactionCreate(ctx, app, req.Data.(distype.MessageReactionCreateRequest))
	case call.DiscordMessageReactionDeleteOwn:
		res, err = h.callDiscordMessageReactionDeleteOwn(ctx, app, req.Data.(distype.MessageReactionDeleteOwnRequest))
	case call.DiscordMessageReactionDeleteUser:
		res, err = h.callDiscordMessageReactionDeleteUser(ctx, app, req.Data.(distype.MessageReactionDeleteRequest))
	case call.DiscordMessageReactionList:
		res, err = h.callDiscordMessageReactionList(ctx, app, req.Data.(distype.MessageReactionListRequest))
	case call.DiscordMessageReactionDeleteAll:
		res, err = h.callDiscordMessageReactionDeleteAll(ctx, app, req.Data.(distype.MessageReactionDeleteAllRequest))
	case call.DiscordMessageReactionDeleteEmoji:
		res, err = h.callDiscordMessageReactionDeleteEmoji(ctx, app, req.Data.(distype.MessageReactionDeleteEmojiRequest))
	case call.DiscordMessageGetPinned:
		res, err = h.callDiscordMessageGetPinned(ctx, app, req.Data.(distype.ChannelPinnedMessageListRequest))
	case call.DiscordMessagePin:
		res, err = h.callDiscordMessagePin(ctx, app, req.Data.(distype.MessagePinRequest))
	case call.DiscordMessageUnpin:
		res, err = h.callDiscordMessageUnpin(ctx, app, req.Data.(distype.MessageUnpinRequest))
	case call.DiscordRoleList:
		res, err = h.callDiscordRoleList(ctx, app, req.Data.(distype.GuildRoleListRequest))
	case call.DiscordRoleCreate:
		res, err = h.callDiscordRoleCreate(ctx, app, req.Data.(distype.RoleCreateRequest))
	case call.DiscordRoleUpdate:
		res, err = h.callDiscordRoleUpdate(ctx, app, req.Data.(distype.RoleModifyRequest))
	case call.DiscordRoleUpdatePositions:
		res, err = h.callDiscordRoleUpdatePositions(ctx, app, req.Data.(distype.RolePositionsModifyRequest))
	case call.DiscordRoleDelete:
		res, err = h.callDiscordRoleDelete(ctx, app, req.Data.(distype.RoleDeleteRequest))
	case call.DiscordScheduledEventList:
		res, err = h.callDiscordScheduledEventList(ctx, app, req.Data.(distype.GuildScheduledEventListRequest))
	case call.DiscordScheduledEventCreate:
		res, err = h.callDiscordScheduledEventCreate(ctx, app, req.Data.(distype.ScheduledEventCreateRequest))
	case call.DiscordScheduledEventGet:
		res, err = h.callDiscordScheduledEventGet(ctx, app, req.Data.(distype.ScheduledEventGetRequest))
	case call.DiscordScheduledEventUpdate:
		res, err = h.callDiscordScheduledEventUpdate(ctx, app, req.Data.(distype.ScheduledEventModifyRequest))
	case call.DiscordScheduledEventDelete:
		res, err = h.callDiscordScheduledEventDelete(ctx, app, req.Data.(distype.ScheduledEventDeleteRequest))
	case call.DiscordScheduledEventUserList:
		res, err = h.callDiscordScheduledEventUserList(ctx, app, req.Data.(distype.ScheduledEventUserListRequest))
	case call.DiscordStageInstanceCreate:
		res, err = h.callDiscordStageInstanceCreate(ctx, app, req.Data.(distype.StageInstanceCreateRequest))
	case call.DiscordStageInstanceGet:
		res, err = h.callDiscordStageInstanceGet(ctx, app, req.Data.(distype.StageInstanceGetRequest))
	case call.DiscordStageInstanceUpdate:
		res, err = h.callDiscordStageInstanceUpdate(ctx, app, req.Data.(distype.StageInstanceModifyRequest))
	case call.DiscordStageInstanceDelete:
		res, err = h.callDiscordStageInstanceDelete(ctx, app, req.Data.(distype.StageInstanceDeleteRequest))
	case call.DiscordStickerList:
		res, err = h.callDiscordStickerList(ctx, app, req.Data.(distype.GuildStickerListRequest))
	case call.DiscordStickerGet:
		res, err = h.callDiscordStickerGet(ctx, app, req.Data.(distype.StickerGetRequest))
	case call.DiscordStickerCreate:
		res, err = h.callDiscordStickerCreate(ctx, app, req.Data.(distype.StickerCreateRequest))
	case call.DiscordStickerUpdate:
		res, err = h.callDiscordStickerUpdate(ctx, app, req.Data.(distype.StickerModifyRequest))
	case call.DiscordStickerDelete:
		res, err = h.callDiscordStickerDelete(ctx, app, req.Data.(distype.StickerDeleteRequest))
	case call.DiscordUserGet:
		res, err = h.callDiscordUserGet(ctx, app, req.Data.(distype.UserGetRequest))
	case call.DiscordWebhookGet:
		res, err = h.callDiscordWebhookGet(ctx, app, req.Data.(distype.WebhookGetRequest))
	case call.DiscordWebhookListForChannel:
		res, err = h.callDiscordWebhookListForChannel(ctx, app, req.Data.(distype.ChannelWebhookListRequest))
	case call.DiscordWebhookListForGuild:
		res, err = h.callDiscordWebhookListForGuild(ctx, app, req.Data.(distype.GuildWebhookListRequest))
	case call.DiscordWebhookCreate:
		res, err = h.callDiscordWebhookCreate(ctx, app, req.Data.(distype.WebhookCreateRequest))
	case call.DiscordWebhookUpdate:
		res, err = h.callDiscordWebhookUpdate(ctx, app, req.Data.(distype.WebhookModifyRequest))
	case call.DiscordWebhookDelete:
		res, err = h.callDiscordWebhookDelete(ctx, app, req.Data.(distype.WebhookDeleteRequest))
	case call.DiscordWebhookGetWithToken:
		res, err = h.callDiscordWebhookGetWithToken(ctx, app, req.Data.(distype.WebhookGetWithTokenRequest))
	case call.DiscordWebhookUpdateWithToken:
		res, err = h.callDiscordWebhookUpdateWithToken(ctx, app, req.Data.(distype.WebhookModifyWithTokenRequest))
	case call.DiscordWebhookDeleteWithToken:
		res, err = h.callDiscordWebhookDeleteWithToken(ctx, app, req.Data.(distype.WebhookDeleteWithTokenRequest))
	case call.DiscordWebhookExecute:
		res, err = h.callDiscordWebhookExecute(ctx, app, req.Data.(distype.WebhookExecuteRequest))
	default:
		return nil, fmt.Errorf("unknown call type: %s", req.Type)
	}

	if err != nil {
		return nil, modelError(err)
	}

	return res, nil
}
