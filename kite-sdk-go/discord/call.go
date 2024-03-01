package discord

import (
	"github.com/merlinfuchs/dismod/distype"
	"github.com/merlinfuchs/kite/kite-sdk-go/call"
	"github.com/merlinfuchs/kite/kite-sdk-go/internal"
)

func BanList(opts ...call.CallOption) (distype.BanListResponse, error) {
	return internal.CallWithResponse[distype.BanListResponse](
		call.DiscordBanList, opts,
		distype.BanListRequest{},
	)
}

func BanGet(userID distype.Snowflake, opts ...call.CallOption) (distype.BanGetResponse, error) {
	return internal.CallWithResponse[distype.BanGetResponse](
		call.DiscordBanGet, opts,
		distype.BanGetRequest{
			UserID: userID,
		},
	)
}

func BanCreate(guildID, userID distype.Snowflake, reason string, deleteMessageSeconds int, opts ...call.CallOption) (distype.BanCreateResponse, error) {
	opts = append(opts, call.WithReason(reason))

	var deleteSeconds *int
	if deleteMessageSeconds > 0 {
		deleteSeconds = &deleteMessageSeconds
	}

	return internal.CallWithResponse[distype.BanCreateResponse](
		call.DiscordBanCreate, opts,
		distype.BanCreateRequest{
			GuildID:              guildID,
			UserID:               userID,
			DeleteMessageSeconds: deleteSeconds,
		},
	)
}

func BanRemove(guildID, userID distype.Snowflake, opts ...call.CallOption) (distype.BanRemoveResponse, error) {
	return internal.CallWithResponse[distype.BanRemoveResponse](
		call.DiscordBanRemove, opts,
		distype.BanRemoveRequest{
			GuildID: guildID,
			UserID:  userID,
		},
	)
}

func ChannelGet(channelID distype.Snowflake) (distype.Channel, error) {
	return internal.CallWithResponse[distype.Channel](
		call.DiscordChannelGet, nil,
		distype.ChannelGetRequest{
			ChannelID: channelID,
		},
	)
}

func ChannelList(guildID distype.Snowflake, opts ...call.CallOption) (distype.GuildChannelListResponse, error) {
	return internal.CallWithResponse[distype.GuildChannelListResponse](
		call.DiscordChannelList, opts,
		distype.GuildChannelListRequest{
			GuildID: guildID,
		},
	)
}

func ChannelCreate(args distype.GuildChannelCreateRequest, opts ...call.CallOption) (distype.Channel, error) {
	return internal.CallWithResponse[distype.GuildChannelCreateResponse](call.DiscordChannelCreate, opts, args)
}

func ChannelUpdate(args distype.ChannelModifyRequest, opts ...call.CallOption) (distype.Channel, error) {
	return internal.CallWithResponse[distype.ChannelModifyResponse](call.DiscordChannelUpdate, opts, args)
}

func ChannelUpdatePositions(args distype.GuildChannelModifyPositionsRequest, opts ...call.CallOption) error {
	return internal.CallWithoutResponse(call.DiscordChannelUpdatePositions, opts, args)
}

func ChannelDelete(channelID distype.Snowflake, opts ...call.CallOption) (distype.ChannelDeleteResponse, error) {
	return internal.CallWithResponse[distype.ChannelDeleteResponse](
		call.DiscordChannelDelete, opts,
		distype.ChannelDeleteRequest{
			ChannelID: channelID,
		},
	)
}

func ChannelUpdatePermissions(args distype.ChannelEditPermissionsRequest, opts ...call.CallOption) error {
	return internal.CallWithoutResponse(call.DiscordChannelUpdatePermissions, opts, args)
}

func ChannelDeletePermissions(args distype.ChannelDeletePermissionsRequest, opts ...call.CallOption) (distype.ChannelDeletePermissionsResponse, error) {
	return internal.CallWithResponse[distype.ChannelDeletePermissionsResponse](call.DiscordChannelDeletePermissions, opts, args)
}

