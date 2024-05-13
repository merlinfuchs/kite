package host

import (
	"context"

	"github.com/disgoorg/disgo/rest"
	"github.com/merlinfuchs/dismod/distype"
	"github.com/merlinfuchs/kite/kite-service/pkg/store"
)

func (h HostEnvironment) callDiscordBanList(ctx context.Context, app store.AppStateProvider, data distype.BanListRequest) (res *distype.BanListResponse, err error) {
	params := make(map[string]any)
	if data.Before != nil {
		params["before"] = *data.Before
	}
	if data.After != nil {
		params["after"] = *data.After
	}
	if data.Limit != nil {
		params["limit"] = *data.Limit
	}

	err = app.Client().Request(
		rest.GetBans.Compile(params, data.GuildID),
		nil, &res, rest.WithCtx(ctx),
	)
	return
}

func (h HostEnvironment) callDiscordBanGet(ctx context.Context, app store.AppStateProvider, data distype.BanGetRequest) (res *distype.BanGetResponse, err error) {
	err = app.Client().Request(
		rest.GetBan.Compile(nil, data.GuildID, data.UserID),
		nil, &res, rest.WithCtx(ctx),
	)
	return
}

func (h HostEnvironment) callDiscordBanCreate(ctx context.Context, app store.AppStateProvider, data distype.BanCreateRequest) (res *distype.BanCreateResponse, err error) {
	err = app.Client().Request(
		rest.AddBan.Compile(nil, data.GuildID, data.UserID),
		data, &res, rest.WithCtx(ctx),
	)
	return
}

func (h HostEnvironment) callDiscordBanRemove(ctx context.Context, app store.AppStateProvider, data distype.BanRemoveRequest) (res *distype.BanRemoveResponse, err error) {
	err = app.Client().Request(
		rest.DeleteBan.Compile(nil, data.GuildID, data.UserID),
		nil, &res, rest.WithCtx(ctx),
	)
	return
}

func (h HostEnvironment) callDiscordChannelGet(ctx context.Context, app store.AppStateProvider, data distype.ChannelGetRequest) (*distype.ChannelGetResponse, error) {
	channel, err := app.State().GetChannel(ctx, data.ChannelID)
	if err != nil {
		return nil, err
	}

	return channel, nil
}

func (h HostEnvironment) callDiscordChannelList(ctx context.Context, app store.AppStateProvider, data distype.GuildChannelListRequest) (*distype.GuildChannelListResponse, error) {
	channels, err := app.State().GetGuildChannels(ctx, data.GuildID)
	if err != nil {
		return nil, err
	}

	return &channels, nil
}

func (h HostEnvironment) callDiscordChannelCreate(ctx context.Context, app store.AppStateProvider, data distype.GuildChannelCreateRequest) (res *distype.GuildChannelCreateResponse, err error) {
	err = app.Client().Request(
		rest.CreateGuildChannel.Compile(nil, data.GuildID),
		data, &res, rest.WithCtx(ctx),
	)
	return
}

func (h HostEnvironment) callDiscordChannelUpdate(ctx context.Context, app store.AppStateProvider, data distype.ChannelModifyRequest) (res *distype.ChannelModifyResponse, err error) {
	err = app.Client().Request(
		rest.UpdateChannel.Compile(nil, data.ChannelID),
		data, &res, rest.WithCtx(ctx),
	)
	return
}

func (h HostEnvironment) callDiscordChannelUpdatePositions(ctx context.Context, app store.AppStateProvider, data distype.GuildChannelModifyPositionsRequest) (*distype.GuildChannelModifyPositionsResponse, error) {
	err := app.Client().Request(
		rest.UpdateChannelPositions.Compile(nil, data.GuildID),
		data.Entries, nil, rest.WithCtx(ctx),
	)
	return &distype.GuildChannelModifyPositionsResponse{}, err
}

func (h HostEnvironment) callDiscordChannelDelete(ctx context.Context, app store.AppStateProvider, data distype.ChannelDeleteRequest) (res *distype.ChannelDeleteResponse, err error) {
	err = app.Client().Request(
		rest.DeleteChannel.Compile(nil, data.ChannelID),
		nil, &res, rest.WithCtx(ctx),
	)
	return
}

func (h HostEnvironment) callDiscordChannelUpdatePermissions(ctx context.Context, app store.AppStateProvider, data distype.ChannelEditPermissionsRequest) (*distype.ChannelEditPermissionsResponse, error) {
	err := app.Client().Request(
		rest.UpdatePermissionOverwrite.Compile(nil, data.ChannelID, data.OverwriteID),
		data, nil, rest.WithCtx(ctx),
	)
	return &distype.ChannelEditPermissionsResponse{}, err
}

func (h HostEnvironment) callDiscordChannelDeletePermissions(ctx context.Context, app store.AppStateProvider, data distype.ChannelDeletePermissionsRequest) (*distype.ChannelDeletePermissionsResponse, error) {
	err := app.Client().Request(
		rest.DeletePermissionOverwrite.Compile(nil, data.ChannelID, data.OverwriteID),
		nil, nil, rest.WithCtx(ctx),
	)
	return &distype.ChannelDeletePermissionsResponse{}, err
}

func (h HostEnvironment) callDiscordThreadStartFromMessage(ctx context.Context, app store.AppStateProvider, data distype.ThreadStartFromMessageRequest) (res *distype.ThreadStartFromMessageResponse, err error) {
	err = app.Client().Request(
		rest.CreateThreadWithMessage.Compile(nil, data.ChannelID, data.MessageID),
		data, &res, rest.WithCtx(ctx),
	)
	return
}

