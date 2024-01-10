package discord

import (
	"github.com/merlinfuchs/kite/go-sdk/internal"
	"github.com/merlinfuchs/kite/go-types/call"
	"github.com/merlinfuchs/kite/go-types/dismodel"
)

func BanList(opts ...call.CallOption) (dismodel.BanListResponse, error) {
	return internal.CallWithResponse[dismodel.BanListResponse](
		call.DiscordBanList, opts,
		dismodel.BanListCall{},
	)
}

func BanGet(userID string, opts ...call.CallOption) (dismodel.BanGetResponse, error) {
	return internal.CallWithResponse[dismodel.BanGetResponse](
		call.DiscordBanGet, opts,
		dismodel.BanGetCall{
			UserID: userID,
		},
	)
}

func BanCreate(userID, reason string, deleteMessageSeconds int, opts ...call.CallOption) (dismodel.BanCreateResponse, error) {
	opts = append(opts, call.WithReason(reason))

	return internal.CallWithResponse[dismodel.BanCreateResponse](
		call.DiscordBanCreate, opts,
		dismodel.BanCreateCall{
			UserID:               userID,
			DeleteMessageSeconds: deleteMessageSeconds,
		},
	)
}

func BanRemove(userID string, opts ...call.CallOption) (dismodel.BanRemoveResponse, error) {
	return internal.CallWithResponse[dismodel.BanRemoveResponse](
		call.DiscordBanRemove, opts,
		dismodel.BanRemoveCall{
			UserID: userID,
		},
	)
}

func ChannelGet(channelID string) (dismodel.Channel, error) {
	return internal.CallWithResponse[dismodel.Channel](
		call.DiscordChannelGet, nil,
		dismodel.ChannelGetCall{
			ID: channelID,
		},
	)
}

func ChannelList(opts ...call.CallOption) (dismodel.ChannelListResponse, error) {
	return internal.CallWithResponse[dismodel.ChannelListResponse](
		call.DiscordChannelList, opts,
		dismodel.ChannelListCall{},
	)
}

func ChannelCreate(args dismodel.ChannelCreateCall, opts ...call.CallOption) (dismodel.Channel, error) {
	return internal.CallWithResponse[dismodel.ChannelCreateResponse](call.DiscordChannelCreate, opts, args)
}

func ChannelUpdate(args dismodel.ChannelUpdateCall, opts ...call.CallOption) (dismodel.Channel, error) {
	return internal.CallWithResponse[dismodel.ChannelUpdateResponse](call.DiscordChannelUpdate, opts, args)
}

func ChannelUpdatePositions(args dismodel.ChannelUpdatePositionsCall, opts ...call.CallOption) (dismodel.ChannelListResponse, error) {
	return internal.CallWithResponse[dismodel.ChannelListResponse](call.DiscordChannelUpdatePositions, opts, args)
}

func ChannelDelete(channelID string, opts ...call.CallOption) (dismodel.ChannelDeleteResponse, error) {
	return internal.CallWithResponse[dismodel.ChannelDeleteResponse](
		call.DiscordChannelDelete, opts,
		dismodel.ChannelDeleteCall{
			ID: channelID,
		},
	)
}

func ChannelUpdatePermissions(args dismodel.ChannelUpdatePermissionsCall, opts ...call.CallOption) (dismodel.ChannelUpdatePermissionsResponse, error) {
	return internal.CallWithResponse[dismodel.ChannelUpdatePermissionsResponse](call.DiscordChannelUpdatePermissions, opts, args)
}

func ChannelDeletePermissions(args dismodel.ChannelDeletePermissionsCall, opts ...call.CallOption) (dismodel.ChannelDeletePermissionsResponse, error) {
	return internal.CallWithResponse[dismodel.ChannelDeletePermissionsResponse](call.DiscordChannelDeletePermissions, opts, args)
}

func ThreadStartFromMessage(args dismodel.ThreadStartFromMessageCall, opts ...call.CallOption) (dismodel.Channel, error) {
	return internal.CallWithResponse[dismodel.ThreadStartFromMessageResponse](call.DiscordThreadStartFromMessage, opts, args)
}

