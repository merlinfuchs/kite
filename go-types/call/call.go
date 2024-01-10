package call

import (
	"encoding/json"

	"github.com/merlinfuchs/kite/go-types/dismodel"
	"github.com/merlinfuchs/kite/go-types/fail"
	"github.com/merlinfuchs/kite/go-types/internal"
	"github.com/merlinfuchs/kite/go-types/kvmodel"
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
		c.Data, err = internal.DecodeT[SleepCall](temp.Data)
	case KVKeyGet:
		c.Data, err = internal.DecodeT[kvmodel.KVKeyGetCall](temp.Data)
	case KVKeySet:
		c.Data, err = internal.DecodeT[kvmodel.KVKeySetCall](temp.Data)
	case KVKeyDelete:
		c.Data, err = internal.DecodeT[kvmodel.KVKeyDeleteCall](temp.Data)
	case KVKeyIncrease:
		c.Data, err = internal.DecodeT[kvmodel.KVKeyIncreaseCall](temp.Data)
	case DiscordBanList:
		c.Data, err = internal.DecodeT[dismodel.BanListCall](temp.Data)
	case DiscordBanGet:
		c.Data, err = internal.DecodeT[dismodel.BanGetCall](temp.Data)
	case DiscordBanCreate:
		c.Data, err = internal.DecodeT[dismodel.BanCreateCall](temp.Data)
	case DiscordBanRemove:
		c.Data, err = internal.DecodeT[dismodel.BanRemoveCall](temp.Data)
	case DiscordChannelGet:
		c.Data, err = internal.DecodeT[dismodel.ChannelGetCall](temp.Data)
	case DiscordChannelList:
		c.Data, err = internal.DecodeT[dismodel.ChannelListCall](temp.Data)
	case DiscordChannelCreate:
		c.Data, err = internal.DecodeT[dismodel.ChannelCreateCall](temp.Data)
	case DiscordChannelUpdate:
		c.Data, err = internal.DecodeT[dismodel.ChannelUpdateCall](temp.Data)
	case DiscordChannelUpdatePositions:
		c.Data, err = internal.DecodeT[dismodel.ChannelUpdatePositionsCall](temp.Data)
	case DiscordChannelDelete:
		c.Data, err = internal.DecodeT[dismodel.ChannelDeleteCall](temp.Data)
	case DiscordChannelUpdatePermissions:
		c.Data, err = internal.DecodeT[dismodel.ChannelUpdatePermissionsCall](temp.Data)
	case DiscordChannelDeletePermissions:
		c.Data, err = internal.DecodeT[dismodel.ChannelDeletePermissionsCall](temp.Data)
	case DiscordThreadStartFromMessage:
		c.Data, err = internal.DecodeT[dismodel.ThreadStartFromMessageCall](temp.Data)
	case DiscordThreadStart:
		c.Data, err = internal.DecodeT[dismodel.ThreadStartCall](temp.Data)
	case DiscordThreadStartInForum:
		c.Data, err = internal.DecodeT[dismodel.ThreadStartInForumCall](temp.Data)
	case DiscordThreadJoin:
		c.Data, err = internal.DecodeT[dismodel.ThreadJoinCall](temp.Data)
	case DiscordThreadMemberAdd:
		c.Data, err = internal.DecodeT[dismodel.ThreadMemberAddCall](temp.Data)
	case DiscordThreadLeave:
		c.Data, err = internal.DecodeT[dismodel.ThreadLeaveCall](temp.Data)
	case DiscordThreadMemberRemove:
		c.Data, err = internal.DecodeT[dismodel.ThreadMemberRemoveCall](temp.Data)
	case DiscordThreadMemberGet:
		c.Data, err = internal.DecodeT[dismodel.ThreadMemberGetCall](temp.Data)
	case DiscordThreadMemberList:
		c.Data, err = internal.DecodeT[dismodel.ThreadMemberListCall](temp.Data)
	case DiscordThreadListPublicArchived:
		c.Data, err = internal.DecodeT[dismodel.ThreadListPublicArchivedCall](temp.Data)
	case DiscordThreadListPrivateArchived:
		c.Data, err = internal.DecodeT[dismodel.ThreadListPrivateArchivedCall](temp.Data)
	case DiscordThreadListJoinedPrivateArchived:
		c.Data, err = internal.DecodeT[dismodel.ThreadListJoinedPrivateArchivedCall](temp.Data)
	case DiscordThreadListActive:
		c.Data, err = internal.DecodeT[dismodel.ThreadListActiveCall](temp.Data)
	case DiscordEmojiList:
		c.Data, err = internal.DecodeT[dismodel.EmojiListCall](temp.Data)
	case DiscordEmojiGet:
		c.Data, err = internal.DecodeT[dismodel.EmojiGetCall](temp.Data)
	case DiscordEmojiCreate:
		c.Data, err = internal.DecodeT[dismodel.EmojiCreateCall](temp.Data)
	case DiscordEmojiUpdate:
		c.Data, err = internal.DecodeT[dismodel.EmojiUpdateCall](temp.Data)
	case DiscordEmojiDelete:
		c.Data, err = internal.DecodeT[dismodel.EmojiDeleteCall](temp.Data)
	case DiscordGuildGet:
		c.Data, err = internal.DecodeT[dismodel.GuildGetCall](temp.Data)
	case DiscordGuildUpdate:
		c.Data, err = internal.DecodeT[dismodel.GuildUpdateCall](temp.Data)
	case DiscordInteractionResponseCreate:
		c.Data, err = internal.DecodeT[dismodel.InteractionResponseCreateCall](temp.Data)
	case DiscordInteractionResponseUpdate:
		c.Data, err = internal.DecodeT[dismodel.InteractionResponseUpdateCall](temp.Data)
	case DiscordInteractionResponseDelete:
		c.Data, err = internal.DecodeT[dismodel.InteractionResponseDeleteCall](temp.Data)
	case DiscordInteractionResponseGet:
		c.Data, err = internal.DecodeT[dismodel.InteractionResponseGetCall](temp.Data)
	case DiscordInteractionFollowupCreate:
		c.Data, err = internal.DecodeT[dismodel.InteractionFollowupCreateCall](temp.Data)
	case DiscordInteractionFollowupUpdate:
		c.Data, err = internal.DecodeT[dismodel.InteractionFollowupUpdateCall](temp.Data)
	case DiscordInteractionFollowupDelete:
		c.Data, err = internal.DecodeT[dismodel.InteractionFollowupDeleteCall](temp.Data)
	case DiscordInteractionFollowupGet:
		c.Data, err = internal.DecodeT[dismodel.InteractionFollowupGetCall](temp.Data)
	case DiscordInviteListForChannel:
		c.Data, err = internal.DecodeT[dismodel.InviteListForChannelCall](temp.Data)
	case DiscordInviteListForGuild:
		c.Data, err = internal.DecodeT[dismodel.InviteListForGuildCall](temp.Data)
	case DiscordInviteCreate:
		c.Data, err = internal.DecodeT[dismodel.InviteCreateCall](temp.Data)
	case DiscordInviteGet:
		c.Data, err = internal.DecodeT[dismodel.InviteGetCall](temp.Data)
	case DiscordInviteDelete:
		c.Data, err = internal.DecodeT[dismodel.InviteDeleteCall](temp.Data)
	case DiscordMemberGet:
		c.Data, err = internal.DecodeT[dismodel.MemberGetCall](temp.Data)
	case DiscordMemberList:
		c.Data, err = internal.DecodeT[dismodel.MemberListCall](temp.Data)
	case DiscordMemberSearch:
		c.Data, err = internal.DecodeT[dismodel.MemberSearchCall](temp.Data)
	case DiscordMemberUpdate:
		c.Data, err = internal.DecodeT[dismodel.MemberUpdateCall](temp.Data)
	case DiscordMemberUpdateOwn:
		c.Data, err = internal.DecodeT[dismodel.MemberUpdateOwnCall](temp.Data)
	case DiscordMemberRoleAdd:
		c.Data, err = internal.DecodeT[dismodel.MemberRoleAddCall](temp.Data)
	case DiscordMemberRoleRemove:
		c.Data, err = internal.DecodeT[dismodel.MemberRoleRemoveCall](temp.Data)
	case DiscordMemberRemove:
		c.Data, err = internal.DecodeT[dismodel.MemberRemoveCall](temp.Data)
	case DiscordMemberPruneCount:
		c.Data, err = internal.DecodeT[dismodel.MemberPruneCountCall](temp.Data)
	case DiscordMemberPruneBegin:
		c.Data, err = internal.DecodeT[dismodel.MemberPruneBeginCall](temp.Data)
	case DiscordMessageList:
		c.Data, err = internal.DecodeT[dismodel.MessageListCall](temp.Data)
	case DiscordMessageGet:
		c.Data, err = internal.DecodeT[dismodel.MessageGetCall](temp.Data)
	case DiscordMessageCreate:
		c.Data, err = internal.DecodeT[dismodel.MessageCreateCall](temp.Data)
	case DiscordMessageUpdate:
		c.Data, err = internal.DecodeT[dismodel.MessageUpdateCall](temp.Data)
	case DiscordMessageDelete:
		c.Data, err = internal.DecodeT[dismodel.MessageDeleteCall](temp.Data)
	case DiscordMessageDeleteBulk:
		c.Data, err = internal.DecodeT[dismodel.MessageDeleteBulkCall](temp.Data)
	case DiscordMessageReactionCreate:
		c.Data, err = internal.DecodeT[dismodel.MessageReactionCreateCall](temp.Data)
	case DiscordMessageReactionDeleteOwn:
		c.Data, err = internal.DecodeT[dismodel.MessageReactionDeleteOwnCall](temp.Data)
	case DiscordMessageReactionDeleteUser:
		c.Data, err = internal.DecodeT[dismodel.MessageReactionDeleteUserCall](temp.Data)
	case DiscordMessageReactionList:
		c.Data, err = internal.DecodeT[dismodel.MessageReactionListCall](temp.Data)
	case DiscordMessageReactionDeleteAll:
		c.Data, err = internal.DecodeT[dismodel.MessageReactionDeleteAllCall](temp.Data)
	case DiscordMessageReactionDeleteEmoji:
		c.Data, err = internal.DecodeT[dismodel.MessageReactionDeleteEmojiCall](temp.Data)
	case DiscordMessageGetPinned:
		c.Data, err = internal.DecodeT[dismodel.MessageGetPinnedCall](temp.Data)
	case DiscordMessagePin:
		c.Data, err = internal.DecodeT[dismodel.MessagePinCall](temp.Data)
	case DiscordMessageUnpin:
		c.Data, err = internal.DecodeT[dismodel.MessageUnpinCall](temp.Data)
	case DiscordRoleList:
		c.Data, err = internal.DecodeT[dismodel.RoleListCall](temp.Data)
	case DiscordRoleCreate:
		c.Data, err = internal.DecodeT[dismodel.RoleCreateCall](temp.Data)
	case DiscordRoleUpdate:
		c.Data, err = internal.DecodeT[dismodel.RoleUpdateCall](temp.Data)
	case DiscordRoleUpdatePositions:
		c.Data, err = internal.DecodeT[dismodel.RoleUpdatePositionsCall](temp.Data)
	case DiscordRoleDelete:
		c.Data, err = internal.DecodeT[dismodel.RoleDeleteCall](temp.Data)
	case DiscordScheduledEventList:
		c.Data, err = internal.DecodeT[dismodel.ScheduledEventListCall](temp.Data)
	case DiscordScheduledEventCreate:
		c.Data, err = internal.DecodeT[dismodel.ScheduledEventCreateCall](temp.Data)
	case DiscordScheduledEventGet:
		c.Data, err = internal.DecodeT[dismodel.ScheduledEventGetCall](temp.Data)
	case DiscordScheduledEventUpdate:
		c.Data, err = internal.DecodeT[dismodel.ScheduledEventUpdateCall](temp.Data)
	case DiscordScheduledEventDelete:
		c.Data, err = internal.DecodeT[dismodel.ScheduledEventDeleteCall](temp.Data)
	case DiscordScheduledEventUserList:
		c.Data, err = internal.DecodeT[dismodel.ScheduledEventUserListCall](temp.Data)
	case DiscordStageInstanceCreate:
		c.Data, err = internal.DecodeT[dismodel.StageInstanceCreateCall](temp.Data)
	case DiscordStageInstanceGet:
		c.Data, err = internal.DecodeT[dismodel.StageInstanceGetCall](temp.Data)
	case DiscordStageInstanceUpdate:
		c.Data, err = internal.DecodeT[dismodel.StageInstanceUpdateCall](temp.Data)
	case DiscordStageInstanceDelete:
		c.Data, err = internal.DecodeT[dismodel.StageInstanceDeleteCall](temp.Data)
	case DiscordStickerList:
		c.Data, err = internal.DecodeT[dismodel.StickerListCall](temp.Data)
	case DiscordStickerGet:
		c.Data, err = internal.DecodeT[dismodel.StickerGetCall](temp.Data)
	case DiscordStickerCreate:
		c.Data, err = internal.DecodeT[dismodel.StickerCreateCall](temp.Data)
	case DiscordStickerUpdate:
		c.Data, err = internal.DecodeT[dismodel.StickerUpdateCall](temp.Data)
	case DiscordStickerDelete:
		c.Data, err = internal.DecodeT[dismodel.StickerDeleteCall](temp.Data)
	case DiscordUserGet:
		c.Data, err = internal.DecodeT[dismodel.UserGetCall](temp.Data)
	case DiscordWebhookGet:
		c.Data, err = internal.DecodeT[dismodel.WebhookGetCall](temp.Data)
	case DiscordWebhookListForChannel:
		c.Data, err = internal.DecodeT[dismodel.WebhookListForChannelCall](temp.Data)
	case DiscordWebhookListForGuild:
		c.Data, err = internal.DecodeT[dismodel.WebhookListForGuildCall](temp.Data)
	case DiscordWebhookCreate:
		c.Data, err = internal.DecodeT[dismodel.WebhookCreateCall](temp.Data)
	case DiscordWebhookUpdate:
		c.Data, err = internal.DecodeT[dismodel.WebhookUpdateCall](temp.Data)
	case DiscordWebhookDelete:
		c.Data, err = internal.DecodeT[dismodel.WebhookDeleteCall](temp.Data)
	case DiscordWebhookGetWithToken:
		c.Data, err = internal.DecodeT[dismodel.WebhookGetWithTokenCall](temp.Data)
	case DiscordWebhookUpdateWithToken:
		c.Data, err = internal.DecodeT[dismodel.WebhookUpdateWithTokenCall](temp.Data)
	case DiscordWebhookDeleteWithToken:
		c.Data, err = internal.DecodeT[dismodel.WebhookDeleteWithTokenCall](temp.Data)
	case DiscordWebhookExecute:
		c.Data, err = internal.DecodeT[dismodel.WebhookExecuteCall](temp.Data)
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