func (h HostEnvironment) callDiscordThreadStart(ctx context.Context, app store.AppStateProvider, data distype.ThreadStartWithoutMessageRequest) (res *distype.ThreadStartWithoutMessageResponse, err error) {
	err = app.Client().Request(
		rest.CreateThread.Compile(nil, data.ChannelID),
		data, &res, rest.WithCtx(ctx),
	)
	return
}

func (h HostEnvironment) callDiscordThreadStartInForum(ctx context.Context, app store.AppStateProvider, data distype.ThreadStartInForumRequest) (res *distype.ThreadStartInForumResponse, err error) {
	err = app.Client().Request(
		rest.CreateThread.Compile(nil, data.ChannelID),
		data, &res, rest.WithCtx(ctx),
	)
	return
}

func (h HostEnvironment) callDiscordThreadJoin(ctx context.Context, app store.AppStateProvider, data distype.ThreadJoinRequest) (*distype.ThreadJoinResponse, error) {
	err := app.Client().Request(
		rest.JoinThread.Compile(nil, data.ChannelID),
		data, nil, rest.WithCtx(ctx),
	)
	return &distype.ThreadJoinResponse{}, err
}

func (h HostEnvironment) callDiscordThreadMemberAdd(ctx context.Context, app store.AppStateProvider, data distype.ThreadMemberAddRequest) (*distype.ThreadMemberAddResponse, error) {
	err := app.Client().Request(
		rest.AddThreadMember.Compile(nil, data.ChannelID, data.UserID),
		data, nil, rest.WithCtx(ctx),
	)
	return &distype.ThreadMemberAddResponse{}, err
}

func (h HostEnvironment) callDiscordThreadLeave(ctx context.Context, app store.AppStateProvider, data distype.ThreadLeaveRequest) (*distype.ThreadLeaveResponse, error) {
	err := app.Client().Request(
		rest.LeaveThread.Compile(nil, data.ChannelID),
		data, nil, rest.WithCtx(ctx),
	)
	return &distype.ThreadLeaveResponse{}, err
}

func (h HostEnvironment) callDiscordThreadMemberRemove(ctx context.Context, app store.AppStateProvider, data distype.ThreadMemberRemoveRequest) (*distype.ThreadMemberRemoveResponse, error) {
	err := app.Client().Request(
		rest.RemoveThreadMember.Compile(nil, data.ChannelID, data.UserID),
		data, nil, rest.WithCtx(ctx),
	)
	return &distype.ThreadMemberRemoveResponse{}, err
}

func (h HostEnvironment) callDiscordThreadMemberGet(ctx context.Context, app store.AppStateProvider, data distype.ThreadMemberGetRequest) (res *distype.ThreadMemberGetResponse, err error) {
	err = app.Client().Request(
		rest.GetThreadMember.Compile(nil, data.ChannelID, data.UserID),
		nil, &res, rest.WithCtx(ctx),
	)
	return
}

func (h HostEnvironment) callDiscordThreadMemberList(ctx context.Context, app store.AppStateProvider, data distype.ThreadMemberListRequest) (res *distype.ThreadMemberListResponse, err error) {
	params := make(map[string]any)
	if data.WithMember != nil {
		params["with_member"] = *data.WithMember
	}
	if data.After != nil {
		params["before"] = *data.After
	}
	if data.Limit != nil {
		params["limit"] = *data.Limit
	}

	err = app.Client().Request(
		rest.GetThreadMembers.Compile(params, data.ChannelID),
		nil, &res, rest.WithCtx(ctx),
	)
	return
}

func (h HostEnvironment) callDiscordThreadListPublicArchived(ctx context.Context, app store.AppStateProvider, data distype.ThreadListPublicArchivedRequest) (res *distype.ThreadListPublicArchivedResponse, err error) {
	params := make(map[string]any)
	if data.Before != nil {
		params["before"] = *data.Before
	}
	if data.Limit != nil {
		params["limit"] = *data.Limit
	}

	err = app.Client().Request(
		rest.GetPublicArchivedThreads.Compile(params, data.ChannelID),
		nil, &res, rest.WithCtx(ctx),
	)
	return
}

func (h HostEnvironment) callDiscordThreadListPrivateArchived(ctx context.Context, app store.AppStateProvider, data distype.ThreadListPrivateArchivedRequest) (res *distype.ThreadListPrivateArchivedResponse, err error) {
	params := make(map[string]any)
	if data.Before != nil {
		params["before"] = *data.Before
	}
	if data.Limit != nil {
		params["limit"] = *data.Limit
	}

	err = app.Client().Request(
		rest.GetPrivateArchivedThreads.Compile(params, data.ChannelID),
		nil, &res, rest.WithCtx(ctx),
	)
	return
}

func (h HostEnvironment) callDiscordThreadListJoinedPrivateArchived(ctx context.Context, app store.AppStateProvider, data distype.ThreadListJoinedPrivateArchivedRequest) (res *distype.ThreadListJoinedPrivateArchivedResponse, err error) {
	params := make(map[string]any)
	if data.Before != nil {
		params["before"] = *data.Before
	}
	if data.Limit != nil {
		params["limit"] = *data.Limit
	}

	err = app.Client().Request(
		rest.GetJoinedPrivateArchivedThreads.Compile(params, data.ChannelID),
		nil, &res, rest.WithCtx(ctx),
	)
	return
}