func ThreadStart(args dismodel.ThreadStartCall, opts ...call.CallOption) (dismodel.Channel, error) {
	return internal.CallWithResponse[dismodel.ThreadStartResponse](call.DiscordThreadStart, opts, args)
}

func ThreadStartInForum(args dismodel.ThreadStartInForumCall, opts ...call.CallOption) (dismodel.Channel, error) {
	return internal.CallWithResponse[dismodel.ThreadStartInForumResponse](call.DiscordThreadStartInForum, opts, args)
}

func ThreadJoin(threadID string, opts ...call.CallOption) (dismodel.ThreadJoinResponse, error) {
	return internal.CallWithResponse[dismodel.ThreadJoinResponse](
		call.DiscordThreadJoin, opts,
		dismodel.ThreadJoinCall{
			ChannelID: threadID,
		},
	)
}

func ThreadMemberAdd(threadId, userID string, opts ...call.CallOption) (dismodel.ThreadMemberAddResponse, error) {
	return internal.CallWithResponse[dismodel.ThreadMemberAddResponse](
		call.DiscordThreadMemberAdd, opts,
		dismodel.ThreadMemberAddCall{
			ChannelID: threadId,
			UserID:    userID,
		},
	)
}

func ThreadLeave(threadID string, opts ...call.CallOption) (dismodel.ThreadLeaveResponse, error) {
	return internal.CallWithResponse[dismodel.ThreadLeaveResponse](
		call.DiscordThreadLeave, opts,
		dismodel.ThreadLeaveCall{
			ChannelID: threadID,
		},
	)
}

func ThreadMemberRemove(threadID, userID string, opts ...call.CallOption) (dismodel.ThreadMemberRemoveResponse, error) {
	return internal.CallWithResponse[dismodel.ThreadMemberRemoveResponse](
		call.DiscordThreadMemberRemove, opts,
		dismodel.ThreadMemberRemoveCall{
			ChannelID: threadID,
			UserID:    userID,
		},
	)
}

func ThreadMemberGet(threadID, userID string, opts ...call.CallOption) (dismodel.ThreadMemberGetResponse, error) {
	return internal.CallWithResponse[dismodel.ThreadMemberGetResponse](
		call.DiscordThreadMemberGet, opts,
		dismodel.ThreadMemberGetCall{
			ChannelID: threadID,
			UserID:    userID,
		},
	)
}

func ThreadMemberList(threadID string, opts ...call.CallOption) (dismodel.ThreadMemberListResponse, error) {
	return internal.CallWithResponse[dismodel.ThreadMemberListResponse](
		call.DiscordThreadMemberList, opts,
		dismodel.ThreadMemberListCall{
			ChannelID: threadID,
		},
	)
}

// --- implementation done up to here ---

func ThreadListPublicArchived(channelID, before string, limit int, opts ...call.CallOption) (dismodel.ThreadListPublicArchivedResponse, error) {
	return internal.CallWithResponse[dismodel.ThreadListPublicArchivedResponse](
		call.DiscordThreadListPublicArchived, opts,
		dismodel.ThreadListPublicArchivedCall{
			ChannelID: channelID,
			Before:    before,
			Limit:     limit,
		},
	)
}

func ThreadListPrivateArchivedCall(channelID, before string, limit int, opts ...call.CallOption) (dismodel.ThreadListPrivateArchivedResponse, error) {
	return internal.CallWithResponse[dismodel.ThreadListPrivateArchivedResponse](
		call.DiscordThreadListPrivateArchived, opts,
		dismodel.ThreadListPrivateArchivedCall{
			ChannelID: channelID,
			Before:    before,
			Limit:     limit,
		},
	)
}