func ThreadStartFromMessage(args distype.ThreadStartFromMessageRequest, opts ...call.CallOption) (distype.Channel, error) {
	return internal.CallWithResponse[distype.ThreadStartFromMessageResponse](call.DiscordThreadStartFromMessage, opts, args)
}

func ThreadStart(args distype.ThreadStartWithoutMessageRequest, opts ...call.CallOption) (distype.Channel, error) {
	return internal.CallWithResponse[distype.ThreadStartWithoutMessageResponse](call.DiscordThreadStart, opts, args)
}

func ThreadStartInForum(args distype.ThreadStartInForumRequest, opts ...call.CallOption) (distype.Channel, error) {
	return internal.CallWithResponse[distype.ThreadStartInForumResponse](call.DiscordThreadStartInForum, opts, args)
}

func ThreadJoin(threadID distype.Snowflake, opts ...call.CallOption) (distype.ThreadJoinResponse, error) {
	return internal.CallWithResponse[distype.ThreadJoinResponse](
		call.DiscordThreadJoin, opts,
		distype.ThreadJoinRequest{
			ChannelID: threadID,
		},
	)
}

func ThreadMemberAdd(threadId, userID distype.Snowflake, opts ...call.CallOption) (distype.ThreadMemberAddResponse, error) {
	return internal.CallWithResponse[distype.ThreadMemberAddResponse](
		call.DiscordThreadMemberAdd, opts,
		distype.ThreadMemberAddRequest{
			ChannelID: threadId,
			UserID:    userID,
		},
	)
}

func ThreadLeave(threadID distype.Snowflake, opts ...call.CallOption) (distype.ThreadLeaveResponse, error) {
	return internal.CallWithResponse[distype.ThreadLeaveResponse](
		call.DiscordThreadLeave, opts,
		distype.ThreadLeaveRequest{
			ChannelID: threadID,
		},
	)
}

func ThreadMemberRemove(threadID, userID distype.Snowflake, opts ...call.CallOption) (distype.ThreadMemberRemoveResponse, error) {
	return internal.CallWithResponse[distype.ThreadMemberRemoveResponse](
		call.DiscordThreadMemberRemove, opts,
		distype.ThreadMemberRemoveRequest{
			ChannelID: threadID,
			UserID:    userID,
		},
	)
}

func ThreadMemberGet(threadID, userID distype.Snowflake, opts ...call.CallOption) (distype.ThreadMemberGetResponse, error) {
	return internal.CallWithResponse[distype.ThreadMemberGetResponse](
		call.DiscordThreadMemberGet, opts,
		distype.ThreadMemberGetRequest{
			ChannelID: threadID,
			UserID:    userID,
		},
	)
}

func ThreadMemberList(args distype.ThreadMemberListRequest, opts ...call.CallOption) (distype.ThreadMemberListResponse, error) {
	return internal.CallWithResponse[distype.ThreadMemberListResponse](
		call.DiscordThreadMemberList, opts,
		args,
	)
}

func ThreadListPublicArchived(args distype.ThreadListPublicArchivedRequest, opts ...call.CallOption) (distype.ThreadListPublicArchivedResponse, error) {
	return internal.CallWithResponse[distype.ThreadListPublicArchivedResponse](
		call.DiscordThreadListPublicArchived, opts,
		args,
	)
}

func ThreadListPrivateArchived(args distype.ThreadListPrivateArchivedRequest, opts ...call.CallOption) (distype.ThreadListPrivateArchivedResponse, error) {
	return internal.CallWithResponse[distype.ThreadListPrivateArchivedResponse](
		call.DiscordThreadListPrivateArchived, opts,
		args,
	)
}

func ThreadListJoinedPrivateArchived(args distype.ThreadListJoinedPrivateArchivedRequest, opts ...call.CallOption) (distype.ThreadListJoinedPrivateArchivedResponse, error) {
	return internal.CallWithResponse[distype.ThreadListJoinedPrivateArchivedResponse](
		call.DiscordThreadListJoinedPrivateArchived, opts,
		args,
	)
}