func (h HostEnvironment) callDiscordThreadListActive(ctx context.Context, app store.AppStateProvider, data distype.GuildThreadListActiveRequest) (res *distype.GuildThreadListActiveResponse, err error) {
	// TODO: Implement
	return
}

func (h HostEnvironment) callDiscordEmojiList(ctx context.Context, app store.AppStateProvider, data distype.GuildEmojiListRequest) (res *distype.GuildEmojiListResponse, err error) {
	err = app.Client().Request(
		rest.GetEmojis.Compile(nil, data.GuildID),
		nil, &res, rest.WithCtx(ctx),
	)
	return
}

func (h HostEnvironment) callDiscordEmojiGet(ctx context.Context, app store.AppStateProvider, data distype.EmojiGetRequest) (res *distype.EmojiGetResponse, err error) {
	err = app.Client().Request(
		rest.GetEmoji.Compile(nil, data.GuildID, data.EmojiID),
		nil, &res, rest.WithCtx(ctx),
	)
	return
}

func (h HostEnvironment) callDiscordEmojiCreate(ctx context.Context, app store.AppStateProvider, data distype.EmojiCreateRequest) (res *distype.EmojiCreateResponse, err error) {
	err = app.Client().Request(
		rest.CreateEmoji.Compile(nil, data.GuildID),
		data, &res, rest.WithCtx(ctx),
	)
	return
}

func (h HostEnvironment) callDiscordEmojiUpdate(ctx context.Context, app store.AppStateProvider, data distype.EmojiModifyRequest) (res *distype.EmojiModifyResponse, err error) {
	err = app.Client().Request(
		rest.UpdateEmoji.Compile(nil, data.GuildID, data.EmojiID),
		data, &res, rest.WithCtx(ctx),
	)
	return
}

func (h HostEnvironment) callDiscordEmojiDelete(ctx context.Context, app store.AppStateProvider, data distype.EmojiDeleteRequest) (*distype.EmojiDeleteResponse, error) {
	err := app.Client().Request(
		rest.DeleteEmoji.Compile(nil, data.GuildID, data.EmojiID),
		nil, nil, rest.WithCtx(ctx),
	)
	return &distype.EmojiDeleteResponse{}, err
}

func (h HostEnvironment) callDiscordGuildGet(ctx context.Context, app store.AppStateProvider, data distype.GuildGetRequest) (*distype.GuildGetResponse, error) {
	guild, err := app.State().GetGuild(ctx, data.GuildID)
	if err != nil {
		return nil, err
	}

	return guild, nil
}

func (h HostEnvironment) callDiscordGuildUpdate(ctx context.Context, app store.AppStateProvider, data distype.GuildUpdateRequest) (res *distype.GuildUpdateResponse, err error) {
	err = app.Client().Request(
		rest.UpdateGuild.Compile(nil, data.GuildID),
		data, &res, rest.WithCtx(ctx),
	)
	return
}

func (h HostEnvironment) callDiscordInteractionResponseCreate(ctx context.Context, app store.AppStateProvider, data distype.InteractionResponseCreateRequest) (*distype.InteractionResponseCreateResponse, error) {
	err := app.Client().Request(
		rest.CreateInteractionResponse.Compile(nil, data.InteractionID, data.InteractionToken),
		data, nil, rest.WithCtx(ctx),
	)
	return &distype.InteractionResponseCreateResponse{}, err
}

func (h HostEnvironment) callDiscordInteractionResponseUpdate(ctx context.Context, app store.AppStateProvider, data distype.InteractionResponseEditRequest) (res *distype.InteractionResponseEditResponse, err error) {
	err = app.Client().Request(
		rest.UpdateInteractionResponse.Compile(nil, data.ApplicationID, data.InteractionToken),
		data, &res, rest.WithCtx(ctx),
	)
	return
}

func (h HostEnvironment) callDiscordInteractionResponseDelete(ctx context.Context, app store.AppStateProvider, data distype.InteractionResponseDeleteRequest) (*distype.InteractionResponseDeleteResponse, error) {
	err := app.Client().Request(
		rest.DeleteInteractionResponse.Compile(nil, data.ApplicationID, data.InteractionToken),
		nil, nil, rest.WithCtx(ctx),
	)
	return &distype.InteractionResponseDeleteResponse{}, err
}

func (h HostEnvironment) callDiscordInteractionResponseGet(ctx context.Context, app store.AppStateProvider, data distype.InteractionResponseGetRequest) (res *distype.InteractionResponseGetResponse, err error) {
	err = app.Client().Request(
		rest.GetInteractionResponse.Compile(nil, data.ApplicationID, data.InteractionToken),
		nil, &res, rest.WithCtx(ctx),
	)
	return
}

func (h HostEnvironment) callDiscordInteractionFollowupCreate(ctx context.Context, app store.AppStateProvider, data distype.InteractionFollowupCreateRequest) (res *distype.InteractionFollowupCreateResponse, err error) {
	err = app.Client().Request(
		rest.CreateFollowupMessage.Compile(nil, data.ApplicationID, data.InteractionToken),
		data, &res, rest.WithCtx(ctx),
	)
	return
}

func (h HostEnvironment) callDiscordInteractionFollowupUpdate(ctx context.Context, app store.AppStateProvider, data distype.InteractionFollowupEditRequest) (res *distype.InteractionFollowupEditResponse, err error) {
	err = app.Client().Request(
		rest.UpdateFollowupMessage.Compile(nil, data.ApplicationID, data.InteractionToken, data.MessageID),
		data, &res, rest.WithCtx(ctx),
	)
	return
}