func ThreadListJoinedPrivateArchivedCall(channelID, before string, limit int, opts ...call.CallOption) (dismodel.ThreadListJoinedPrivateArchivedResponse, error) {
	return internal.CallWithResponse[dismodel.ThreadListJoinedPrivateArchivedResponse](
		call.DiscordThreadListJoinedPrivateArchived, opts,
		dismodel.ThreadListJoinedPrivateArchivedCall{
			ChannelID: channelID,
			Before:    before,
			Limit:     limit,
		},
	)
}

func ThreadListActive(opts ...call.CallOption) (dismodel.ThreadListActiveResponse, error) {
	return internal.CallWithResponse[dismodel.ThreadListActiveResponse](
		call.DiscordThreadListActive, opts,
		dismodel.ThreadListActiveCall{},
	)
}

func EmojiList(opts ...call.CallOption) (dismodel.EmojiListResponse, error) {
	return internal.CallWithResponse[dismodel.EmojiListResponse](
		call.DiscordEmojiList, opts,
		dismodel.EmojiListCall{},
	)
}

func EmojiGet(emojiID string, opts ...call.CallOption) (dismodel.EmojiGetResponse, error) {
	return internal.CallWithResponse[dismodel.EmojiGetResponse](
		call.DiscordEmojiGet, opts,
		dismodel.EmojiGetCall{
			EmojiID: emojiID,
		},
	)
}

func EmojiCreate(args dismodel.EmojiCreateCall, opts ...call.CallOption) (dismodel.EmojiCreateResponse, error) {
	return internal.CallWithResponse[dismodel.EmojiCreateResponse](call.DiscordEmojiCreate, opts, args)
}

func EmojiUpdate(args dismodel.EmojiUpdateCall, opts ...call.CallOption) (dismodel.EmojiUpdateResponse, error) {
	return internal.CallWithResponse[dismodel.EmojiUpdateResponse](call.DiscordEmojiUpdate, opts, args)
}

func EmojiDelete(emojiID string, opts ...call.CallOption) (dismodel.EmojiDeleteResponse, error) {
	return internal.CallWithResponse[dismodel.EmojiDeleteResponse](
		call.DiscordEmojiDelete, opts,
		dismodel.EmojiDeleteCall{
			EmojiID: emojiID,
		},
	)
}

func GuildGet(opts ...call.CallOption) (dismodel.GuildGetResponse, error) {
	return internal.CallWithResponse[dismodel.GuildGetResponse](
		call.DiscordGuildGet, opts,
		dismodel.GuildGetCall{},
	)
}

func GuildUpdate(args dismodel.GuildUpdateCall, opts ...call.CallOption) (dismodel.GuildUpdateResponse, error) {
	return internal.CallWithResponse[dismodel.GuildUpdateResponse](call.DiscordGuildUpdate, opts, args)
}

func InteractionResponseCreate(args dismodel.InteractionResponseCreateCall, opts ...call.CallOption) (dismodel.InteractionResponseCreateResponse, error) {
	return internal.CallWithResponse[dismodel.InteractionResponseCreateResponse](call.DiscordInteractionResponseCreate, opts, args)
}

func InteractionResponseUpdate(args dismodel.InteractionResponseUpdateCall, opts ...call.CallOption) (dismodel.InteractionResponseUpdateResponse, error) {
	return internal.CallWithResponse[dismodel.InteractionResponseUpdateResponse](call.DiscordInteractionResponseUpdate, opts, args)
}

func InteractionResponseDelete(interactionID, interactionToken string, opts ...call.CallOption) error {
	return internal.CallWithoutResponse(
		call.DiscordInteractionResponseDelete, opts,
		dismodel.InteractionResponseDeleteCall{
			ID:    interactionID,
			Token: interactionToken,
		},
	)
}

func InteractionFollowupCreate(args dismodel.InteractionFollowupCreateCall, opts ...call.CallOption) (dismodel.InteractionFollowupCreateResponse, error) {
	return internal.CallWithResponse[dismodel.InteractionFollowupCreateResponse](call.DiscordInteractionFollowupCreate, opts, args)
}