func ThreadListActive(guildID distype.Snowflake, opts ...call.CallOption) (distype.GuildThreadListActiveResponse, error) {
	return internal.CallWithResponse[distype.GuildThreadListActiveResponse](
		call.DiscordThreadListActive, opts,
		distype.GuildThreadListActiveRequest{
			GuildID: guildID,
		},
	)
}

func EmojiList(opts ...call.CallOption) ([]distype.Emoji, error) {
	return internal.CallWithResponse[distype.GuildEmojiListResponse](
		call.DiscordEmojiList, opts,
		distype.GuildEmojiListRequest{},
	)
}

func EmojiGet(guildID, emojiID distype.Snowflake, opts ...call.CallOption) (distype.EmojiGetResponse, error) {
	return internal.CallWithResponse[distype.EmojiGetResponse](
		call.DiscordEmojiGet, opts,
		distype.EmojiGetRequest{
			GuildID: guildID,
			EmojiID: emojiID,
		},
	)
}

func EmojiCreate(args distype.EmojiCreateRequest, opts ...call.CallOption) (distype.Emoji, error) {
	return internal.CallWithResponse[distype.EmojiCreateResponse](call.DiscordEmojiCreate, opts, args)
}

func EmojiUpdate(args distype.EmojiModifyRequest, opts ...call.CallOption) (distype.Emoji, error) {
	return internal.CallWithResponse[distype.EmojiModifyResponse](call.DiscordEmojiUpdate, opts, args)
}

func EmojiDelete(guildID, emojiID distype.Snowflake, opts ...call.CallOption) error {
	return internal.CallWithoutResponse(
		call.DiscordEmojiDelete, opts,
		distype.EmojiDeleteRequest{
			GuildID: guildID,
			EmojiID: emojiID,
		},
	)
}

func GuildGet(guildID distype.Snowflake, opts ...call.CallOption) (distype.Guild, error) {
	return internal.CallWithResponse[distype.GuildGetResponse](
		call.DiscordGuildGet, opts,
		distype.GuildGetRequest{
			GuildID: guildID,
		},
	)
}

func GuildUpdate(args distype.GuildUpdateRequest, opts ...call.CallOption) (distype.GuildUpdateResponse, error) {
	return internal.CallWithResponse[distype.GuildUpdateResponse](call.DiscordGuildUpdate, opts, args)
}

func InteractionResponseCreate(args distype.InteractionResponseCreateRequest, opts ...call.CallOption) (distype.InteractionResponseCreateResponse, error) {
	return internal.CallWithResponse[distype.InteractionResponseCreateResponse](call.DiscordInteractionResponseCreate, opts, args)
}

func InteractionResponseUpdate(args distype.InteractionResponseEditRequest, opts ...call.CallOption) (distype.Message, error) {
	return internal.CallWithResponse[distype.InteractionResponseEditResponse](call.DiscordInteractionResponseUpdate, opts, args)
}

func InteractionResponseDelete(applicationID distype.Snowflake, interactionToken string, opts ...call.CallOption) error {
	return internal.CallWithoutResponse(
		call.DiscordInteractionResponseDelete, opts,
		distype.InteractionResponseDeleteRequest{
			ApplicationID:    applicationID,
			InteractionToken: interactionToken,
		},
	)
}

func InteractionFollowupCreate(args distype.InteractionFollowupCreateRequest, opts ...call.CallOption) (distype.InteractionFollowupCreateResponse, error) {
	return internal.CallWithResponse[distype.InteractionFollowupCreateResponse](call.DiscordInteractionFollowupCreate, opts, args)
}

func InteractionFollowupUpdate(args distype.InteractionFollowupEditRequest, opts ...call.CallOption) (distype.Message, error) {
	return internal.CallWithResponse[distype.InteractionFollowupEditResponse](call.DiscordInteractionFollowupUpdate, opts, args)
}