func (h HostEnvironment) callDiscordInteractionFollowupDelete(ctx context.Context, app store.AppStateProvider, data distype.InteractionFollowupDeleteRequest) (*distype.InteractionFollowupDeleteResponse, error) {
	err := app.Client().Request(
		rest.DeleteFollowupMessage.Compile(nil, data.ApplicationID, data.InteractionToken, data.MessageID),
		nil, nil, rest.WithCtx(ctx),
	)
	return &distype.InteractionFollowupDeleteResponse{}, err
}

func (h HostEnvironment) callDiscordInteractionFollowupGet(ctx context.Context, app store.AppStateProvider, data distype.InteractionFollowupGetRequest) (res *distype.InteractionFollowupGetResponse, err error) {
	err = app.Client().Request(
		rest.GetFollowupMessage.Compile(nil, data.ApplicationID, data.InteractionToken, data.MessageID),
		nil, &res, rest.WithCtx(ctx),
	)
	return
}

func (h HostEnvironment) callDiscordInviteListForChannel(ctx context.Context, app store.AppStateProvider, data distype.ChannelInviteListRequest) (res *distype.ChannelInviteListResponse, err error) {
	err = app.Client().Request(
		rest.GetChannelInvites.Compile(nil, data.ChannelID),
		nil, &res, rest.WithCtx(ctx),
	)
	return
}

func (h HostEnvironment) callDiscordInviteListForGuild(ctx context.Context, app store.AppStateProvider, data distype.GuildInviteListRequest) (res *distype.GuildInviteListResponse, err error) {
	err = app.Client().Request(
		rest.GetGuildInvites.Compile(nil, data.GuildID),
		nil, &res, rest.WithCtx(ctx),
	)
	return
}

func (h HostEnvironment) callDiscordInviteCreate(ctx context.Context, app store.AppStateProvider, data distype.ChannelInviteCreateRequest) (res *distype.ChannelInviteCreateResponse, err error) {
	err = app.Client().Request(
		rest.CreateInvite.Compile(nil, data.ChannelID),
		data, &res, rest.WithCtx(ctx),
	)
	return
}

func (h HostEnvironment) callDiscordInviteGet(ctx context.Context, app store.AppStateProvider, data distype.InviteGetRequest) (res *distype.InviteGetResponse, err error) {
	params := make(map[string]any)
	if data.WithCounts != nil {
		params["with_counts"] = *data.WithCounts
	}
	if data.WithExpiration != nil {
		params["with_expiration"] = *data.WithExpiration
	}
	if data.GuildScheduledEventID != nil {
		params["guild_scheduled_event_id"] = *data.GuildScheduledEventID
	}

	err = app.Client().Request(
		rest.GetInvite.Compile(params, data.Code),
		nil, &res, rest.WithCtx(ctx),
	)
	return
}

func (h HostEnvironment) callDiscordInviteDelete(ctx context.Context, app store.AppStateProvider, data distype.InviteDeleteRequest) (*distype.InviteDeleteResponse, error) {
	err := app.Client().Request(
		rest.DeleteInvite.Compile(nil, data.Code),
		nil, nil, rest.WithCtx(ctx),
	)
	return &distype.InviteDeleteResponse{}, err
}

func (h HostEnvironment) callDiscordMemberGet(ctx context.Context, app store.AppStateProvider, data distype.MemberGetRequest) (res *distype.MemberGetResponse, err error) {
	err = app.Client().Request(
		rest.GetMember.Compile(nil, data.GuildID, data.UserID),
		nil, &res, rest.WithCtx(ctx),
	)
	return
}

func (h HostEnvironment) callDiscordMemberList(ctx context.Context, app store.AppStateProvider, data distype.GuildMemberListRequest) (res *distype.GuildMemberListResponse, err error) {
	params := make(map[string]any)
	if data.After != nil {
		params["after"] = *data.After
	}
	if data.Limit != nil {
		params["limit"] = *data.Limit
	}

	err = app.Client().Request(
		rest.GetMembers.Compile(params, data.GuildID),
		nil, &res, rest.WithCtx(ctx),
	)
	return
}

func (h HostEnvironment) callDiscordMemberSearch(ctx context.Context, app store.AppStateProvider, data distype.GuildMemberSearchRequest) (res *distype.GuildMemberSearchResponse, err error) {
	params := make(map[string]any)
	params["query"] = data.Query
	if data.Limit != nil {
		params["limit"] = *data.Limit
	}

	err = app.Client().Request(
		rest.SearchMembers.Compile(params, data.GuildID),
		nil, &res, rest.WithCtx(ctx),
	)
	return
}

func (h HostEnvironment) callDiscordMemberUpdate(ctx context.Context, app store.AppStateProvider, data distype.MemberModifyRequest) (res *distype.MemberModifyResponse, err error) {
	err = app.Client().Request(
		rest.UpdateMember.Compile(nil, data.GuildID, data.UserID),
		data, &res, rest.WithCtx(ctx),
	)
	return
}

func (h HostEnvironment) callDiscordMemberUpdateOwn(ctx context.Context, app store.AppStateProvider, data distype.MemberModifyCurrentRequest) (res *distype.MemberModifyCurrentResponse, err error) {
	err = app.Client().Request(
		rest.UpdateCurrentMember.Compile(nil, data.GuildID),
		data, &res, rest.WithCtx(ctx),
	)
	return
}

func (h HostEnvironment) callDiscordMemberRoleAdd(ctx context.Context, app store.AppStateProvider, data distype.MemberRoleAddRequest) (*distype.MemberRoleAddResponse, error) {
	err := app.Client().Request(
		rest.AddMemberRole.Compile(nil, data.GuildID, data.UserID, data.RoleID),
		nil, nil, rest.WithCtx(ctx),
	)
	return &distype.MemberRoleAddResponse{}, err
}