func InteractionFollowupUpdate(args dismodel.InteractionFollowupUpdateCall, opts ...call.CallOption) (dismodel.InteractionFollowupUpdateResponse, error) {
	return internal.CallWithResponse[dismodel.InteractionFollowupUpdateResponse](call.DiscordInteractionFollowupUpdate, opts, args)
}

func InteractionFollowupDelete(interactionID, interactionToken, messageID string, opts ...call.CallOption) error {
	return internal.CallWithoutResponse(
		call.DiscordInteractionFollowupDelete, opts,
		dismodel.InteractionFollowupDeleteCall{
			ID:        interactionID,
			Token:     interactionToken,
			MessageID: messageID,
		},
	)
}

func InteractionFollowupGet(interactionID, interactionToken, messageID string, opts ...call.CallOption) (dismodel.InteractionFollowupGetResponse, error) {
	return internal.CallWithResponse[dismodel.InteractionFollowupGetResponse](
		call.DiscordInteractionFollowupGet, opts,
		dismodel.InteractionFollowupGetCall{
			ID:        interactionID,
			Token:     interactionToken,
			MessageID: messageID,
		},
	)
}

func InviteListForChannel(channelID string, opts ...call.CallOption) (dismodel.InviteListForChannelResponse, error) {
	return internal.CallWithResponse[dismodel.InviteListForChannelResponse](
		call.DiscordInviteListForChannel, opts,
		dismodel.InviteListForChannelCall{
			ChannelID: channelID,
		},
	)
}

func InviteListForGuild(opts ...call.CallOption) (dismodel.InviteListForGuildResponse, error) {
	return internal.CallWithResponse[dismodel.InviteListForGuildResponse](
		call.DiscordInviteListForGuild, opts,
		dismodel.InviteListForGuildCall{},
	)
}

func InviteCreate(args dismodel.InviteCreateCall, opts ...call.CallOption) (dismodel.InviteCreateResponse, error) {
	return internal.CallWithResponse[dismodel.InviteCreateResponse](call.DiscordInviteCreate, opts, args)
}

func InviteGet(args dismodel.InviteGetCall, opts ...call.CallOption) (dismodel.InviteGetResponse, error) {
	return internal.CallWithResponse[dismodel.InviteGetResponse](call.DiscordInviteGet, opts, args)
}

func InviteDelete(code string, opts ...call.CallOption) (dismodel.InviteDeleteResponse, error) {
	return internal.CallWithResponse[dismodel.InviteDeleteResponse](
		call.DiscordInviteDelete, opts,
		dismodel.InviteDeleteCall{
			Code: code,
		},
	)
}

func MemberGet(userID string, opts ...call.CallOption) (dismodel.MemberGetResponse, error) {
	return internal.CallWithResponse[dismodel.MemberGetResponse](
		call.DiscordMemberGet, opts,
		dismodel.MemberGetCall{
			UserID: userID,
		},
	)
}

func MemberList(opts ...call.CallOption) (dismodel.MemberListResponse, error) {
	return internal.CallWithResponse[dismodel.MemberListResponse](
		call.DiscordMemberList, opts,
		dismodel.MemberListCall{},
	)
}

func MemberSearch(query string, limit int, opts ...call.CallOption) (dismodel.MemberSearchResponse, error) {
	return internal.CallWithResponse[dismodel.MemberSearchResponse](
		call.DiscordMemberSearch, opts,
		dismodel.MemberSearchCall{
			Query: query,
			Limit: limit,
		},
	)
}

func MemberUpdate(args dismodel.MemberUpdateCall, opts ...call.CallOption) (dismodel.MemberUpdateResponse, error) {
	return internal.CallWithResponse[dismodel.MemberUpdateResponse](call.DiscordMemberUpdate, opts, args)
}

func MemberUpdateOwn(args dismodel.MemberUpdateOwnCall, opts ...call.CallOption) (dismodel.MemberUpdateOwnResponse, error) {
	return internal.CallWithResponse[dismodel.MemberUpdateOwnResponse](call.DiscordMemberUpdateOwn, opts, args)
}