func InteractionFollowupDelete(applicationID distype.Snowflake, interactionToken string, messageID distype.Snowflake, opts ...call.CallOption) error {
	return internal.CallWithoutResponse(
		call.DiscordInteractionFollowupDelete, opts,
		distype.InteractionFollowupDeleteRequest{
			ApplicationID:    applicationID,
			InteractionToken: interactionToken,
			MessageID:        messageID,
		},
	)
}

func InteractionFollowupGet(applicationID distype.Snowflake, interactionToken string, messageID distype.Snowflake, opts ...call.CallOption) (distype.InteractionFollowupGetResponse, error) {
	return internal.CallWithResponse[distype.InteractionFollowupGetResponse](
		call.DiscordInteractionFollowupGet, opts,
		distype.InteractionFollowupGetRequest{
			ApplicationID:    applicationID,
			InteractionToken: interactionToken,
			MessageID:        messageID,
		},
	)
}

func InviteListForChannel(channelID distype.Snowflake, opts ...call.CallOption) ([]distype.Invite, error) {
	return internal.CallWithResponse[distype.ChannelInviteListResponse](
		call.DiscordInviteListForChannel, opts,
		distype.ChannelInviteListRequest{
			ChannelID: channelID,
		},
	)
}

func InviteListForGuild(guildID distype.Snowflake, opts ...call.CallOption) ([]distype.Invite, error) {
	return internal.CallWithResponse[distype.GuildInviteListResponse](
		call.DiscordInviteListForGuild, opts,
		distype.GuildInviteListRequest{
			GuildID: guildID,
		},
	)
}

func InviteCreate(args distype.ChannelInviteCreateRequest, opts ...call.CallOption) (distype.Invite, error) {
	return internal.CallWithResponse[distype.ChannelInviteCreateResponse](call.DiscordInviteCreate, opts, args)
}

func InviteGet(args distype.InviteGetRequest, opts ...call.CallOption) (distype.InviteGetResponse, error) {
	return internal.CallWithResponse[distype.InviteGetResponse](call.DiscordInviteGet, opts, args)
}

func InviteDelete(code string, opts ...call.CallOption) (distype.InviteDeleteResponse, error) {
	return internal.CallWithResponse[distype.InviteDeleteResponse](
		call.DiscordInviteDelete, opts,
		distype.InviteDeleteRequest{
			Code: code,
		},
	)
}

func MemberGet(guildID, userID distype.Snowflake, opts ...call.CallOption) (distype.MemberGetResponse, error) {
	return internal.CallWithResponse[distype.MemberGetResponse](
		call.DiscordMemberGet, opts,
		distype.MemberGetRequest{
			GuildID: guildID,
			UserID:  userID,
		},
	)
}

func MemberList(guildID distype.Snowflake, opts ...call.CallOption) ([]distype.Member, error) {
	return internal.CallWithResponse[distype.GuildMemberListResponse](
		call.DiscordMemberList, opts,
		distype.GuildMemberListRequest{
			GuildID: guildID,
		},
	)
}

func MemberSearch(args distype.GuildMemberSearchRequest, opts ...call.CallOption) ([]distype.Member, error) {
	return internal.CallWithResponse[distype.GuildMemberSearchResponse](
		call.DiscordMemberSearch, opts,
		args,
	)
}

func MemberUpdate(args distype.MemberModifyRequest, opts ...call.CallOption) (distype.Member, error) {
	return internal.CallWithResponse[distype.MemberModifyResponse](call.DiscordMemberUpdate, opts, args)
}

func MemberUpdateOwn(args distype.MemberModifyCurrentRequest, opts ...call.CallOption) error {
	return internal.CallWithoutResponse(call.DiscordMemberUpdateOwn, opts, args)
}