func (h HostEnvironment) callDiscordMemberRoleRemove(ctx context.Context, app store.AppStateProvider, data distype.MemberRoleRemoveRequest) (*distype.MemberRoleRemoveResponse, error) {
	err := app.Client().Request(
		rest.RemoveMemberRole.Compile(nil, data.GuildID, data.UserID, data.RoleID),
		nil, nil, rest.WithCtx(ctx),
	)
	return &distype.MemberRoleRemoveResponse{}, err
}

func (h HostEnvironment) callDiscordMemberRemove(ctx context.Context, app store.AppStateProvider, data distype.MemberRemoveRequest) (*distype.MemberRemoveResponse, error) {
	err := app.Client().Request(
		rest.RemoveMember.Compile(nil, data.GuildID, data.UserID),
		nil, nil, rest.WithCtx(ctx),
	)
	return &distype.MemberRemoveResponse{}, err
}

func (h HostEnvironment) callDiscordMemberPruneCount(ctx context.Context, app store.AppStateProvider, data distype.MemberPruneCountRequest) (res *distype.MemberPruneCountResponse, err error) {
	params := make(map[string]any)
	params["days"] = data.Days
	if data.IncludeRoles != nil && len(data.IncludeRoles) != 0 {
		includeRoles := ""
		for _, role := range data.IncludeRoles {
			includeRoles += string(role) + ","
		}
		params["include_roles"] = includeRoles[:len(includeRoles)-1]
	}

	err = app.Client().Request(
		rest.GetGuildPruneCount.Compile(params, data.GuildID),
		nil, &res, rest.WithCtx(ctx),
	)
	return
}

func (h HostEnvironment) callDiscordMemberPruneBegin(ctx context.Context, app store.AppStateProvider, data distype.MemberPruneRequest) (res *distype.MemberPruneResponse, err error) {
	err = app.Client().Request(
		rest.BeginGuildPrune.Compile(nil, data.GuildID),
		data, &res, rest.WithCtx(ctx),
	)
	return
}

func (h HostEnvironment) callDiscordMessageList(ctx context.Context, app store.AppStateProvider, data distype.ChannelMessageListRequest) (res *distype.ChannelMessageListResponse, err error) {
	params := make(map[string]any)
	if data.Around != nil {
		params["around"] = *data.Around
	}
	if data.Before != nil {
		params["before"] = *data.Before
	}
	if data.After != nil {
		params["after"] = *data.After
	}
	if data.Limit != nil {
		params["limit"] = *data.Limit
	}

	err = app.Client().Request(
		rest.GetMessages.Compile(params, data.ChannelID),
		nil, &res, rest.WithCtx(ctx),
	)
	return
}

func (h HostEnvironment) callDiscordMessageGet(ctx context.Context, app store.AppStateProvider, data distype.MessageGetRequest) (res *distype.MessageGetResponse, err error) {
	err = app.Client().Request(
		rest.GetMessage.Compile(nil, data.ChannelID, data.MessageID),
		nil, &res, rest.WithCtx(ctx),
	)
	return
}

func (h HostEnvironment) callDiscordMessageCreate(ctx context.Context, app store.AppStateProvider, data distype.MessageCreateRequest) (res *distype.MessageCreateResponse, err error) {
	err = app.Client().Request(
		rest.CreateMessage.Compile(nil, data.ChannelID),
		data, &res, rest.WithCtx(ctx),
	)
	return
}

func (h HostEnvironment) callDiscordMessageUpdate(ctx context.Context, app store.AppStateProvider, data distype.MessageEditRequest) (res *distype.MessageEditResponse, err error) {
	err = app.Client().Request(
		rest.UpdateMessage.Compile(nil, data.ChannelID, data.MessageID),
		data, &res, rest.WithCtx(ctx),
	)
	return
}

func (h HostEnvironment) callDiscordMessageDelete(ctx context.Context, app store.AppStateProvider, data distype.MessageDeleteRequest) (*distype.MessageDeleteResponse, error) {
	err := app.Client().Request(
		rest.DeleteMessage.Compile(nil, data.ChannelID, data.MessageID),
		nil, nil, rest.WithCtx(ctx),
	)
	return &distype.MessageDeleteResponse{}, err
}

func (h HostEnvironment) callDiscordMessageDeleteBulk(ctx context.Context, app store.AppStateProvider, data distype.MessageBulkDeleteRequest) (*distype.MessageBulkDeleteResponse, error) {
	err := app.Client().Request(
		rest.BulkDeleteMessages.Compile(nil, data.ChannelID),
		data, nil, rest.WithCtx(ctx),
	)
	return &distype.MessageBulkDeleteResponse{}, err
}

func (h HostEnvironment) callDiscordMessageReactionCreate(ctx context.Context, app store.AppStateProvider, data distype.MessageReactionCreateRequest) (*distype.MessageReactionCreateResponse, error) {
	err := app.Client().Request(
		rest.AddReaction.Compile(nil, data.ChannelID, data.MessageID, data.Emoji),
		data, nil, rest.WithCtx(ctx),
	)
	return &distype.MessageReactionCreateResponse{}, err
}