func MemberRoleAdd(userID, roleID string, opts ...call.CallOption) error {
	return internal.CallWithoutResponse(
		call.DiscordMemberRoleAdd, opts,
		dismodel.MemberRoleAddCall{
			UserID: userID,
			RoleID: roleID,
		},
	)
}

func MemberRoleRemove(userID, roleID string, opts ...call.CallOption) error {
	return internal.CallWithoutResponse(
		call.DiscordMemberRoleRemove, opts,
		dismodel.MemberRoleRemoveCall{
			UserID: userID,
			RoleID: roleID,
		},
	)
}

func MemberRemove(userID string, opts ...call.CallOption) error {
	return internal.CallWithoutResponse(
		call.DiscordMemberRemove, opts,
		dismodel.MemberRemoveCall{
			UserID: userID,
		},
	)
}

func MemberPruneCount(args dismodel.MemberPruneCountCall, opts ...call.CallOption) (dismodel.MemberPruneCountResponse, error) {
	return internal.CallWithResponse[dismodel.MemberPruneCountResponse](call.DiscordMemberPruneCount, opts, args)
}

func MemberPruneBegin(args dismodel.MemberPruneBeginCall, opts ...call.CallOption) (dismodel.MemberPruneBeginResponse, error) {
	return internal.CallWithResponse[dismodel.MemberPruneBeginResponse](call.DiscordMemberPruneBegin, opts, args)
}

func MessageList(args dismodel.MessageListCall, opts ...call.CallOption) (dismodel.MessageListResponse, error) {
	return internal.CallWithResponse[dismodel.MessageListResponse](call.DiscordMessageList, opts, args)
}

func MessageGet(channelID, messageID string, opts ...call.CallOption) (dismodel.MessageGetResponse, error) {
	return internal.CallWithResponse[dismodel.MessageGetResponse](
		call.DiscordMessageGet, opts,
		dismodel.MessageGetCall{
			ChannelID: channelID,
			MessageID: messageID,
		},
	)
}

func MessageCreate(args dismodel.MessageCreateCall, opts ...call.CallOption) (dismodel.MessageCreateResponse, error) {
	return internal.CallWithResponse[dismodel.MessageCreateResponse](call.DiscordMessageCreate, opts, args)
}

func MessageUpdate(args dismodel.MessageUpdateCall, opts ...call.CallOption) (dismodel.MessageUpdateResponse, error) {
	return internal.CallWithResponse[dismodel.MessageUpdateResponse](call.DiscordMessageUpdate, opts, args)
}

func MessageDelete(channelID, messageID string, opts ...call.CallOption) error {
	return internal.CallWithoutResponse(
		call.DiscordMessageDelete, opts,
		dismodel.MessageDeleteCall{
			ChannelID: channelID,
			MessageID: messageID,
		},
	)
}

func MessageDeleteBulk(channelID string, messages []string, opts ...call.CallOption) error {
	return internal.CallWithoutResponse(
		call.DiscordMessageDeleteBulk, opts,
		dismodel.MessageDeleteBulkCall{
			ChannelID: channelID,
			Messages:  messages,
		},
	)
}

func MessageReactionCreate(channelID, messageID, emoji string, opts ...call.CallOption) error {
	return internal.CallWithoutResponse(
		call.DiscordMessageReactionCreate, opts,
		dismodel.MessageReactionCreateCall{
			ChannelID: channelID,
			MessageID: messageID,
			Emoji:     emoji,
		},
	)
}

func MessageReactionDeleteOwn(channelID, messageID, emoji string, opts ...call.CallOption) error {
	return internal.CallWithoutResponse(
		call.DiscordMessageReactionDeleteOwn, opts,
		dismodel.MessageReactionDeleteOwnCall{
			ChannelID: channelID,
			MessageID: messageID,
			Emoji:     emoji,
		},
	)
}