func MemberRoleAdd(guildID, userID, roleID distype.Snowflake, opts ...call.CallOption) error {
	return internal.CallWithoutResponse(
		call.DiscordMemberRoleAdd, opts,
		distype.MemberRoleAddRequest{
			GuildID: guildID,
			UserID:  userID,
			RoleID:  roleID,
		},
	)
}

func MemberRoleRemove(guildID, userID, roleID distype.Snowflake, opts ...call.CallOption) error {
	return internal.CallWithoutResponse(
		call.DiscordMemberRoleRemove, opts,
		distype.MemberRoleRemoveRequest{
			GuildID: guildID,
			UserID:  userID,
			RoleID:  roleID,
		},
	)
}

func MemberRemove(guildID, userID distype.Snowflake, opts ...call.CallOption) error {
	return internal.CallWithoutResponse(
		call.DiscordMemberRemove, opts,
		distype.MemberRemoveRequest{
			UserID: userID,
		},
	)
}

func MemberPruneCount(args distype.MemberPruneCountRequest, opts ...call.CallOption) (distype.MemberPruneCountResponse, error) {
	return internal.CallWithResponse[distype.MemberPruneCountResponse](call.DiscordMemberPruneCount, opts, args)
}

func MemberPruneBegin(args distype.MemberPruneRequest, opts ...call.CallOption) (distype.MemberPruneResponse, error) {
	return internal.CallWithResponse[distype.MemberPruneResponse](call.DiscordMemberPruneBegin, opts, args)
}

func MessageList(args distype.ChannelMessageListRequest, opts ...call.CallOption) ([]distype.Message, error) {
	return internal.CallWithResponse[distype.ChannelMessageListResponse](call.DiscordMessageList, opts, args)
}

func MessageGet(channelID, messageID distype.Snowflake, opts ...call.CallOption) (distype.MessageGetResponse, error) {
	return internal.CallWithResponse[distype.MessageGetResponse](
		call.DiscordMessageGet, opts,
		distype.MessageGetRequest{
			ChannelID: channelID,
			MessageID: messageID,
		},
	)
}

func MessageCreate(args distype.MessageCreateRequest, opts ...call.CallOption) (distype.MessageCreateResponse, error) {
	return internal.CallWithResponse[distype.MessageCreateResponse](call.DiscordMessageCreate, opts, args)
}

func MessageUpdate(args distype.MessageEditRequest, opts ...call.CallOption) (distype.Message, error) {
	return internal.CallWithResponse[distype.MessageEditResponse](call.DiscordMessageUpdate, opts, args)
}

func MessageDelete(channelID, messageID distype.Snowflake, opts ...call.CallOption) error {
	return internal.CallWithoutResponse(
		call.DiscordMessageDelete, opts,
		distype.MessageDeleteRequest{
			ChannelID: channelID,
			MessageID: messageID,
		},
	)
}

func MessageDeleteBulk(channelID distype.Snowflake, messages []distype.Snowflake, opts ...call.CallOption) error {
	return internal.CallWithoutResponse(
		call.DiscordMessageDeleteBulk, opts,
		distype.MessageBulkDeleteRequest{
			ChannelID: channelID,
			Messages:  messages,
		},
	)
}

func MessageReactionCreate(channelID, messageID distype.Snowflake, emoji string, opts ...call.CallOption) error {
	return internal.CallWithoutResponse(
		call.DiscordMessageReactionCreate, opts,
		distype.MessageReactionCreateRequest{
			ChannelID: channelID,
			MessageID: messageID,
			Emoji:     emoji,
		},
	)
}

func MessageReactionDeleteOwn(channelID, messageID distype.Snowflake, emoji string, opts ...call.CallOption) error {
	return internal.CallWithoutResponse(
		call.DiscordMessageReactionDeleteOwn, opts,
		distype.MessageReactionDeleteOwnRequest{
			ChannelID: channelID,
			MessageID: messageID,
			Emoji:     emoji,
		},
	)
}