func (h HostEnvironment) callDiscordMessageReactionDeleteOwn(ctx context.Context, app store.AppStateProvider, data distype.MessageReactionDeleteOwnRequest) (*distype.MessageReactionDeleteOwnResponse, error) {
	err := app.Client().Request(
		rest.RemoveOwnReaction.Compile(nil, data.ChannelID, data.MessageID, data.Emoji),
		nil, nil, rest.WithCtx(ctx),
	)
	return &distype.MessageReactionDeleteOwnResponse{}, err
}

func (h HostEnvironment) callDiscordMessageReactionDeleteUser(ctx context.Context, app store.AppStateProvider, data distype.MessageReactionDeleteRequest) (*distype.MessageReactionDeleteResponse, error) {
	err := app.Client().Request(
		rest.RemoveUserReaction.Compile(nil, data.ChannelID, data.MessageID, data.Emoji, data.UserID),
		nil, nil, rest.WithCtx(ctx),
	)
	return &distype.MessageReactionDeleteResponse{}, err
}

func (h HostEnvironment) callDiscordMessageReactionList(ctx context.Context, app store.AppStateProvider, data distype.MessageReactionListRequest) (res *distype.MessageReactionListResponse, err error) {
	params := make(map[string]any)
	if data.After != nil {
		params["after"] = *data.After
	}
	if data.Limit != nil {
		params["limit"] = *data.Limit
	}

	err = app.Client().Request(
		rest.GetReactions.Compile(params, data.ChannelID, data.MessageID, data.Emoji),
		nil, &res, rest.WithCtx(ctx),
	)
	return
}

func (h HostEnvironment) callDiscordMessageReactionDeleteAll(ctx context.Context, app store.AppStateProvider, data distype.MessageReactionDeleteAllRequest) (*distype.MessageReactionDeleteAllResponse, error) {
	err := app.Client().Request(
		rest.RemoveAllReactions.Compile(nil, data.ChannelID, data.MessageID),
		nil, nil, rest.WithCtx(ctx),
	)
	return &distype.MessageReactionDeleteAllResponse{}, err
}

func (h HostEnvironment) callDiscordMessageReactionDeleteEmoji(ctx context.Context, app store.AppStateProvider, data distype.MessageReactionDeleteEmojiRequest) (*distype.MessageReactionDeleteEmojiResponse, error) {
	err := app.Client().Request(
		rest.RemoveAllReactionsForEmoji.Compile(nil, data.ChannelID, data.MessageID, data.Emoji),
		nil, nil, rest.WithCtx(ctx),
	)
	return &distype.MessageReactionDeleteEmojiResponse{}, err
}

func (h HostEnvironment) callDiscordMessageGetPinned(ctx context.Context, app store.AppStateProvider, data distype.ChannelPinnedMessageListRequest) (res *distype.ChannelPinnedMessageListResponse, err error) {
	err = app.Client().Request(
		rest.GetPinnedMessages.Compile(nil, data.ChannelID),
		nil, &res, rest.WithCtx(ctx),
	)
	return
}

func (h HostEnvironment) callDiscordMessagePin(ctx context.Context, app store.AppStateProvider, data distype.MessagePinRequest) (*distype.MessagePinResponse, error) {
	err := app.Client().Request(
		rest.PinMessage.Compile(nil, data.ChannelID, data.MessageID),
		nil, nil, rest.WithCtx(ctx),
	)
	return &distype.MessagePinResponse{}, err
}

func (h HostEnvironment) callDiscordMessageUnpin(ctx context.Context, app store.AppStateProvider, data distype.MessageUnpinRequest) (*distype.MessageUnpinResponse, error) {
	err := app.Client().Request(
		rest.UnpinMessage.Compile(nil, data.ChannelID, data.MessageID),
		nil, nil, rest.WithCtx(ctx),
	)
	return &distype.MessageUnpinResponse{}, err
}

func (h HostEnvironment) callDiscordRoleList(ctx context.Context, app store.AppStateProvider, data distype.GuildRoleListRequest) (res *distype.GuildRoleListResponse, err error) {
	err = app.Client().Request(
		rest.GetRoles.Compile(nil, data.GuildID),
		nil, &res, rest.WithCtx(ctx),
	)
	return
}

func (h HostEnvironment) callDiscordRoleCreate(ctx context.Context, app store.AppStateProvider, data distype.RoleCreateRequest) (res *distype.RoleCreateResponse, err error) {
	err = app.Client().Request(
		rest.CreateRole.Compile(nil, data.GuildID),
		data, &res, rest.WithCtx(ctx),
	)
	return
}

func (h HostEnvironment) callDiscordRoleUpdate(ctx context.Context, app store.AppStateProvider, data distype.RoleModifyRequest) (res *distype.RoleModifyResponse, err error) {
	err = app.Client().Request(
		rest.UpdateRole.Compile(nil, data.GuildID, data.RoleID),
		data, &res, rest.WithCtx(ctx),
	)
	return
}

func (h HostEnvironment) callDiscordRoleUpdatePositions(ctx context.Context, app store.AppStateProvider, data distype.RolePositionsModifyRequest) (*distype.RolePositionsModifyResponse, error) {
	err := app.Client().Request(
		rest.UpdateRolePositions.Compile(nil, data.GuildID),
		data.Entries, nil, rest.WithCtx(ctx),
	)
	return &distype.RolePositionsModifyResponse{}, err
}

func (h HostEnvironment) callDiscordRoleDelete(ctx context.Context, app store.AppStateProvider, data distype.RoleDeleteRequest) (*distype.RoleDeleteResponse, error) {
	err := app.Client().Request(
		rest.DeleteRole.Compile(nil, data.GuildID, data.RoleID),
		nil, nil, rest.WithCtx(ctx),
	)
	return &distype.RoleDeleteResponse{}, err
}

