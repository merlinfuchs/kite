package call

import (
	"encoding/json"

	"github.com/merlinfuchs/dismod/distype"

	"github.com/merlinfuchs/kite/kite-sdk-go/fail"
	"github.com/merlinfuchs/kite/kite-sdk-go/internal/util"
	"github.com/merlinfuchs/kite/kite-sdk-go/kv"
)

type CallType string

const (
	Sleep                                  CallType = "SLEEP"
	KVKeyGet                               CallType = "KV_KEY_GET"
	KVKeySet                               CallType = "KV_KEY_SET"
	KVKeyDelete                            CallType = "KV_KEY_DELETE"
	KVKeyIncrease                          CallType = "KV_KEY_INCREASE"
	DiscordBanList                         CallType = "DISCORD_BAN_LIST"
	DiscordBanGet                          CallType = "DISCORD_BAN_GET"
	DiscordBanCreate                       CallType = "DISCORD_BAN_CREATE"
	DiscordBanRemove                       CallType = "DISCORD_BAN_REMOVE"
	DiscordChannelGet                      CallType = "DISCORD_CHANNEL_GET"
	DiscordChannelList                     CallType = "DISCORD_CHANNEL_LIST"
	DiscordChannelCreate                   CallType = "DISCORD_CHANNEL_CREATE"
	DiscordChannelUpdate                   CallType = "DISCORD_CHANNEL_UPDATE"
	DiscordChannelUpdatePositions          CallType = "DISCORD_CHANNEL_UPDATE_POSITIONS"
	DiscordChannelDelete                   CallType = "DISCORD_CHANNEL_DELETE"
	DiscordChannelUpdatePermissions        CallType = "DISCORD_CHANNEL_UPDATE_PERMISSIONS"
	DiscordChannelDeletePermissions        CallType = "DISCORD_CHANNEL_DELETE_PERMISSIONS"
	DiscordThreadStartFromMessage          CallType = "DISCORD_THREAD_START_FROM_MESSAGE"
	DiscordThreadStart                     CallType = "DISCORD_THREAD_START"
	DiscordThreadStartInForum              CallType = "DISCORD_THREAD_START_IN_FORUM"
	DiscordThreadJoin                      CallType = "DISCORD_THREAD_JOIN"
	DiscordThreadMemberAdd                 CallType = "DISCORD_THREAD_ADD_MEMBER"
	DiscordThreadLeave                     CallType = "DISCORD_THREAD_LEAVE"
	DiscordThreadMemberRemove              CallType = "DISCORD_THREAD_REMOVE_MEMBER"
	DiscordThreadMemberGet                 CallType = "DISCORD_THREAD_MEMBER_GET"
	DiscordThreadMemberList                CallType = "DISCORD_THREAD_MEMBER_LIST"
	DiscordThreadListPublicArchived        CallType = "DISCORD_THREAD_LIST_PUBLIC_ARCHIVED"
	DiscordThreadListPrivateArchived       CallType = "DISCORD_THREAD_LIST_PRIVATE_ARCHIVED"
	DiscordThreadListJoinedPrivateArchived CallType = "DISCORD_THREAD_LIST_JOINED_PRIVATE_ARCHIVED"
	DiscordThreadListActive                CallType = "DISCORD_THREAD_LIST_ACTIVE"
	DiscordEmojiList                       CallType = "DISCORD_EMOJI_LIST"
	DiscordEmojiGet                        CallType = "DISCORD_EMOJI_GET"
	DiscordEmojiCreate                     CallType = "DISCORD_EMOJI_CREATE"
	DiscordEmojiUpdate                     CallType = "DISCORD_EMOJI_UPDATE"
	DiscordEmojiDelete                     CallType = "DISCORD_EMOJI_DELETE"
	DiscordGuildGet                        CallType = "DISCORD_GUILD_GET"
	DiscordGuildUpdate                     CallType = "DISCORD_GUILD_UPDATE"
	DiscordInteractionResponseCreate       CallType = "DISCORD_INTERACTION_RESPONSE_CREATE"
	DiscordInteractionResponseUpdate       CallType = "DISCORD_INTERACTION_RESPONSE_UPDATE"
	DiscordInteractionResponseDelete       CallType = "DISCORD_INTERACTION_RESPONSE_DELETE"
	DiscordInteractionResponseGet          CallType = "DISCORD_INTERACTION_RESPONSE_GET"
	DiscordInteractionFollowupCreate       CallType = "DISCORD_INTERACTION_FOLLOWUP_CREATE"
	DiscordInteractionFollowupUpdate       CallType = "DISCORD_INTERACTION_FOLLOWUP_UPDATE"
	DiscordInteractionFollowupDelete       CallType = "DISCORD_INTERACTION_FOLLOWUP_DELETE"
	DiscordInteractionFollowupGet          CallType = "DISCORD_INTERACTION_FOLLOWUP_GET"
	DiscordInviteListForChannel            CallType = "DISCORD_INVITE_LIST_FOR_CHANNEL"
	DiscordInviteListForGuild              CallType = "DISCORD_INVITE_LIST_FOR_GUILD"
	DiscordInviteCreate                    CallType = "DISCORD_INVITE_CREATE"
	DiscordInviteGet                       CallType = "DISCORD_INVITE_GET"
	DiscordInviteDelete                    CallType = "DISCORD_INVITE_DELETE"
	DiscordMemberGet                       CallType = "DISCORD_MEMBER_GET"
	DiscordMemberList                      CallType = "DISCORD_MEMBER_LIST"
	DiscordMemberSearch                    CallType = "DISCORD_MEMBER_SEARCH"
	DiscordMemberUpdate                    CallType = "DISCORD_MEMBER_UPDATE"
	DiscordMemberUpdateOwn                 CallType = "DISCORD_MEMBER_UPDATE_OWN"
	DiscordMemberRoleAdd                   CallType = "DISCORD_MEMBER_ADD_ROLE"
	DiscordMemberRoleRemove                CallType = "DISCORD_MEMBER_REMOVE_ROLE"
	DiscordMemberRemove                    CallType = "DISCORD_MEMBER_REMOVE"
	DiscordMemberPruneCount                CallType = "DISCORD_MEMBER_PRUNE_COUNT"
	DiscordMemberPruneBegin                CallType = "DISCORD_MEMBER_PRUNE_BEGIN"
	DiscordMessageList                     CallType = "DISCORD_MESSAGE_LIST"
	DiscordMessageGet                      CallType = "DISCORD_MESSAGE_GET"
	DiscordMessageCreate                   CallType = "DISCORD_MESSAGE_CREATE"
	DiscordMessageUpdate                   CallType = "DISCORD_MESSAGE_UPDATE"
	DiscordMessageDelete                   CallType = "DISCORD_MESSAGE_DELETE"
	DiscordMessageDeleteBulk               CallType = "DISCORD_MESSAGE_DELETE_BULK"
	DiscordMessageReactionCreate           CallType = "DISCORD_MESSAGE_REACTION_CREATE"
	DiscordMessageReactionDeleteOwn        CallType = "DISCORD_MESSAGE_REACTION_DELETE_OWN"
	DiscordMessageReactionDeleteUser       CallType = "DISCORD_MESSAGE_REACTION_DELETE_USER"
	DiscordMessageReactionList             CallType = "DISCORD_MESSAGE_REACTION_LIST"
	DiscordMessageReactionDeleteAll        CallType = "DISCORD_MESSAGE_REACTION_DELETE_ALL"
	DiscordMessageReactionDeleteEmoji      CallType = "DISCORD_MESSAGE_REACTION_DELETE_EMOJI"
	DiscordMessageGetPinned                CallType = "DISCORD_MESSAGE_GET_PINNED"
	DiscordMessagePin                      CallType = "DISCORD_MESSAGE_PIN"
	DiscordMessageUnpin                    CallType = "DISCORD_MESSAGE_UNPIN"
	DiscordRoleList                        CallType = "DISCORD_ROLE_LIST"
	DiscordRoleCreate                      CallType = "DISCORD_ROLE_CREATE"
	DiscordRoleUpdate                      CallType = "DISCORD_ROLE_UPDATE"
	DiscordRoleUpdatePositions             CallType = "DISCORD_ROLE_UPDATE_POSITIONS"
	DiscordRoleDelete                      CallType = "DISCORD_ROLE_DELETE"
	DiscordScheduledEventList              CallType = "DISCORD_SCHEDULED_EVENT_LIST"
	DiscordScheduledEventCreate            CallType = "DISCORD_SCHEDULED_EVENT_CREATE"
	DiscordScheduledEventGet               CallType = "DISCORD_SCHEDULED_EVENT_GET"
	DiscordScheduledEventUpdate            CallType = "DISCORD_SCHEDULED_EVENT_UPDATE"
	DiscordScheduledEventDelete            CallType = "DISCORD_SCHEDULED_EVENT_DELETE"
	DiscordScheduledEventUserList          CallType = "DISCORD_SCHEDULED_EVENT_USER_LIST"
	DiscordStageInstanceCreate             CallType = "DISCORD_STAGE_INSTANCE_CREATE"
	DiscordStageInstanceGet                CallType = "DISCORD_STAGE_INSTANCE_GET"
	DiscordStageInstanceUpdate             CallType = "DISCORD_STAGE_INSTANCE_UPDATE"
	DiscordStageInstanceDelete             CallType = "DISCORD_STAGE_INSTANCE_DELETE"
	DiscordStickerList                     CallType = "DISCORD_STICKER_LIST"
	DiscordStickerGet                      CallType = "DISCORD_STICKER_GET"
	DiscordStickerCreate                   CallType = "DISCORD_STICKER_CREATE"
	DiscordStickerUpdate                   CallType = "DISCORD_STICKER_UPDATE"
	DiscordStickerDelete                   CallType = "DISCORD_STICKER_DELETE"
	DiscordUserGet                         CallType = "DISCORD_USER_GET"
	DiscordWebhookGet                      CallType = "DISCORD_WEBHOOK_GET"
	DiscordWebhookListForChannel           CallType = "DISCORD_WEBHOOK_LIST_FOR_CHANNEL"
	DiscordWebhookListForGuild             CallType = "DISCORD_WEBHOOK_LIST_FOR_GUILD"
	DiscordWebhookCreate                   CallType = "DISCORD_WEBHOOK_CREATE"
	DiscordWebhookUpdate                   CallType = "DISCORD_WEBHOOK_UPDATE"
	DiscordWebhookDelete                   CallType = "DISCORD_WEBHOOK_DELETE"
	DiscordWebhookGetWithToken             CallType = "DISCORD_WEBHOOK_GET_WITH_TOKEN"
	DiscordWebhookUpdateWithToken          CallType = "DISCORD_WEBHOOK_UPDATE_WITH_TOKEN"
	DiscordWebhookDeleteWithToken          CallType = "DISCORD_WEBHOOK_DELETE_WITH_TOKEN"
	DiscordWebhookExecute                  CallType = "DISCORD_WEBHOOK_EXECUTE"
)