func MessageReactionDeleteUser(channelID, messageID distype.Snowflake, emoji string, userID distype.Snowflake, opts ...call.CallOption) error {
	return internal.CallWithoutResponse(
		call.DiscordMessageReactionDeleteUser, opts,
		distype.MessageReactionDeleteRequest{
			ChannelID: channelID,
			MessageID: messageID,
			Emoji:     emoji,
			UserID:    userID,
		},
	)
}

func MessageReactionList(args distype.MessageReactionListRequest, opts ...call.CallOption) (distype.MessageReactionListResponse, error) {
	return internal.CallWithResponse[distype.MessageReactionListResponse](call.DiscordMessageReactionList, opts, args)
}

func MessageReactionDeleteAll(channelID, messageID distype.Snowflake, opts ...call.CallOption) error {
	return internal.CallWithoutResponse(
		call.DiscordMessageReactionDeleteAll, opts,
		distype.MessageReactionDeleteAllRequest{
			ChannelID: channelID,
			MessageID: messageID,
		},
	)
}

func MessageReactionDeleteEmoji(channelID, messageID distype.Snowflake, emoji string, opts ...call.CallOption) error {
	return internal.CallWithoutResponse(
		call.DiscordMessageReactionDeleteEmoji, opts,
		distype.MessageReactionDeleteEmojiRequest{
			ChannelID: channelID,
			MessageID: messageID,
			Emoji:     emoji,
		},
	)
}

func MessageGetPinned(channelID distype.Snowflake, opts ...call.CallOption) ([]distype.Message, error) {
	return internal.CallWithResponse[distype.ChannelPinnedMessageListResponse](
		call.DiscordMessageGetPinned, opts,
		distype.ChannelPinnedMessageListRequest{
			ChannelID: channelID,
		},
	)
}

func MessagePin(channelID, messageID distype.Snowflake, opts ...call.CallOption) error {
	return internal.CallWithoutResponse(
		call.DiscordMessagePin, opts,
		distype.MessagePinRequest{
			ChannelID: channelID,
			MessageID: messageID,
		},
	)
}

func MessageUnpin(channelID, messageID distype.Snowflake, opts ...call.CallOption) error {
	return internal.CallWithoutResponse(
		call.DiscordMessageUnpin, opts,
		distype.MessageUnpinRequest{
			ChannelID: channelID,
			MessageID: messageID,
		},
	)
}

func RoleList(guildID distype.Snowflake, opts ...call.CallOption) ([]distype.Role, error) {
	return internal.CallWithResponse[distype.GuildRoleListResponse](
		call.DiscordRoleList, opts,
		distype.GuildRoleListRequest{
			GuildID: guildID,
		},
	)
}

func RoleCreate(args distype.RoleCreateRequest, opts ...call.CallOption) (distype.Role, error) {
	return internal.CallWithResponse[distype.RoleCreateResponse](call.DiscordRoleCreate, opts, args)
}

func RoleUpdate(args distype.RoleModifyRequest, opts ...call.CallOption) (distype.Role, error) {
	return internal.CallWithResponse[distype.RoleModifyResponse](call.DiscordRoleUpdate, opts, args)
}

func RoleUpdatePositions(args distype.RolePositionsModifyRequest, opts ...call.CallOption) error {
	return internal.CallWithoutResponse(call.DiscordRoleUpdatePositions, opts, args)
}

func RoleDelete(guildID, roleID distype.Snowflake, opts ...call.CallOption) (distype.RoleDeleteResponse, error) {
	return internal.CallWithResponse[distype.RoleDeleteResponse](
		call.DiscordRoleDelete, opts,
		distype.RoleDeleteRequest{
			GuildID: guildID,
			RoleID:  roleID,
		},
	)
}

func ScheduledEventList(guildID distype.Snowflake, withUserCount bool, opts ...call.CallOption) ([]distype.ScheduledEvent, error) {
	var withCount *bool
	if withUserCount {
		withCount = &withUserCount
	}

	return internal.CallWithResponse[distype.GuildScheduledEventListResponse](
		call.DiscordScheduledEventList, opts,
		distype.GuildScheduledEventListRequest{
			GuildID:       guildID,
			WithUserCount: withCount,
		},
	)
}