func (h HostEnvironment) callDiscordScheduledEventList(ctx context.Context, app store.AppStateProvider, data distype.GuildScheduledEventListRequest) (res *distype.GuildScheduledEventListResponse, err error) {
	params := make(map[string]any)
	if data.WithUserCount != nil {
		params["with_user_count"] = *data.WithUserCount
	}

	err = app.Client().Request(
		rest.GetGuildScheduledEvents.Compile(params, data.GuildID),
		nil, &res, rest.WithCtx(ctx),
	)
	return
}

func (h HostEnvironment) callDiscordScheduledEventCreate(ctx context.Context, app store.AppStateProvider, data distype.ScheduledEventCreateRequest) (res *distype.ScheduledEventCreateResponse, err error) {
	err = app.Client().Request(
		rest.CreateGuildScheduledEvent.Compile(nil, data.GuildID),
		data, &res, rest.WithCtx(ctx),
	)
	return
}

func (h HostEnvironment) callDiscordScheduledEventGet(ctx context.Context, app store.AppStateProvider, data distype.ScheduledEventGetRequest) (res *distype.ScheduledEventGetResponse, err error) {
	err = app.Client().Request(
		rest.GetGuildScheduledEvent.Compile(nil, data.GuildID, data.GuildScheduledEventID),
		nil, &res, rest.WithCtx(ctx),
	)
	return
}

func (h HostEnvironment) callDiscordScheduledEventUpdate(ctx context.Context, app store.AppStateProvider, data distype.ScheduledEventModifyRequest) (res *distype.ScheduledEventModifyResponse, err error) {
	err = app.Client().Request(
		rest.UpdateGuildScheduledEvent.Compile(nil, data.GuildID, data.GuildScheduledEventID),
		data, &res, rest.WithCtx(ctx),
	)
	return
}

func (h HostEnvironment) callDiscordScheduledEventDelete(ctx context.Context, app store.AppStateProvider, data distype.ScheduledEventDeleteRequest) (*distype.ScheduledEventDeleteResponse, error) {
	err := app.Client().Request(
		rest.DeleteGuildScheduledEvent.Compile(nil, data.GuildID, data.GuildScheduledEventID),
		nil, nil, rest.WithCtx(ctx),
	)
	return &distype.ScheduledEventDeleteResponse{}, err
}

func (h HostEnvironment) callDiscordScheduledEventUserList(ctx context.Context, app store.AppStateProvider, data distype.ScheduledEventUserListRequest) (res *distype.ScheduledEventUserListResponse, err error) {
	params := make(map[string]any)
	if data.After != nil {
		params["after"] = *data.After
	}
	if data.Before != nil {
		params["before"] = *data.Before
	}
	if data.Limit != nil {
		params["limit"] = *data.Limit
	}
	if data.WithMember != nil {
		params["with_member"] = *data.WithMember
	}

	err = app.Client().Request(
		rest.GetGuildScheduledEventUsers.Compile(params, data.GuildID, data.GuildScheduledEventID),
		nil, &res, rest.WithCtx(ctx),
	)
	return
}

func (h HostEnvironment) callDiscordStageInstanceCreate(ctx context.Context, app store.AppStateProvider, data distype.StageInstanceCreateRequest) (res *distype.StageInstanceCreateResponse, err error) {
	err = app.Client().Request(
		rest.CreateStageInstance.Compile(nil, data.ChannelID),
		data, &res, rest.WithCtx(ctx),
	)
	return
}

func (h HostEnvironment) callDiscordStageInstanceGet(ctx context.Context, app store.AppStateProvider, data distype.StageInstanceGetRequest) (res *distype.StageInstanceGetResponse, err error) {
	err = app.Client().Request(
		rest.GetStageInstance.Compile(nil, data.ChannelID),
		nil, &res, rest.WithCtx(ctx),
	)
	return
}

func (h HostEnvironment) callDiscordStageInstanceUpdate(ctx context.Context, app store.AppStateProvider, data distype.StageInstanceModifyRequest) (res *distype.StageInstanceModifyResponse, err error) {
	err = app.Client().Request(
		rest.UpdateStageInstance.Compile(nil, data.ChannelID),
		data, &res, rest.WithCtx(ctx),
	)
	return
}

func (h HostEnvironment) callDiscordStageInstanceDelete(ctx context.Context, app store.AppStateProvider, data distype.StageInstanceDeleteRequest) (*distype.StageInstanceDeleteResponse, error) {
	err := app.Client().Request(
		rest.DeleteStageInstance.Compile(nil, data.ChannelID),
		nil, nil, rest.WithCtx(ctx),
	)
	return &distype.StageInstanceDeleteResponse{}, err
}

func (h HostEnvironment) callDiscordStickerList(ctx context.Context, app store.AppStateProvider, data distype.GuildStickerListRequest) (res *distype.GuildStickerListResponse, err error) {
	err = app.Client().Request(
		rest.GetGuildStickers.Compile(nil, data.GuildID),
		nil, &res, rest.WithCtx(ctx),
	)
	return
}

func (h HostEnvironment) callDiscordStickerGet(ctx context.Context, app store.AppStateProvider, data distype.StickerGetRequest) (res *distype.StickerGetResponse, err error) {
	err = app.Client().Request(
		rest.GetSticker.Compile(nil, data.GuildID, data.StickerID),
		nil, &res, rest.WithCtx(ctx),
	)
	return
}