type Call struct {
	Type   CallType    `json:"type"`
	Config CallConfig  `json:"config"`
	Data   interface{} `json:"data"`
}

func (c *Call) UnmarshalJSON(b []byte) error {
	var temp struct {
		Type   CallType        `json:"type"`
		Config CallConfig      `json:"config"`
		Data   json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(b, &temp); err != nil {
		return err
	}

	c.Type = temp.Type
	c.Config = temp.Config
	var err error

	switch c.Type {
	case Sleep:
		c.Data, err = util.DecodeT[SleepCall](temp.Data)
	case KVKeyGet:
		c.Data, err = util.DecodeT[kv.KVKeyGetCall](temp.Data)
	case KVKeySet:
		c.Data, err = util.DecodeT[kv.KVKeySetCall](temp.Data)
	case KVKeyDelete:
		c.Data, err = util.DecodeT[kv.KVKeyDeleteCall](temp.Data)
	case KVKeyIncrease:
		c.Data, err = util.DecodeT[kv.KVKeyIncreaseCall](temp.Data)
	case DiscordBanList:
		c.Data, err = util.DecodeT[distype.BanListRequest](temp.Data)
	case DiscordBanGet:
		c.Data, err = util.DecodeT[distype.BanGetRequest](temp.Data)
	case DiscordBanCreate:
		c.Data, err = util.DecodeT[distype.BanCreateRequest](temp.Data)
	case DiscordBanRemove:
		c.Data, err = util.DecodeT[distype.BanRemoveRequest](temp.Data)
	case DiscordChannelGet:
		c.Data, err = util.DecodeT[distype.ChannelGetRequest](temp.Data)
	case DiscordChannelList:
		c.Data, err = util.DecodeT[distype.GuildChannelListRequest](temp.Data)
	case DiscordChannelCreate:
		c.Data, err = util.DecodeT[distype.GuildChannelCreateRequest](temp.Data)
	case DiscordChannelUpdate:
		c.Data, err = util.DecodeT[distype.ChannelModifyRequest](temp.Data)
	case DiscordChannelUpdatePositions:
		c.Data, err = util.DecodeT[distype.GuildChannelModifyPositionsRequest](temp.Data)
	case DiscordChannelDelete:
		c.Data, err = util.DecodeT[distype.ChannelDeleteRequest](temp.Data)
	case DiscordChannelUpdatePermissions:
		c.Data, err = util.DecodeT[distype.ChannelEditPermissionsRequest](temp.Data)
	case DiscordChannelDeletePermissions:
		c.Data, err = util.DecodeT[distype.ChannelDeletePermissionsRequest](temp.Data)
	case DiscordThreadStartFromMessage:
		c.Data, err = util.DecodeT[distype.ThreadStartFromMessageRequest](temp.Data)
	case DiscordThreadStart:
		c.Data, err = util.DecodeT[distype.ThreadStartWithoutMessageRequest](temp.Data)
	case DiscordThreadStartInForum:
		c.Data, err = util.DecodeT[distype.ThreadStartInForumRequest](temp.Data)
	case DiscordThreadJoin:
		c.Data, err = util.DecodeT[distype.ThreadJoinRequest](temp.Data)
	case DiscordThreadMemberAdd:
		c.Data, err = util.DecodeT[distype.ThreadMemberAddRequest](temp.Data)
	case DiscordThreadLeave:
		c.Data, err = util.DecodeT[distype.ThreadLeaveRequest](temp.Data)
	case DiscordThreadMemberRemove:
		c.Data, err = util.DecodeT[distype.ThreadMemberRemoveRequest](temp.Data)
	case DiscordThreadMemberGet:
		c.Data, err = util.DecodeT[distype.ThreadMemberGetRequest](temp.Data)
	case DiscordThreadMemberList:
		c.Data, err = util.DecodeT[distype.ThreadMemberListRequest](temp.Data)
	case DiscordThreadListPublicArchived:
		c.Data, err = util.DecodeT[distype.ThreadListPublicArchivedRequest](temp.Data)
	case DiscordThreadListPrivateArchived:
		c.Data, err = util.DecodeT[distype.ThreadListPrivateArchivedRequest](temp.Data)
	case DiscordThreadListJoinedPrivateArchived:
		c.Data, err = util.DecodeT[distype.ThreadListJoinedPrivateArchivedRequest](temp.Data)
	case DiscordThreadListActive:
		c.Data, err = util.DecodeT[distype.GuildThreadListActiveRequest](temp.Data)
	case DiscordEmojiList:
		c.Data, err = util.DecodeT[distype.GuildEmojiListRequest](temp.Data)
	case DiscordEmojiGet:
		c.Data, err = util.DecodeT[distype.EmojiGetRequest](temp.Data)
	case DiscordEmojiCreate:
		c.Data, err = util.DecodeT[distype.EmojiCreateRequest](temp.Data)
	case DiscordEmojiUpdate:
		c.Data, err = util.DecodeT[distype.EmojiModifyRequest](temp.Data)
	case DiscordEmojiDelete:
		c.Data, err = util.DecodeT[distype.EmojiDeleteRequest](temp.Data)
	case DiscordGuildGet:
		c.Data, err = util.DecodeT[distype.GuildGetRequest](temp.Data)
	case DiscordGuildUpdate:
		c.Data, err = util.DecodeT[distype.GuildUpdateRequest](temp.Data)
	case DiscordInteractionResponseCreate:
		c.Data, err = util.DecodeT[distype.InteractionResponseCreateRequest](temp.Data)
	case DiscordInteractionResponseUpdate:
		c.Data, err = util.DecodeT[distype.InteractionResponseEditRequest](temp.Data)
	case DiscordInteractionResponseDelete:
		c.Data, err = util.DecodeT[distype.InteractionResponseDeleteRequest](temp.Data)
	case DiscordInteractionResponseGet:
		c.Data, err = util.DecodeT[distype.InteractionResponseGetRequest](temp.Data)
	case DiscordInteractionFollowupCreate:
		c.Data, err = util.DecodeT[distype.InteractionFollowupCreateRequest](temp.Data)
	case DiscordInteractionFollowupUpdate:
		c.Data, err = util.DecodeT[distype.InteractionFollowupEditRequest](temp.Data)
	case DiscordInteractionFollowupDelete:
		c.Data, err = util.DecodeT[distype.InteractionFollowupDeleteRequest](temp.Data)
	case DiscordInteractionFollowupGet:
		c.Data, err = util.DecodeT[distype.InteractionFollowupGetRequest](temp.Data)
	case DiscordInviteListForChannel:
		c.Data, err = util.DecodeT[distype.ChannelInviteListRequest](temp.Data)
	case DiscordInviteListForGuild:
		c.Data, err = util.DecodeT[distype.GuildInviteListRequest](temp.Data)
	case DiscordInviteCreate:
		c.Data, err = util.DecodeT[distype.ChannelInviteCreateRequest](temp.Data)
	case DiscordInviteGet:
		c.Data, err = util.DecodeT[distype.InviteGetRequest](temp.Data)
	case DiscordInviteDelete:
		c.Data, err = util.DecodeT[distype.InviteDeleteRequest](temp.Data)
	case DiscordMemberGet:
		c.Data, err = util.DecodeT[distype.MemberGetRequest](temp.Data)
	case DiscordMemberList:
		c.Data, err = util.DecodeT[distype.GuildMemberListRequest](temp.Data)
	case DiscordMemberSearch:
		c.Data, err = util.DecodeT[distype.GuildMemberSearchRequest](temp.Data)
	case DiscordMemberUpdate:
		c.Data, err = util.DecodeT[distype.MemberModifyRequest](temp.Data)
	case DiscordMemberUpdateOwn:
		c.Data, err = util.DecodeT[distype.MemberModifyCurrentRequest](temp.Data)
	case DiscordMemberRoleAdd:
		c.Data, err = util.DecodeT[distype.MemberRoleAddRequest](temp.Data)
	case DiscordMemberRoleRemove:
		c.Data, err = util.DecodeT[distype.MemberRoleRemoveRequest](temp.Data)
	case DiscordMemberRemove:
		c.Data, err = util.DecodeT[distype.MemberRemoveRequest](temp.Data)
	case DiscordMemberPruneCount:
		c.Data, err = util.DecodeT[distype.MemberPruneCountRequest](temp.Data)
	case DiscordMemberPruneBegin:
		c.Data, err = util.DecodeT[distype.MemberPruneRequest](temp.Data)
	case DiscordMessageList:
		c.Data, err = util.DecodeT[distype.ChannelMessageListRequest](temp.Data)
	case DiscordMessageGet:
		c.Data, err = util.DecodeT[distype.MessageGetRequest](temp.Data)
	case DiscordMessageCreate:
		c.Data, err = util.DecodeT[distype.MessageCreateRequest](temp.Data)
	case DiscordMessageUpdate:
		c.Data, err = util.DecodeT[distype.MessageEditRequest](temp.Data)
	case DiscordMessageDelete:
		c.Data, err = util.DecodeT[distype.MessageDeleteRequest](temp.Data)
	case DiscordMessageDeleteBulk:
		c.Data, err = util.DecodeT[distype.MessageBulkDeleteRequest](temp.Data)
	case DiscordMessageReactionCreate:
		c.Data, err = util.DecodeT[distype.MessageReactionCreateRequest](temp.Data)
	case DiscordMessageReactionDeleteOwn:
		c.Data, err = util.DecodeT[distype.MessageReactionDeleteOwnRequest](temp.Data)
	case DiscordMessageReactionDeleteUser:
		c.Data, err = util.DecodeT[distype.MessageReactionDeleteRequest](temp.Data)
	case DiscordMessageReactionList:
		c.Data, err = util.DecodeT[distype.MessageReactionListRequest](temp.Data)
	case DiscordMessageReactionDeleteAll:
		c.Data, err = util.DecodeT[distype.MessageReactionDeleteAllRequest](temp.Data)
	case DiscordMessageReactionDeleteEmoji:
		c.Data, err = util.DecodeT[distype.MessageReactionDeleteEmojiRequest](temp.Data)
	case DiscordMessageGetPinned:
		c.Data, err = util.DecodeT[distype.ChannelPinnedMessageListRequest](temp.Data)
	case DiscordMessagePin:
		c.Data, err = util.DecodeT[distype.MessagePinRequest](temp.Data)
	case DiscordMessageUnpin:
		c.Data, err = util.DecodeT[distype.MessageUnpinRequest](temp.Data)
	case DiscordRoleList:
		c.Data, err = util.DecodeT[distype.GuildRoleListRequest](temp.Data)
	case DiscordRoleCreate:
		c.Data, err = util.DecodeT[distype.RoleCreateRequest](temp.Data)
	case DiscordRoleUpdate:
		c.Data, err = util.DecodeT[distype.RoleModifyRequest](temp.Data)
	case DiscordRoleUpdatePositions:
		c.Data, err = util.DecodeT[distype.RolePositionsModifyRequest](temp.Data)
	case DiscordRoleDelete:
		c.Data, err = util.DecodeT[distype.RoleDeleteRequest](temp.Data)
	case DiscordScheduledEventList:
		c.Data, err = util.DecodeT[distype.GuildScheduledEventListRequest](temp.Data)
	case DiscordScheduledEventCreate:
		c.Data, err = util.DecodeT[distype.ScheduledEventCreateRequest](temp.Data)
	case DiscordScheduledEventGet:
		c.Data, err = util.DecodeT[distype.ScheduledEventGetRequest](temp.Data)
	case DiscordScheduledEventUpdate:
		c.Data, err = util.DecodeT[distype.ScheduledEventModifyRequest](temp.Data)
	case DiscordScheduledEventDelete:
		c.Data, err = util.DecodeT[distype.ScheduledEventDeleteRequest](temp.Data)
	case DiscordScheduledEventUserList:
		c.Data, err = util.DecodeT[distype.ScheduledEventUserListRequest](temp.Data)
	case DiscordStageInstanceCreate:
		c.Data, err = util.DecodeT[distype.StageInstanceCreateRequest](temp.Data)
	case DiscordStageInstanceGet:
		c.Data, err = util.DecodeT[distype.StageInstanceGetRequest](temp.Data)
	case DiscordStageInstanceUpdate:
		c.Data, err = util.DecodeT[distype.StageInstanceModifyRequest](temp.Data)
	case DiscordStageInstanceDelete:
		c.Data, err = util.DecodeT[distype.StageInstanceDeleteRequest](temp.Data)
	case DiscordStickerList:
		c.Data, err = util.DecodeT[distype.GuildStickerListRequest](temp.Data)
	case DiscordStickerGet:
		c.Data, err = util.DecodeT[distype.StickerGetRequest](temp.Data)
	case DiscordStickerCreate:
		c.Data, err = util.DecodeT[distype.StickerCreateRequest](temp.Data)
	case DiscordStickerUpdate:
		c.Data, err = util.DecodeT[distype.StickerModifyRequest](temp.Data)
	case DiscordStickerDelete:
		c.Data, err = util.DecodeT[distype.StickerDeleteRequest](temp.Data)
	case DiscordUserGet:
		c.Data, err = util.DecodeT[distype.UserGetRequest](temp.Data)
	case DiscordWebhookGet:
		c.Data, err = util.DecodeT[distype.WebhookGetRequest](temp.Data)
	case DiscordWebhookListForChannel:
		c.Data, err = util.DecodeT[distype.ChannelWebhookListRequest](temp.Data)
	case DiscordWebhookListForGuild:
		c.Data, err = util.DecodeT[distype.GuildWebhookListRequest](temp.Data)
	case DiscordWebhookCreate:
		c.Data, err = util.DecodeT[distype.WebhookCreateRequest](temp.Data)
	case DiscordWebhookUpdate:
		c.Data, err = util.DecodeT[distype.WebhookModifyRequest](temp.Data)
	case DiscordWebhookDelete:
		c.Data, err = util.DecodeT[distype.WebhookDeleteRequest](temp.Data)
	case DiscordWebhookGetWithToken:
		c.Data, err = util.DecodeT[distype.WebhookGetWithTokenRequest](temp.Data)
	case DiscordWebhookUpdateWithToken:
		c.Data, err = util.DecodeT[distype.WebhookModifyWithTokenRequest](temp.Data)
	case DiscordWebhookDeleteWithToken:
		c.Data, err = util.DecodeT[distype.WebhookDeleteWithTokenRequest](temp.Data)
	case DiscordWebhookExecute:
		c.Data, err = util.DecodeT[distype.WebhookExecuteRequest](temp.Data)
	}

	return err
}

type CallResponse struct {
	Success bool            `json:"success"`
	Error   *fail.HostError `json:"error"`
	Data    interface{}     `json:"data"`
}

func (c *CallResponse) UnmarshalJSON(b []byte) error {
	var temp struct {
		Success bool            `json:"success"`
		Error   *fail.HostError `json:"error"`
		Data    json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(b, &temp); err != nil {
		return err
	}

	c.Success = temp.Success
	c.Error = temp.Error
	c.Data = temp.Data

	return nil
}