func ScheduledEventGet(guildID, eventID distype.Snowflake, opts ...call.CallOption) (distype.ScheduledEvent, error) {
	return internal.CallWithResponse[distype.ScheduledEventGetResponse](
		call.DiscordScheduledEventGet, opts,
		distype.ScheduledEventGetRequest{
			GuildID:               guildID,
			GuildScheduledEventID: eventID,
		},
	)
}

func ScheduledEventCreate(args distype.ScheduledEventCreateRequest, opts ...call.CallOption) (distype.ScheduledEventCreateResponse, error) {
	return internal.CallWithResponse[distype.ScheduledEventCreateResponse](call.DiscordScheduledEventCreate, opts, args)
}

func ScheduledEventUpdate(args distype.ScheduledEventModifyRequest, opts ...call.CallOption) (distype.ScheduledEvent, error) {
	return internal.CallWithResponse[distype.ScheduledEventModifyResponse](call.DiscordScheduledEventUpdate, opts, args)
}

func ScheduledEventDelete(guildID, eventID distype.Snowflake, opts ...call.CallOption) error {
	return internal.CallWithoutResponse(
		call.DiscordScheduledEventDelete, opts,
		distype.ScheduledEventDeleteRequest{
			GuildID:               guildID,
			GuildScheduledEventID: eventID,
		},
	)
}

func ScheduledEventUserList(args distype.ScheduledEventUserListRequest, opts ...call.CallOption) (distype.ScheduledEventUserListResponse, error) {
	return internal.CallWithResponse[distype.ScheduledEventUserListResponse](call.DiscordScheduledEventUserList, opts, args)
}

func StageInstanceCreate(args distype.StageInstanceCreateRequest, opts ...call.CallOption) (distype.StageInstanceCreateResponse, error) {
	return internal.CallWithResponse[distype.StageInstanceCreateResponse](call.DiscordStageInstanceCreate, opts, args)
}

func StageInstanceGet(channelID distype.Snowflake, opts ...call.CallOption) (distype.StageInstanceGetResponse, error) {
	return internal.CallWithResponse[distype.StageInstanceGetResponse](
		call.DiscordStageInstanceGet, opts,
		distype.StageInstanceGetRequest{
			ChannelID: channelID,
		},
	)
}

func StageInstanceUpdate(args distype.StageInstanceModifyRequest, opts ...call.CallOption) (distype.StageInstance, error) {
	return internal.CallWithResponse[distype.StageInstanceModifyResponse](call.DiscordStageInstanceUpdate, opts, args)
}

func StageInstanceDelete(channelID distype.Snowflake, opts ...call.CallOption) error {
	return internal.CallWithoutResponse(
		call.DiscordStageInstanceDelete, opts,
		distype.StageInstanceDeleteRequest{
			ChannelID: channelID,
		},
	)
}

func StickerList(guildID distype.Snowflake, opts ...call.CallOption) ([]distype.Sticker, error) {
	return internal.CallWithResponse[distype.GuildStickerListResponse](
		call.DiscordStickerList, opts,
		distype.GuildStickerListRequest{
			GuildID: guildID,
		},
	)
}

func StickerGet(guildID, stickerID distype.Snowflake, opts ...call.CallOption) (distype.StickerGetResponse, error) {
	return internal.CallWithResponse[distype.StickerGetResponse](
		call.DiscordStickerGet, opts,
		distype.StickerGetRequest{
			GuildID:   guildID,
			StickerID: stickerID,
		},
	)
}

func StickerCreate(args distype.StickerCreateRequest, opts ...call.CallOption) (distype.StickerCreateResponse, error) {
	return internal.CallWithResponse[distype.StickerCreateResponse](call.DiscordStickerCreate, opts, args)
}