func MessageReactionDeleteUser(channelID, messageID, emoji, userID string, opts ...call.CallOption) error {
	return internal.CallWithoutResponse(
		call.DiscordMessageReactionDeleteUser, opts,
		dismodel.MessageReactionDeleteUserCall{
			ChannelID: channelID,
			MessageID: messageID,
			Emoji:     emoji,
			UserID:    userID,
		},
	)
}

func MessageReactionList(args dismodel.MessageReactionListCall, opts ...call.CallOption) (dismodel.MessageReactionListResponse, error) {
	return internal.CallWithResponse[dismodel.MessageReactionListResponse](call.DiscordMessageReactionList, opts, args)
}

func MessageReactionDeleteAll(channelID, messageID string, opts ...call.CallOption) error {
	return internal.CallWithoutResponse(
		call.DiscordMessageReactionDeleteAll, opts,
		dismodel.MessageReactionDeleteAllCall{
			ChannelID: channelID,
			MessageID: messageID,
		},
	)
}

func MessageReactionDeleteEmoji(channelID, messageID, emoji string, opts ...call.CallOption) error {
	return internal.CallWithoutResponse(
		call.DiscordMessageReactionDeleteEmoji, opts,
		dismodel.MessageReactionDeleteEmojiCall{
			ChannelID: channelID,
			MessageID: messageID,
			Emoji:     emoji,
		},
	)
}

func MessageGetPinned(channelID string, opts ...call.CallOption) (dismodel.MessageGetPinnedResponse, error) {
	return internal.CallWithResponse[dismodel.MessageGetPinnedResponse](
		call.DiscordMessageGetPinned, opts,
		dismodel.MessageGetPinnedCall{
			ChannelID: channelID,
		},
	)
}

func MessagePin(channelID, messageID string, opts ...call.CallOption) error {
	return internal.CallWithoutResponse(
		call.DiscordMessagePin, opts,
		dismodel.MessagePinCall{
			ChannelID: channelID,
			MessageID: messageID,
		},
	)
}

func MessageUnpin(channelID, messageID string, opts ...call.CallOption) error {
	return internal.CallWithoutResponse(
		call.DiscordMessageUnpin, opts,
		dismodel.MessageUnpinCall{
			ChannelID: channelID,
			MessageID: messageID,
		},
	)
}

func RoleList(opts ...call.CallOption) (dismodel.RoleListResponse, error) {
	return internal.CallWithResponse[dismodel.RoleListResponse](
		call.DiscordRoleList, opts,
		dismodel.RoleListCall{},
	)
}

func RoleCreate(args dismodel.RoleCreateCall, opts ...call.CallOption) (dismodel.RoleCreateResponse, error) {
	return internal.CallWithResponse[dismodel.RoleCreateResponse](call.DiscordRoleCreate, opts, args)
}

func RoleUpdate(args dismodel.RoleUpdateCall, opts ...call.CallOption) (dismodel.RoleUpdateResponse, error) {
	return internal.CallWithResponse[dismodel.RoleUpdateResponse](call.DiscordRoleUpdate, opts, args)
}

func RoleUpdatePositions(args dismodel.RoleUpdatePositionsCall, opts ...call.CallOption) (dismodel.RoleListResponse, error) {
	return internal.CallWithResponse[dismodel.RoleListResponse](call.DiscordRoleUpdatePositions, opts, args)
}

func RoleDelete(roleID string, opts ...call.CallOption) (dismodel.RoleDeleteResponse, error) {
	return internal.CallWithResponse[dismodel.RoleDeleteResponse](
		call.DiscordRoleDelete, opts,
		dismodel.RoleDeleteCall{
			RoleID: roleID,
		},
	)
}

func ScheduledEventList(withUserCount bool, opts ...call.CallOption) (dismodel.ScheduledEventListResponse, error) {
	return internal.CallWithResponse[dismodel.ScheduledEventListResponse](
		call.DiscordScheduledEventList, opts,
		dismodel.ScheduledEventListCall{
			WithUserCount: withUserCount,
		},
	)
}