func (h HostEnvironment) callDiscordStickerCreate(ctx context.Context, app store.AppStateProvider, data distype.StickerCreateRequest) (res *distype.StickerCreateResponse, err error) {
	err = app.Client().Request(
		rest.CreateGuildSticker.Compile(nil, data.GuildID),
		data, &res, rest.WithCtx(ctx),
	)
	return
}

func (h HostEnvironment) callDiscordStickerUpdate(ctx context.Context, app store.AppStateProvider, data distype.StickerModifyRequest) (res *distype.StickerModifyResponse, err error) {
	err = app.Client().Request(
		rest.UpdateGuildSticker.Compile(nil, data.GuildID, data.StickerID),
		data, &res, rest.WithCtx(ctx),
	)
	return
}

func (h HostEnvironment) callDiscordStickerDelete(ctx context.Context, app store.AppStateProvider, data distype.StickerDeleteRequest) (*distype.StickerDeleteResponse, error) {
	err := app.Client().Request(
		rest.DeleteGuildSticker.Compile(nil, data.GuildID, data.StickerID),
		nil, nil, rest.WithCtx(ctx),
	)
	return &distype.StickerDeleteResponse{}, err
}

func (h HostEnvironment) callDiscordUserGet(ctx context.Context, app store.AppStateProvider, data distype.UserGetRequest) (res *distype.UserGetResponse, err error) {
	err = app.Client().Request(
		rest.GetUser.Compile(nil, data.UserID),
		nil, &res, rest.WithCtx(ctx),
	)
	return
}

func (h HostEnvironment) callDiscordWebhookGet(ctx context.Context, app store.AppStateProvider, data distype.WebhookGetRequest) (res *distype.WebhookGetResponse, err error) {
	err = app.Client().Request(
		rest.GetWebhook.Compile(nil, data.WebhookID),
		nil, &res, rest.WithCtx(ctx),
	)
	return
}

func (h HostEnvironment) callDiscordWebhookListForChannel(ctx context.Context, app store.AppStateProvider, data distype.ChannelWebhookListRequest) (res *distype.ChannelWebhookListResponse, err error) {
	err = app.Client().Request(
		rest.GetChannelWebhooks.Compile(nil, data.ChannelID),
		nil, &res, rest.WithCtx(ctx),
	)
	return
}

func (h HostEnvironment) callDiscordWebhookListForGuild(ctx context.Context, app store.AppStateProvider, data distype.GuildWebhookListRequest) (res *distype.GuildWebhookListResponse, err error) {
	err = app.Client().Request(
		rest.GetGuildWebhooks.Compile(nil, data.GuildID),
		nil, &res, rest.WithCtx(ctx),
	)
	return
}

func (h HostEnvironment) callDiscordWebhookCreate(ctx context.Context, app store.AppStateProvider, data distype.WebhookCreateRequest) (res *distype.WebhookCreateResponse, err error) {
	err = app.Client().Request(
		rest.CreateWebhook.Compile(nil, data.ChannelID),
		data, &res, rest.WithCtx(ctx),
	)
	return
}

func (h HostEnvironment) callDiscordWebhookUpdate(ctx context.Context, app store.AppStateProvider, data distype.WebhookModifyRequest) (res *distype.WebhookModifyResponse, err error) {
	err = app.Client().Request(
		rest.UpdateWebhook.Compile(nil, data.WebhookID),
		data, &res, rest.WithCtx(ctx),
	)
	return
}

func (h HostEnvironment) callDiscordWebhookDelete(ctx context.Context, app store.AppStateProvider, data distype.WebhookDeleteRequest) (*distype.WebhookDeleteResponse, error) {
	err := app.Client().Request(
		rest.DeleteWebhook.Compile(nil, data.WebhookID),
		nil, nil, rest.WithCtx(ctx),
	)
	return &distype.WebhookDeleteResponse{}, err
}

func (h HostEnvironment) callDiscordWebhookGetWithToken(ctx context.Context, app store.AppStateProvider, data distype.WebhookGetWithTokenRequest) (res *distype.WebhookGetWithTokenResponse, err error) {
	err = app.Client().Request(
		rest.GetWebhookWithToken.Compile(nil, data.WebhookID, data.WebhookToken),
		nil, &res, rest.WithCtx(ctx),
	)
	return
}

func (h HostEnvironment) callDiscordWebhookUpdateWithToken(ctx context.Context, app store.AppStateProvider, data distype.WebhookModifyWithTokenRequest) (res *distype.WebhookModifyWithTokenResponse, err error) {
	err = app.Client().Request(
		rest.UpdateWebhookWithToken.Compile(nil, data.WebhookID, data.WebhookToken),
		data, &res, rest.WithCtx(ctx),
	)
	return
}

func (h HostEnvironment) callDiscordWebhookDeleteWithToken(ctx context.Context, app store.AppStateProvider, data distype.WebhookDeleteWithTokenRequest) (*distype.WebhookDeleteWithTokenResponse, error) {
	err := app.Client().Request(
		rest.DeleteWebhookWithToken.Compile(nil, data.WebhookID, data.WebhookToken),
		nil, nil, rest.WithCtx(ctx),
	)
	return &distype.WebhookDeleteWithTokenResponse{}, err
}

func (h HostEnvironment) callDiscordWebhookExecute(ctx context.Context, app store.AppStateProvider, data distype.WebhookExecuteRequest) (res *distype.WebhookExecuteResponse, err error) {
	err = app.Client().Request(
		rest.CreateWebhookMessage.Compile(nil, data.WebhookID, data.WebhookToken),
		data, &res, rest.WithCtx(ctx),
	)
	return
}