func StickerUpdate(args distype.StickerModifyRequest, opts ...call.CallOption) (distype.Sticker, error) {
	return internal.CallWithResponse[distype.StickerModifyResponse](call.DiscordStickerUpdate, opts, args)
}

func StickerDelete(guildID, stickerID distype.Snowflake, opts ...call.CallOption) error {
	return internal.CallWithoutResponse(
		call.DiscordStickerDelete, opts,
		distype.StickerDeleteRequest{
			GuildID:   guildID,
			StickerID: stickerID,
		},
	)
}

func UserGet(userID distype.Snowflake, opts ...call.CallOption) (distype.UserGetResponse, error) {
	return internal.CallWithResponse[distype.UserGetResponse](
		call.DiscordUserGet, opts,
		distype.UserGetRequest{
			UserID: userID,
		},
	)
}

func WebhookGet(webhookID distype.Snowflake, opts ...call.CallOption) (distype.WebhookGetResponse, error) {
	return internal.CallWithResponse[distype.WebhookGetResponse](
		call.DiscordWebhookGet, opts,
		distype.WebhookGetRequest{
			WebhookID: webhookID,
		},
	)
}

func WebhookListForChannel(channelID distype.Snowflake, opts ...call.CallOption) ([]distype.Webhook, error) {
	return internal.CallWithResponse[distype.ChannelWebhookListResponse](
		call.DiscordWebhookListForChannel, opts,
		distype.ChannelWebhookListRequest{
			ChannelID: channelID,
		},
	)
}

func WebhookListForGuild(guildID distype.Snowflake, opts ...call.CallOption) ([]distype.Webhook, error) {
	return internal.CallWithResponse[distype.GuildWebhookListResponse](
		call.DiscordWebhookListForGuild, opts,
		distype.GuildWebhookListRequest{
			GuildID: guildID,
		},
	)
}

func WebhookCreate(args distype.WebhookCreateRequest, opts ...call.CallOption) (distype.WebhookCreateResponse, error) {
	return internal.CallWithResponse[distype.WebhookCreateResponse](call.DiscordWebhookCreate, opts, args)
}

func WebhookUpdate(args distype.WebhookModifyRequest, opts ...call.CallOption) (distype.Webhook, error) {
	return internal.CallWithResponse[distype.WebhookModifyResponse](call.DiscordWebhookUpdate, opts, args)
}

func WebhookDelete(webhookID distype.Snowflake, opts ...call.CallOption) error {
	return internal.CallWithoutResponse(
		call.DiscordWebhookDelete, opts,
		distype.WebhookDeleteRequest{
			WebhookID: webhookID,
		},
	)
}

func WebhookGetWithToken(webhookID distype.Snowflake, webhookToken string, opts ...call.CallOption) (distype.WebhookGetWithTokenResponse, error) {
	return internal.CallWithResponse[distype.WebhookGetWithTokenResponse](
		call.DiscordWebhookGetWithToken, opts,
		distype.WebhookGetWithTokenRequest{
			WebhookID:    webhookID,
			WebhookToken: webhookToken,
		},
	)
}

func WebhookUpdateWithToken(args distype.WebhookModifyWithTokenRequest, opts ...call.CallOption) (distype.Webhook, error) {
	return internal.CallWithResponse[distype.WebhookModifyWithTokenResponse](call.DiscordWebhookUpdateWithToken, opts, args)
}

func WebhookDeleteWithToken(webhookID distype.Snowflake, webhookToken string, opts ...call.CallOption) error {
	return internal.CallWithoutResponse(
		call.DiscordWebhookDeleteWithToken, opts,
		distype.WebhookDeleteWithTokenRequest{
			WebhookID:    webhookID,
			WebhookToken: webhookToken,
		},
	)
}

func WebhookExecute(args distype.WebhookExecuteRequest, opts ...call.CallOption) (distype.WebhookExecuteResponse, error) {
	return internal.CallWithResponse[distype.WebhookExecuteResponse](call.DiscordWebhookExecute, opts, args)
}