func ScheduledEventGet(eventID string, opts ...call.CallOption) (dismodel.ScheduledEventGetResponse, error) {
	return internal.CallWithResponse[dismodel.ScheduledEventGetResponse](
		call.DiscordScheduledEventGet, opts,
		dismodel.ScheduledEventGetCall{
			ID: eventID,
		},
	)
}

func ScheduledEventCreate(args dismodel.ScheduledEventCreateCall, opts ...call.CallOption) (dismodel.ScheduledEventCreateResponse, error) {
	return internal.CallWithResponse[dismodel.ScheduledEventCreateResponse](call.DiscordScheduledEventCreate, opts, args)
}

func ScheduledEventUpdate(args dismodel.ScheduledEventUpdateCall, opts ...call.CallOption) (dismodel.ScheduledEventUpdateResponse, error) {
	return internal.CallWithResponse[dismodel.ScheduledEventUpdateResponse](call.DiscordScheduledEventUpdate, opts, args)
}

func ScheduledEventDelete(eventID string, opts ...call.CallOption) (dismodel.ScheduledEventDeleteResponse, error) {
	return internal.CallWithResponse[dismodel.ScheduledEventDeleteResponse](
		call.DiscordScheduledEventDelete, opts,
		dismodel.ScheduledEventDeleteCall{
			ID: eventID,
		},
	)
}

func ScheduledEventUserList(args dismodel.ScheduledEventUserListCall, opts ...call.CallOption) (dismodel.ScheduledEventUserListResponse, error) {
	return internal.CallWithResponse[dismodel.ScheduledEventUserListResponse](call.DiscordScheduledEventUserList, opts, args)
}

func StageInstanceCreate(args dismodel.StageInstanceCreateCall, opts ...call.CallOption) (dismodel.StageInstanceCreateResponse, error) {
	return internal.CallWithResponse[dismodel.StageInstanceCreateResponse](call.DiscordStageInstanceCreate, opts, args)
}

func StageInstanceGet(channelID string, opts ...call.CallOption) (dismodel.StageInstanceGetResponse, error) {
	return internal.CallWithResponse[dismodel.StageInstanceGetResponse](
		call.DiscordStageInstanceGet, opts,
		dismodel.StageInstanceGetCall{
			ChannelID: channelID,
		},
	)
}

func StageInstanceUpdate(args dismodel.StageInstanceUpdateCall, opts ...call.CallOption) (dismodel.StageInstanceUpdateResponse, error) {
	return internal.CallWithResponse[dismodel.StageInstanceUpdateResponse](call.DiscordStageInstanceUpdate, opts, args)
}

func StageInstanceDelete(channelID string, opts ...call.CallOption) (dismodel.StageInstanceDeleteResponse, error) {
	return internal.CallWithResponse[dismodel.StageInstanceDeleteResponse](
		call.DiscordStageInstanceDelete, opts,
		dismodel.StageInstanceDeleteCall{
			ChannelID: channelID,
		},
	)
}

func StickerList(opts ...call.CallOption) (dismodel.StickerListResponse, error) {
	return internal.CallWithResponse[dismodel.StickerListResponse](
		call.DiscordStickerList, opts,
		dismodel.StickerListCall{},
	)
}

func StickerGet(stickerID string, opts ...call.CallOption) (dismodel.StickerGetResponse, error) {
	return internal.CallWithResponse[dismodel.StickerGetResponse](
		call.DiscordStickerGet, opts,
		dismodel.StickerGetCall{
			ID: stickerID,
		},
	)
}

func StickerCreate(args dismodel.StickerCreateCall, opts ...call.CallOption) (dismodel.StickerCreateResponse, error) {
	return internal.CallWithResponse[dismodel.StickerCreateResponse](call.DiscordStickerCreate, opts, args)
}

func StickerUpdate(args dismodel.StickerUpdateCall, opts ...call.CallOption) (dismodel.StickerUpdateResponse, error) {
	return internal.CallWithResponse[dismodel.StickerUpdateResponse](call.DiscordStickerUpdate, opts, args)
}

