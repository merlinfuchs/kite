// Code generated by tygo. DO NOT EDIT.
type Nullable<T> = T | null

//////////
// source: call.go

export type CallType = string;
export const Sleep: CallType = "SLEEP";
export const KVKeyGet: CallType = "KV_KEY_GET";
export const KVKeySet: CallType = "KV_KEY_SET";
export const KVKeyDelete: CallType = "KV_KEY_DELETE";
export const KVKeyIncrease: CallType = "KV_KEY_INCREASE";
export const DiscordBanList: CallType = "DISCORD_BAN_LIST";
export const DiscordBanGet: CallType = "DISCORD_BAN_GET";
export const DiscordBanCreate: CallType = "DISCORD_BAN_CREATE";
export const DiscordBanRemove: CallType = "DISCORD_BAN_REMOVE";
export const DiscordChannelGet: CallType = "DISCORD_CHANNEL_GET";
export const DiscordChannelList: CallType = "DISCORD_CHANNEL_LIST";
export const DiscordChannelCreate: CallType = "DISCORD_CHANNEL_CREATE";
export const DiscordChannelUpdate: CallType = "DISCORD_CHANNEL_UPDATE";
export const DiscordChannelUpdatePositions: CallType = "DISCORD_CHANNEL_UPDATE_POSITIONS";
export const DiscordChannelDelete: CallType = "DISCORD_CHANNEL_DELETE";
export const DiscordChannelUpdatePermissions: CallType = "DISCORD_CHANNEL_UPDATE_PERMISSIONS";
export const DiscordChannelDeletePermissions: CallType = "DISCORD_CHANNEL_DELETE_PERMISSIONS";
export const DiscordThreadStartFromMessage: CallType = "DISCORD_THREAD_START_FROM_MESSAGE";
export const DiscordThreadStart: CallType = "DISCORD_THREAD_START";
export const DiscordThreadStartInForum: CallType = "DISCORD_THREAD_START_IN_FORUM";
export const DiscordThreadJoin: CallType = "DISCORD_THREAD_JOIN";
export const DiscordThreadMemberAdd: CallType = "DISCORD_THREAD_ADD_MEMBER";
export const DiscordThreadLeave: CallType = "DISCORD_THREAD_LEAVE";
export const DiscordThreadMemberRemove: CallType = "DISCORD_THREAD_REMOVE_MEMBER";
export const DiscordThreadMemberGet: CallType = "DISCORD_THREAD_MEMBER_GET";
export const DiscordThreadMemberList: CallType = "DISCORD_THREAD_MEMBER_LIST";
export const DiscordThreadListPublicArchived: CallType = "DISCORD_THREAD_LIST_PUBLIC_ARCHIVED";
export const DiscordThreadListPrivateArchived: CallType = "DISCORD_THREAD_LIST_PRIVATE_ARCHIVED";
export const DiscordThreadListJoinedPrivateArchived: CallType = "DISCORD_THREAD_LIST_JOINED_PRIVATE_ARCHIVED";
export const DiscordThreadListActive: CallType = "DISCORD_THREAD_LIST_ACTIVE";
export const DiscordEmojiList: CallType = "DISCORD_EMOJI_LIST";
export const DiscordEmojiGet: CallType = "DISCORD_EMOJI_GET";
export const DiscordEmojiCreate: CallType = "DISCORD_EMOJI_CREATE";
export const DiscordEmojiUpdate: CallType = "DISCORD_EMOJI_UPDATE";
export const DiscordEmojiDelete: CallType = "DISCORD_EMOJI_DELETE";
export const DiscordGuildGet: CallType = "DISCORD_GUILD_GET";
export const DiscordGuildUpdate: CallType = "DISCORD_GUILD_UPDATE";
export const DiscordInteractionResponseCreate: CallType = "DISCORD_INTERACTION_RESPONSE_CREATE";
export const DiscordInteractionResponseUpdate: CallType = "DISCORD_INTERACTION_RESPONSE_UPDATE";
export const DiscordInteractionResponseDelete: CallType = "DISCORD_INTERACTION_RESPONSE_DELETE";
export const DiscordInteractionResponseGet: CallType = "DISCORD_INTERACTION_RESPONSE_GET";
export const DiscordInteractionFollowupCreate: CallType = "DISCORD_INTERACTION_FOLLOWUP_CREATE";
export const DiscordInteractionFollowupUpdate: CallType = "DISCORD_INTERACTION_FOLLOWUP_UPDATE";
export const DiscordInteractionFollowupDelete: CallType = "DISCORD_INTERACTION_FOLLOWUP_DELETE";
export const DiscordInteractionFollowupGet: CallType = "DISCORD_INTERACTION_FOLLOWUP_GET";
export const DiscordInviteListForChannel: CallType = "DISCORD_INVITE_LIST_FOR_CHANNEL";
export const DiscordInviteListForGuild: CallType = "DISCORD_INVITE_LIST_FOR_GUILD";
export const DiscordInviteCreate: CallType = "DISCORD_INVITE_CREATE";
export const DiscordInviteGet: CallType = "DISCORD_INVITE_GET";
export const DiscordInviteDelete: CallType = "DISCORD_INVITE_DELETE";
export const DiscordMemberGet: CallType = "DISCORD_MEMBER_GET";
export const DiscordMemberList: CallType = "DISCORD_MEMBER_LIST";
export const DiscordMemberSearch: CallType = "DISCORD_MEMBER_SEARCH";
export const DiscordMemberUpdate: CallType = "DISCORD_MEMBER_UPDATE";
export const DiscordMemberUpdateOwn: CallType = "DISCORD_MEMBER_UPDATE_OWN";
export const DiscordMemberRoleAdd: CallType = "DISCORD_MEMBER_ADD_ROLE";
export const DiscordMemberRoleRemove: CallType = "DISCORD_MEMBER_REMOVE_ROLE";
export const DiscordMemberRemove: CallType = "DISCORD_MEMBER_REMOVE";
export const DiscordMemberPruneCount: CallType = "DISCORD_MEMBER_PRUNE_COUNT";
export const DiscordMemberPruneBegin: CallType = "DISCORD_MEMBER_PRUNE_BEGIN";
export const DiscordMessageList: CallType = "DISCORD_MESSAGE_LIST";
export const DiscordMessageGet: CallType = "DISCORD_MESSAGE_GET";
export const DiscordMessageCreate: CallType = "DISCORD_MESSAGE_CREATE";
export const DiscordMessageUpdate: CallType = "DISCORD_MESSAGE_UPDATE";
export const DiscordMessageDelete: CallType = "DISCORD_MESSAGE_DELETE";
export const DiscordMessageDeleteBulk: CallType = "DISCORD_MESSAGE_DELETE_BULK";
export const DiscordMessageReactionCreate: CallType = "DISCORD_MESSAGE_REACTION_CREATE";
export const DiscordMessageReactionDeleteOwn: CallType = "DISCORD_MESSAGE_REACTION_DELETE_OWN";
export const DiscordMessageReactionDeleteUser: CallType = "DISCORD_MESSAGE_REACTION_DELETE_USER";
export const DiscordMessageReactionList: CallType = "DISCORD_MESSAGE_REACTION_LIST";
export const DiscordMessageReactionDeleteAll: CallType = "DISCORD_MESSAGE_REACTION_DELETE_ALL";
export const DiscordMessageReactionDeleteEmoji: CallType = "DISCORD_MESSAGE_REACTION_DELETE_EMOJI";
export const DiscordMessageGetPinned: CallType = "DISCORD_MESSAGE_GET_PINNED";
export const DiscordMessagePin: CallType = "DISCORD_MESSAGE_PIN";
export const DiscordMessageUnpin: CallType = "DISCORD_MESSAGE_UNPIN";
export const DiscordRoleList: CallType = "DISCORD_ROLE_LIST";
export const DiscordRoleCreate: CallType = "DISCORD_ROLE_CREATE";
export const DiscordRoleUpdate: CallType = "DISCORD_ROLE_UPDATE";
export const DiscordRoleUpdatePositions: CallType = "DISCORD_ROLE_UPDATE_POSITIONS";
export const DiscordRoleDelete: CallType = "DISCORD_ROLE_DELETE";
export const DiscordScheduledEventList: CallType = "DISCORD_SCHEDULED_EVENT_LIST";
export const DiscordScheduledEventCreate: CallType = "DISCORD_SCHEDULED_EVENT_CREATE";
export const DiscordScheduledEventGet: CallType = "DISCORD_SCHEDULED_EVENT_GET";
export const DiscordScheduledEventUpdate: CallType = "DISCORD_SCHEDULED_EVENT_UPDATE";
export const DiscordScheduledEventDelete: CallType = "DISCORD_SCHEDULED_EVENT_DELETE";
export const DiscordScheduledEventUserList: CallType = "DISCORD_SCHEDULED_EVENT_USER_LIST";
export const DiscordStageInstanceCreate: CallType = "DISCORD_STAGE_INSTANCE_CREATE";
export const DiscordStageInstanceGet: CallType = "DISCORD_STAGE_INSTANCE_GET";
export const DiscordStageInstanceUpdate: CallType = "DISCORD_STAGE_INSTANCE_UPDATE";
export const DiscordStageInstanceDelete: CallType = "DISCORD_STAGE_INSTANCE_DELETE";
export const DiscordStickerList: CallType = "DISCORD_STICKER_LIST";
export const DiscordStickerGet: CallType = "DISCORD_STICKER_GET";
export const DiscordStickerCreate: CallType = "DISCORD_STICKER_CREATE";
export const DiscordStickerUpdate: CallType = "DISCORD_STICKER_UPDATE";
export const DiscordStickerDelete: CallType = "DISCORD_STICKER_DELETE";
export const DiscordUserGet: CallType = "DISCORD_USER_GET";
export const DiscordWebhookGet: CallType = "DISCORD_WEBHOOK_GET";
export const DiscordWebhookListForChannel: CallType = "DISCORD_WEBHOOK_LIST_FOR_CHANNEL";
export const DiscordWebhookListForGuild: CallType = "DISCORD_WEBHOOK_LIST_FOR_GUILD";
export const DiscordWebhookCreate: CallType = "DISCORD_WEBHOOK_CREATE";
export const DiscordWebhookUpdate: CallType = "DISCORD_WEBHOOK_UPDATE";
export const DiscordWebhookDelete: CallType = "DISCORD_WEBHOOK_DELETE";
export const DiscordWebhookGetWithToken: CallType = "DISCORD_WEBHOOK_GET_WITH_TOKEN";
export const DiscordWebhookUpdateWithToken: CallType = "DISCORD_WEBHOOK_UPDATE_WITH_TOKEN";
export const DiscordWebhookDeleteWithToken: CallType = "DISCORD_WEBHOOK_DELETE_WITH_TOKEN";
export const DiscordWebhookExecute: CallType = "DISCORD_WEBHOOK_EXECUTE";
export interface Call {
  type: CallType;
  config: CallConfig;
  data: any;
}
export interface CallResponse {
  success: boolean;
  error?: any /* fail.HostError */;
  data: any;
}

//////////
// source: option.go

export interface CallConfig {
  reason?: string;
  timeout?: number /* int */;
  wait?: boolean;
}
export type CallOption = any;

//////////
// source: sleep.go

export interface SleepCall {
  duration: number /* int */;
}
export interface SleepResponse {
}