func StickerDelete(stickerID string, opts ...call.CallOption) (dismodel.StickerDeleteResponse, error) {
	return internal.CallWithResponse[dismodel.StickerDeleteResponse](
		call.DiscordStickerDelete, opts,
		dismodel.StickerDeleteCall{
			ID: stickerID,
		},
	)
}

func UserGet(userID string, opts ...call.CallOption) (dismodel.UserGetResponse, error) {
	return internal.CallWithResponse[dismodel.UserGetResponse](
		call.DiscordUserGet, opts,
		dismodel.UserGetCall{
			UserID: userID,
		},
	)
}

func WebhookGet(webhookID string, opts ...call.CallOption) (dismodel.WebhookGetResponse, error) {
	return internal.CallWithResponse[dismodel.WebhookGetResponse](
		call.DiscordWebhookGet, opts,
		dismodel.WebhookGetCall{
			ID: webhookID,
		},
	)
}

func WebhookListForChannel(channelID string, opts ...call.CallOption) (dismodel.WebhookListForChannelResponse, error) {
	return internal.CallWithResponse[dismodel.WebhookListForChannelResponse](
		call.DiscordWebhookListForChannel, opts,
		dismodel.WebhookListForChannelCall{
			ChannelID: channelID,
		},
	)
}

func WebhookListForGuild(opts ...call.CallOption) (dismodel.WebhookListForGuildResponse, error) {
	return internal.CallWithResponse[dismodel.WebhookListForGuildResponse](
		call.DiscordWebhookListForGuild, opts,
		dismodel.WebhookListForGuildCall{},
	)
}

func WebhookCreate(args dismodel.WebhookCreateCall, opts ...call.CallOption) (dismodel.WebhookCreateResponse, error) {
	return internal.CallWithResponse[dismodel.WebhookCreateResponse](call.DiscordWebhookCreate, opts, args)
}

func WebhookUpdate(args dismodel.WebhookUpdateCall, opts ...call.CallOption) (dismodel.WebhookUpdateResponse, error) {
	return internal.CallWithResponse[dismodel.WebhookUpdateResponse](call.DiscordWebhookUpdate, opts, args)
}

func WebhookDelete(webhookID string, opts ...call.CallOption) (dismodel.WebhookDeleteResponse, error) {
	return internal.CallWithResponse[dismodel.WebhookDeleteResponse](
		call.DiscordWebhookDelete, opts,
		dismodel.WebhookDeleteCall{
			ID: webhookID,
		},
	)
}

func WebhookGetWithToken(webhookID, webhookToken string, opts ...call.CallOption) (dismodel.WebhookGetWithTokenResponse, error) {
	return internal.CallWithResponse[dismodel.WebhookGetWithTokenResponse](
		call.DiscordWebhookGetWithToken, opts,
		dismodel.WebhookGetWithTokenCall{
			ID:    webhookID,
			Token: webhookToken,
		},
	)
}

func WebhookUpdateWithToken(args dismodel.WebhookUpdateWithTokenCall, opts ...call.CallOption) (dismodel.WebhookUpdateWithTokenResponse, error) {
	return internal.CallWithResponse[dismodel.WebhookUpdateWithTokenResponse](call.DiscordWebhookUpdateWithToken, opts, args)
}

func WebhookDeleteWithToken(webhookID, webhookToken string, opts ...call.CallOption) (dismodel.WebhookDeleteWithTokenResponse, error) {
	return internal.CallWithResponse[dismodel.WebhookDeleteWithTokenResponse](
		call.DiscordWebhookDeleteWithToken, opts,
		dismodel.WebhookDeleteWithTokenCall{
			ID:    webhookID,
			Token: webhookToken,
		},
	)
}

func WebhookExecute(args dismodel.WebhookExecuteCall, opts ...call.CallOption) (dismodel.WebhookExecuteResponse, error) {
	return internal.CallWithResponse[dismodel.WebhookExecuteResponse](call.DiscordWebhookExecute, opts, args)
}
