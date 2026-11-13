// What the code of Kite's Run JavaScript block can use besides JavaScript
// itself. See https://docs.kite.onl/reference/blocks/actions/action_javascript_run

/** The values of the block's inputs, by name. */
declare const input: Record<string, any>;

/** Sends a request to a URL with http.request. Costs 3 credits. */
declare function fetch(url: string, init?: {
  method?: "GET" | "POST" | "PUT" | "PATCH" | "DELETE" | "get" | "post" | "put" | "patch" | "delete";
  headers?: Record<string, string> | [string, string][];
  body?: string;
}): Promise<{
  ok: boolean;
  status: number;
  statusText: string;
  url: string;
  headers: { get(name: string): string | null; has(name: string): boolean };
  text(): Promise<string>;
  json(): Promise<any>;
}>;

/** Formats a date with tokens like "dddd HH:mm" in a time zone like "Europe/Berlin". */
declare function formatDate(date?: Date | number | string, format?: string, timeZone?: string): string;

declare const discord: {
  /** Action guild join request, PATCH /guilds/{guild_id}/requests/{request_id}. Costs 1 credit. */
  actionGuildJoinRequest(arg: {
    guild_id: string;
    request_id: string;
    /** The JSON body of the request. */ body?: { action?: discord.GuildJoinRequestApplicationStatus; rejection_reason?: string | null };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<discord.GuildJoinRequestResponse>;
  /** Add group dm user, PUT /channels/{channel_id}/recipients/{user_id}. Costs 1 credit. */
  addGroupDmUser(arg: {
    channel_id: string;
    user_id: string;
    /** The JSON body of the request. */ body?: { access_token?: string | null; nick?: string | null };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<discord.PrivateChannelResponse | discord.PrivateGroupChannelResponse | null>;
  /** Add guild member, PUT /guilds/{guild_id}/members/{user_id}. Costs 1 credit. */
  addGuildMember(arg: {
    guild_id: string;
    user_id: string;
    /** The JSON body of the request. */ body: discord.BotAddGuildMemberRequest;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<discord.GuildMemberResponse | null>;
  /** Add guild member role, PUT /guilds/{guild_id}/members/{user_id}/roles/{role_id}. Costs 1 credit. */
  addGuildMemberRole(arg: {
    guild_id: string;
    user_id: string;
    role_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<null>;
  /** Add invite target user, PUT /invites/{code}/target-users/{user_id}. Costs 1 credit. */
  addInviteTargetUser(arg: {
    code: string;
    user_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<null>;
  /** Add lobby member, PUT /lobbies/{lobby_id}/members/{user_id}. Costs 1 credit. */
  addLobbyMember(arg: {
    lobby_id: string;
    user_id: string;
    /** The JSON body of the request. */ body?: { metadata?: Record<string, string> | null; flags?: 1 | null; additional_name?: string | null };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<discord.LobbyMemberResponse>;
  /** Add my message reaction, PUT /channels/{channel_id}/messages/{message_id}/reactions/{emoji_name}/@me. Costs 1 credit. */
  addMyMessageReaction(arg: {
    channel_id: string;
    message_id: string;
    emoji_name: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<null>;
  /** Add thread member, PUT /channels/{channel_id}/thread-members/{user_id}. Costs 1 credit. */
  addThreadMember(arg: {
    channel_id: string;
    user_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<null>;
  /** Applications get activity instance, GET /applications/{application_id}/activity-instances/{instance_id}. Costs nothing. */
  applicationsGetActivityInstance(arg: {
    application_id: string;
    instance_id: string;
  }): Promise<discord.EmbeddedActivityInstance>;
  /** Ban user from guild, PUT /guilds/{guild_id}/bans/{user_id}. Costs 1 credit. */
  banUserFromGuild(arg: {
    guild_id: string;
    user_id: string;
    /** The JSON body of the request. */ body?: discord.BanUserFromGuildRequest;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<null>;
  /** Bot partner sdk token, POST /partner-sdk/token/bot. Costs 1 credit. */
  botPartnerSdkToken(arg: {
    /** The JSON body of the request. */ body: { provisional_user_id?: string | null; external_user_id: string; preferred_global_name?: string | null };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<discord.ProvisionalTokenResponse>;
  /** Bot partner sdk unmerge provisional account, POST /partner-sdk/provisional-accounts/unmerge/bot. Costs 1 credit. */
  botPartnerSdkUnmergeProvisionalAccount(arg: {
    /** The JSON body of the request. */ body: { external_user_id: string };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<null>;
  /** Bulk add invite target users, POST /invites/{code}/target-users/bulk-add. Costs 1 credit. */
  bulkAddInviteTargetUsers(arg: {
    code: string;
    /** The JSON body of the request. */ body: { user_ids: string[] };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<null>;
  /** Bulk delete messages, POST /channels/{channel_id}/messages/bulk-delete. Costs 1 credit. */
  bulkDeleteMessages(arg: {
    channel_id: string;
    /** The JSON body of the request. */ body: { messages: string[] };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<null>;
  /** Bulk remove invite target users, POST /invites/{code}/target-users/bulk-delete. Costs 1 credit. */
  bulkRemoveInviteTargetUsers(arg: {
    code: string;
    /** The JSON body of the request. */ body: { user_ids: string[] };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<null>;
  /** Bulk update guild channels, PATCH /guilds/{guild_id}/channels. Costs 1 credit. */
  bulkUpdateGuildChannels(arg: {
    guild_id: string;
    /** The JSON body of the request. */ body?: { id?: string | null; position?: number | null; parent_id?: string | null; lock_permissions?: boolean | null }[];
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<null>;
  /** Bulk update guild roles, PATCH /guilds/{guild_id}/roles. Costs 1 credit. */
  bulkUpdateGuildRoles(arg: {
    guild_id: string;
    /** The JSON body of the request. */ body?: discord.UpdateRolePositionsRequest[];
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<discord.GuildRoleResponse[]>;
  /** Bulk update lobby members, POST /lobbies/{lobby_id}/members/bulk. Costs 1 credit. */
  bulkUpdateLobbyMembers(arg: {
    lobby_id: string;
    /** The JSON body of the request. */ body?: discord.BulkLobbyMemberRequest[] | null;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<discord.LobbyMemberResponse[]>;
  /** Count guild scheduled event users, GET /guilds/{guild_id}/scheduled-events/{guild_scheduled_event_id}/users/counts. Costs nothing. */
  countGuildScheduledEventUsers(arg: {
    guild_id: string;
    guild_scheduled_event_id: string;
    guild_scheduled_event_exception_ids?: (string | number)[];
  }): Promise<discord.ScheduledEventUserCountResponse>;
  /** Create auto moderation rule, POST /guilds/{guild_id}/auto-moderation/rules. Costs 1 credit. */
  createAutoModerationRule(arg: {
    guild_id: string;
    /** The JSON body of the request. */ body: discord.DefaultKeywordListUpsertRequest | discord.KeywordUpsertRequest | discord.MLSpamUpsertRequest | discord.MentionSpamUpsertRequest | discord.UserProfileUpsertRequest;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<discord.DefaultKeywordRuleResponse | discord.KeywordRuleResponse | discord.MLSpamRuleResponse | discord.MentionSpamRuleResponse | discord.UserProfileRuleResponse>;
  /** Create channel invite, POST /channels/{channel_id}/invites. Costs 1 credit. */
  createChannelInvite(arg: {
    channel_id: string;
    /** The JSON body of the request. */ body?: discord.CreateGroupDMInviteRequest | discord.CreateGuildInviteRequest;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<discord.FriendInviteResponse | discord.GroupDMInviteResponse | discord.GuildInviteResponse | null>;
  /** Create dm, POST /users/@me/channels. Costs 1 credit. */
  createDm(arg?: {
    /** The JSON body of the request. */ body?: discord.CreatePrivateChannelRequest;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<discord.PrivateChannelResponse | discord.PrivateGroupChannelResponse>;
  /** Create guild channel, POST /guilds/{guild_id}/channels. Costs 1 credit. */
  createGuildChannel(arg: {
    guild_id: string;
    /** The JSON body of the request. */ body: discord.CreateGuildChannelRequest;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<discord.GuildChannelResponse>;
  /** Create guild emoji, POST /guilds/{guild_id}/emojis. Costs 1 credit. */
  createGuildEmoji(arg: {
    guild_id: string;
    /** The JSON body of the request. */ body: { name: string; image: string; roles?: (string | null)[] | null };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<discord.EmojiResponse>;
  /** Create guild role, POST /guilds/{guild_id}/roles. Costs 1 credit. */
  createGuildRole(arg: {
    guild_id: string;
    /** The JSON body of the request. */ body?: discord.CreateRoleRequest;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<discord.GuildRoleResponse>;
  /** Create guild scheduled event, POST /guilds/{guild_id}/scheduled-events. Costs 1 credit. */
  createGuildScheduledEvent(arg: {
    guild_id: string;
    /** The JSON body of the request. */ body: discord.ExternalScheduledEventCreateRequest | discord.StageScheduledEventCreateRequest | discord.VoiceScheduledEventCreateRequest;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<discord.ExternalScheduledEventResponse | discord.StageScheduledEventResponse | discord.VoiceScheduledEventResponse>;
  /** Create guild scheduled event exception, POST /guilds/{guild_id}/scheduled-events/{guild_scheduled_event_id}/exceptions. Costs 1 credit. */
  createGuildScheduledEventException(arg: {
    guild_id: string;
    guild_scheduled_event_id: string;
    /** The JSON body of the request. */ body: discord.GuildScheduledEventExceptionCreateRequest;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<discord.GuildScheduledEventExceptionResponse>;
  /** Create guild soundboard sound, POST /guilds/{guild_id}/soundboard-sounds. Costs 1 credit. */
  createGuildSoundboardSound(arg: {
    guild_id: string;
    /** The JSON body of the request. */ body: discord.SoundboardCreateRequest;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<discord.SoundboardSoundResponse>;
  /** Create guild template, POST /guilds/{guild_id}/templates. Costs 1 credit. */
  createGuildTemplate(arg: {
    guild_id: string;
    /** The JSON body of the request. */ body: { name: string; description?: string | null };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<discord.GuildTemplateResponse>;
  /** Create linked lobby guild invite for self, POST /lobbies/{lobby_id}/members/@me/invites. Costs 1 credit. */
  createLinkedLobbyGuildInviteForSelf(arg: {
    lobby_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<discord.LobbyGuildInviteResponse>;
  /** Create linked lobby guild invite for user, POST /lobbies/{lobby_id}/members/{user_id}/invites. Costs 1 credit. */
  createLinkedLobbyGuildInviteForUser(arg: {
    lobby_id: string;
    user_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<discord.LobbyGuildInviteResponse>;
  /** Create lobby, POST /lobbies. Costs 1 credit. */
  createLobby(arg?: {
    /** The JSON body of the request. */ body?: { idle_timeout_seconds?: number | null; members?: discord.LobbyMemberRequest[] | null; metadata?: Record<string, string> | null; flags?: 1 | null; override_event_webhooks_url?: string | null };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<discord.LobbyResponse>;
  /** Create lobby message, POST /lobbies/{lobby_id}/messages. Costs 1 credit. */
  createLobbyMessage(arg: {
    lobby_id: string;
    /** The JSON body of the request. */ body?: discord.SDKMessageRequest;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<discord.LobbyMessageResponse>;
  /** Create message, POST /channels/{channel_id}/messages. Costs 1 credit. */
  createMessage(arg: {
    channel_id: string;
    /** The JSON body of the request. */ body?: discord.MessageCreateRequest;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<discord.MessageResponse>;
  /** Create or join lobby, PUT /lobbies. Costs 1 credit. */
  createOrJoinLobby(arg: {
    /** The JSON body of the request. */ body: { idle_timeout_seconds?: number | null; lobby_metadata?: Record<string, string> | null; member_metadata?: Record<string, string> | null; secret: string; flags?: 1 | null };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<discord.LobbyResponse>;
  /** Create pin, PUT /channels/{channel_id}/messages/pins/{message_id}. Costs 1 credit. */
  createPin(arg: {
    channel_id: string;
    message_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<null>;
  /** Create stage instance, POST /stage-instances. Costs 1 credit. */
  createStageInstance(arg: {
    /** The JSON body of the request. */ body: { topic: string; channel_id: string; privacy_level?: discord.StageInstancesPrivacyLevels | null; guild_scheduled_event_id?: string | null; send_start_notification?: boolean | null };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<discord.StageInstanceResponse>;
  /** Create thread, POST /channels/{channel_id}/threads. Costs 1 credit. */
  createThread(arg: {
    channel_id: string;
    /** The JSON body of the request. */ body: discord.CreateForumThreadRequest | discord.CreateTextThreadWithoutMessageRequest;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<discord.CreatedThreadResponse>;
  /** Create thread from message, POST /channels/{channel_id}/messages/{message_id}/threads. Costs 1 credit. */
  createThreadFromMessage(arg: {
    channel_id: string;
    message_id: string;
    /** The JSON body of the request. */ body: discord.CreateTextThreadWithMessageRequest;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<discord.ThreadResponse>;
  /** Create webhook, POST /channels/{channel_id}/webhooks. Costs 1 credit. */
  createWebhook(arg: {
    channel_id: string;
    /** The JSON body of the request. */ body: { name: string; avatar?: string | null };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<discord.GuildIncomingWebhookResponse>;
  /** Crosspost message, POST /channels/{channel_id}/messages/{message_id}/crosspost. Costs 1 credit. */
  crosspostMessage(arg: {
    channel_id: string;
    message_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<discord.MessageResponse>;
  /** Delete all message reactions, DELETE /channels/{channel_id}/messages/{message_id}/reactions. Costs 1 credit. */
  deleteAllMessageReactions(arg: {
    channel_id: string;
    message_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<null>;
  /** Delete all message reactions by emoji, DELETE /channels/{channel_id}/messages/{message_id}/reactions/{emoji_name}. Costs 1 credit. */
  deleteAllMessageReactionsByEmoji(arg: {
    channel_id: string;
    message_id: string;
    emoji_name: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<null>;
  /** Delete auto moderation rule, DELETE /guilds/{guild_id}/auto-moderation/rules/{rule_id}. Costs 1 credit. */
  deleteAutoModerationRule(arg: {
    guild_id: string;
    rule_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<null>;
  /** Delete channel, DELETE /channels/{channel_id}. Costs 1 credit. */
  deleteChannel(arg: {
    channel_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<discord.GuildChannelResponse | discord.PrivateChannelResponse | discord.PrivateGroupChannelResponse | discord.ThreadResponse>;
  /** Delete channel permission overwrite, DELETE /channels/{channel_id}/permissions/{overwrite_id}. Costs 1 credit. */
  deleteChannelPermissionOverwrite(arg: {
    channel_id: string;
    overwrite_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<null>;
  /** Delete group dm user, DELETE /channels/{channel_id}/recipients/{user_id}. Costs 1 credit. */
  deleteGroupDmUser(arg: {
    channel_id: string;
    user_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<null>;
  /** Delete guild emoji, DELETE /guilds/{guild_id}/emojis/{emoji_id}. Costs 1 credit. */
  deleteGuildEmoji(arg: {
    guild_id: string;
    emoji_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<null>;
  /** Delete guild integration, DELETE /guilds/{guild_id}/integrations/{integration_id}. Costs 1 credit. */
  deleteGuildIntegration(arg: {
    guild_id: string;
    integration_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<null>;
  /** Delete guild member, DELETE /guilds/{guild_id}/members/{user_id}. Costs 1 credit. */
  deleteGuildMember(arg: {
    guild_id: string;
    user_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<null>;
  /** Delete guild member role, DELETE /guilds/{guild_id}/members/{user_id}/roles/{role_id}. Costs 1 credit. */
  deleteGuildMemberRole(arg: {
    guild_id: string;
    user_id: string;
    role_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<null>;
  /** Delete guild role, DELETE /guilds/{guild_id}/roles/{role_id}. Costs 1 credit. */
  deleteGuildRole(arg: {
    guild_id: string;
    role_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<null>;
  /** Delete guild scheduled event, DELETE /guilds/{guild_id}/scheduled-events/{guild_scheduled_event_id}. Costs 1 credit. */
  deleteGuildScheduledEvent(arg: {
    guild_id: string;
    guild_scheduled_event_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<null>;
  /** Delete guild scheduled event exception, DELETE /guilds/{guild_id}/scheduled-events/{guild_scheduled_event_id}/exceptions/{exception_id}. Costs 1 credit. */
  deleteGuildScheduledEventException(arg: {
    guild_id: string;
    guild_scheduled_event_id: string;
    exception_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<null>;
  /** Delete guild soundboard sound, DELETE /guilds/{guild_id}/soundboard-sounds/{sound_id}. Costs 1 credit. */
  deleteGuildSoundboardSound(arg: {
    guild_id: string;
    sound_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<null>;
  /** Delete guild sticker, DELETE /guilds/{guild_id}/stickers/{sticker_id}. Costs 1 credit. */
  deleteGuildSticker(arg: {
    guild_id: string;
    sticker_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<null>;
  /** Delete guild template, DELETE /guilds/{guild_id}/templates/{code}. Costs 1 credit. */
  deleteGuildTemplate(arg: {
    guild_id: string;
    code: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<discord.GuildTemplateResponse>;
  /** Delete lobby, DELETE /lobbies/{lobby_id}. Costs 1 credit. */
  deleteLobby(arg: {
    lobby_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<null>;
  /** Delete lobby member, DELETE /lobbies/{lobby_id}/members/{user_id}. Costs 1 credit. */
  deleteLobbyMember(arg: {
    lobby_id: string;
    user_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<null>;
  /** Delete message, DELETE /channels/{channel_id}/messages/{message_id}. Costs 1 credit. */
  deleteMessage(arg: {
    channel_id: string;
    message_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<null>;
  /** Delete my message reaction, DELETE /channels/{channel_id}/messages/{message_id}/reactions/{emoji_name}/@me. Costs 1 credit. */
  deleteMyMessageReaction(arg: {
    channel_id: string;
    message_id: string;
    emoji_name: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<null>;
  /** Delete pin, DELETE /channels/{channel_id}/messages/pins/{message_id}. Costs 1 credit. */
  deletePin(arg: {
    channel_id: string;
    message_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<null>;
  /** Delete stage instance, DELETE /stage-instances/{channel_id}. Costs 1 credit. */
  deleteStageInstance(arg: {
    channel_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<null>;
  /** Delete thread member, DELETE /channels/{channel_id}/thread-members/{user_id}. Costs 1 credit. */
  deleteThreadMember(arg: {
    channel_id: string;
    user_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<null>;
  /** Delete user message reaction, DELETE /channels/{channel_id}/messages/{message_id}/reactions/{emoji_name}/{user_id}. Costs 1 credit. */
  deleteUserMessageReaction(arg: {
    channel_id: string;
    message_id: string;
    emoji_name: string;
    user_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<null>;
  /** Delete webhook, DELETE /webhooks/{webhook_id}. Costs 1 credit. */
  deleteWebhook(arg: {
    webhook_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<null>;
  /** Deprecated create pin, PUT /channels/{channel_id}/pins/{message_id}. Costs 1 credit. */
  deprecatedCreatePin(arg: {
    channel_id: string;
    message_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<null>;
  /** Deprecated delete pin, DELETE /channels/{channel_id}/pins/{message_id}. Costs 1 credit. */
  deprecatedDeletePin(arg: {
    channel_id: string;
    message_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<null>;
  /** Deprecated list pins, GET /channels/{channel_id}/pins. Costs nothing. */
  deprecatedListPins(arg: {
    channel_id: string;
  }): Promise<discord.MessageResponse[]>;
  /** Edit lobby, PATCH /lobbies/{lobby_id}. Costs 1 credit. */
  editLobby(arg: {
    lobby_id: string;
    /** The JSON body of the request. */ body?: { idle_timeout_seconds?: number | null; metadata?: Record<string, string> | null; members?: discord.LobbyMemberRequest[] | null; flags?: 1 | null; override_event_webhooks_url?: string | null };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<discord.LobbyResponse>;
  /** Edit lobby channel link, PATCH /lobbies/{lobby_id}/channel-linking. Costs 1 credit. */
  editLobbyChannelLink(arg: {
    lobby_id: string;
    /** The JSON body of the request. */ body?: { channel_id?: string | null };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<discord.LobbyResponse>;
  /** Follow channel, POST /channels/{channel_id}/followers. Costs 1 credit. */
  followChannel(arg: {
    channel_id: string;
    /** The JSON body of the request. */ body: { webhook_channel_id: string };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<discord.ChannelFollowerResponse>;
  /** Get active guild threads, GET /guilds/{guild_id}/threads/active. Costs nothing. */
  getActiveGuildThreads(arg: {
    guild_id: string;
  }): Promise<discord.ThreadsResponse>;
  /** Get answer voters, GET /channels/{channel_id}/polls/{message_id}/answers/{answer_id}. Costs nothing. */
  getAnswerVoters(arg: {
    channel_id: string;
    message_id: string;
    answer_id: number;
    after?: string;
    limit?: number;
  }): Promise<discord.PollAnswerDetailsResponse>;
  /** Get application, GET /applications/{application_id}. Costs nothing. */
  getApplication(arg: {
    application_id: string;
  }): Promise<discord.PrivateApplicationResponse>;
  /** Get application command, GET /applications/{application_id}/commands/{command_id}. Costs nothing. */
  getApplicationCommand(arg: {
    application_id: string;
    command_id: string;
  }): Promise<discord.ApplicationCommandResponse>;
  /** Get application emoji, GET /applications/{application_id}/emojis/{emoji_id}. Costs nothing. */
  getApplicationEmoji(arg: {
    application_id: string;
    emoji_id: string;
  }): Promise<discord.EmojiResponse>;
  /** Get application role connections metadata, GET /applications/{application_id}/role-connections/metadata. Costs nothing. */
  getApplicationRoleConnectionsMetadata(arg: {
    application_id: string;
  }): Promise<discord.ApplicationRoleConnectionsMetadataItemResponse[]>;
  /** Get auto moderation rule, GET /guilds/{guild_id}/auto-moderation/rules/{rule_id}. Costs nothing. */
  getAutoModerationRule(arg: {
    guild_id: string;
    rule_id: string;
  }): Promise<discord.DefaultKeywordRuleResponse | discord.KeywordRuleResponse | discord.MLSpamRuleResponse | discord.MentionSpamRuleResponse | discord.UserProfileRuleResponse>;
  /** Get bot gateway, GET /gateway/bot. Costs nothing. */
  getBotGateway(): Promise<discord.GatewayBotResponse>;
  /** Get channel, GET /channels/{channel_id}. Costs nothing. */
  getChannel(arg: {
    channel_id: string;
  }): Promise<discord.GuildChannelResponse | discord.PrivateChannelResponse | discord.PrivateGroupChannelResponse | discord.ThreadResponse>;
  /** Get entitlement, GET /applications/{application_id}/entitlements/{entitlement_id}. Costs nothing. */
  getEntitlement(arg: {
    application_id: string;
    entitlement_id: string;
  }): Promise<discord.EntitlementResponse>;
  /** Get entitlements, GET /applications/{application_id}/entitlements. Costs nothing. */
  getEntitlements(arg: {
    application_id: string;
    user_id?: string;
    sku_ids?: string;
    guild_id?: string;
    before?: string;
    after?: string;
    limit?: number;
    exclude_ended?: boolean;
    exclude_deleted?: boolean;
    only_active?: boolean;
  }): Promise<discord.EntitlementResponse[]>;
  /** Get gateway, GET /gateway. Costs nothing. */
  getGateway(): Promise<discord.GatewayResponse>;
  /** Get guild, GET /guilds/{guild_id}. Costs nothing. */
  getGuild(arg: {
    guild_id: string;
    with_counts?: boolean;
  }): Promise<discord.GuildWithCountsResponse>;
  /** Get guild application command, GET /applications/{application_id}/guilds/{guild_id}/commands/{command_id}. Costs nothing. */
  getGuildApplicationCommand(arg: {
    application_id: string;
    guild_id: string;
    command_id: string;
  }): Promise<discord.ApplicationCommandResponse>;
  /** Get guild application command permissions, GET /applications/{application_id}/guilds/{guild_id}/commands/{command_id}/permissions. Costs nothing. */
  getGuildApplicationCommandPermissions(arg: {
    application_id: string;
    guild_id: string;
    command_id: string;
  }): Promise<discord.CommandPermissionsResponse>;
  /** Get guild ban, GET /guilds/{guild_id}/bans/{user_id}. Costs nothing. */
  getGuildBan(arg: {
    guild_id: string;
    user_id: string;
  }): Promise<discord.GuildBanResponse>;
  /** Get guild emoji, GET /guilds/{guild_id}/emojis/{emoji_id}. Costs nothing. */
  getGuildEmoji(arg: {
    guild_id: string;
    emoji_id: string;
  }): Promise<discord.EmojiResponse>;
  /** Get guild join requests, GET /guilds/{guild_id}/requests. Costs nothing. */
  getGuildJoinRequests(arg: {
    guild_id: string;
    status?: string;
    limit?: number;
    before?: string;
    after?: string;
  }): Promise<discord.GuildJoinRequestsListResponse>;
  /** Get guild member, GET /guilds/{guild_id}/members/{user_id}. Costs nothing. */
  getGuildMember(arg: {
    guild_id: string;
    user_id: string;
  }): Promise<discord.GuildMemberResponse>;
  /** Get guild new member welcome, GET /guilds/{guild_id}/new-member-welcome. Costs nothing. */
  getGuildNewMemberWelcome(arg: {
    guild_id: string;
  }): Promise<discord.GuildHomeSettingsResponse | null>;
  /** Get guild preview, GET /guilds/{guild_id}/preview. Costs nothing. */
  getGuildPreview(arg: {
    guild_id: string;
  }): Promise<discord.GuildPreviewResponse>;
  /** Get guild role, GET /guilds/{guild_id}/roles/{role_id}. Costs nothing. */
  getGuildRole(arg: {
    guild_id: string;
    role_id: string;
  }): Promise<discord.GuildRoleResponse>;
  /** Get guild scheduled event, GET /guilds/{guild_id}/scheduled-events/{guild_scheduled_event_id}. Costs nothing. */
  getGuildScheduledEvent(arg: {
    guild_id: string;
    guild_scheduled_event_id: string;
    with_user_count?: boolean;
  }): Promise<discord.ExternalScheduledEventResponse | discord.StageScheduledEventResponse | discord.VoiceScheduledEventResponse>;
  /** Get guild soundboard sound, GET /guilds/{guild_id}/soundboard-sounds/{sound_id}. Costs nothing. */
  getGuildSoundboardSound(arg: {
    guild_id: string;
    sound_id: string;
  }): Promise<discord.SoundboardSoundResponse>;
  /** Get guild sticker, GET /guilds/{guild_id}/stickers/{sticker_id}. Costs nothing. */
  getGuildSticker(arg: {
    guild_id: string;
    sticker_id: string;
  }): Promise<discord.GuildStickerResponse>;
  /** Get guild template, GET /guilds/templates/{code}. Costs nothing. */
  getGuildTemplate(arg: {
    code: string;
  }): Promise<discord.GuildTemplateResponse>;
  /** Get guild vanity url, GET /guilds/{guild_id}/vanity-url. Costs nothing. */
  getGuildVanityUrl(arg: {
    guild_id: string;
  }): Promise<discord.VanityURLResponse>;
  /** Get guild webhooks, GET /guilds/{guild_id}/webhooks. Costs nothing. */
  getGuildWebhooks(arg: {
    guild_id: string;
  }): Promise<(discord.ApplicationIncomingWebhookResponse | discord.ChannelFollowerWebhookResponse | discord.GuildIncomingWebhookResponse)[]>;
  /** Get guild welcome screen, GET /guilds/{guild_id}/welcome-screen. Costs nothing. */
  getGuildWelcomeScreen(arg: {
    guild_id: string;
  }): Promise<discord.GuildWelcomeScreenResponse>;
  /** Get guild widget, GET /guilds/{guild_id}/widget.json. Costs nothing. */
  getGuildWidget(arg: {
    guild_id: string;
  }): Promise<discord.WidgetResponse>;
  /** Get guild widget png, GET /guilds/{guild_id}/widget.png. Costs nothing. */
  getGuildWidgetPng(arg: {
    guild_id: string;
    style?: string;
  }): Promise<null>;
  /** Get guild widget settings, GET /guilds/{guild_id}/widget. Costs nothing. */
  getGuildWidgetSettings(arg: {
    guild_id: string;
  }): Promise<discord.WidgetSettingsResponse>;
  /** Get guilds onboarding, GET /guilds/{guild_id}/onboarding. Costs nothing. */
  getGuildsOnboarding(arg: {
    guild_id: string;
  }): Promise<discord.UserGuildOnboardingResponse>;
  /** Get invite target users, GET /invites/{code}/target-users. Costs nothing. */
  getInviteTargetUsers(arg: {
    code: string;
  }): Promise<null>;
  /** Get invite target users job status, GET /invites/{code}/target-users/job-status. Costs nothing. */
  getInviteTargetUsersJobStatus(arg: {
    code: string;
  }): Promise<discord.TargetUsersJobStatusResponse>;
  /** Get lobby, GET /lobbies/{lobby_id}. Costs nothing. */
  getLobby(arg: {
    lobby_id: string;
  }): Promise<discord.LobbyResponse>;
  /** Get lobby messages, GET /lobbies/{lobby_id}/messages. Costs nothing. */
  getLobbyMessages(arg: {
    lobby_id: string;
    limit?: number;
  }): Promise<discord.LobbyMessageResponse[]>;
  /** Get message, GET /channels/{channel_id}/messages/{message_id}. Costs nothing. */
  getMessage(arg: {
    channel_id: string;
    message_id: string;
  }): Promise<discord.MessageResponse>;
  /** Get my application, GET /applications/@me. Costs nothing. */
  getMyApplication(): Promise<discord.PrivateApplicationResponse>;
  /** Get my oauth2 application, GET /oauth2/applications/@me. Costs nothing. */
  getMyOauth2Application(): Promise<discord.PrivateApplicationResponse>;
  /** Get my oauth2 authorization, GET /oauth2/@me. Costs nothing. */
  getMyOauth2Authorization(): Promise<discord.OAuth2GetAuthorizationResponse>;
  /** Get my user, GET /users/@me. Costs nothing. */
  getMyUser(): Promise<discord.UserPIIResponse>;
  /** Get openid connect userinfo, GET /oauth2/userinfo. Costs nothing. */
  getOpenidConnectUserinfo(): Promise<discord.OAuth2GetOpenIDConnectUserInfoResponse>;
  /** Get public keys, GET /oauth2/keys. Costs nothing. */
  getPublicKeys(): Promise<discord.OAuth2GetKeys>;
  /** Get self voice state, GET /guilds/{guild_id}/voice-states/@me. Costs nothing. */
  getSelfVoiceState(arg: {
    guild_id: string;
  }): Promise<discord.VoiceStateResponse>;
  /** Get sku subscription, GET /skus/{sku_id}/subscriptions/{subscription_id}. Costs nothing. */
  getSkuSubscription(arg: {
    sku_id: string;
    subscription_id: string;
    user_id?: string;
  }): Promise<discord.SubscriptionResponse>;
  /** Get sku subscriptions, GET /skus/{sku_id}/subscriptions. Costs nothing. */
  getSkuSubscriptions(arg: {
    sku_id: string;
    before?: string;
    after?: string;
    limit?: number;
    user_id?: string;
  }): Promise<discord.SubscriptionResponse[]>;
  /** Get soundboard default sounds, GET /soundboard-default-sounds. Costs nothing. */
  getSoundboardDefaultSounds(): Promise<discord.SoundboardSoundResponse[]>;
  /** Get stage instance, GET /stage-instances/{channel_id}. Costs nothing. */
  getStageInstance(arg: {
    channel_id: string;
  }): Promise<discord.StageInstanceResponse>;
  /** Get sticker, GET /stickers/{sticker_id}. Costs nothing. */
  getSticker(arg: {
    sticker_id: string;
  }): Promise<discord.GuildStickerResponse | discord.StandardStickerResponse>;
  /** Get sticker pack, GET /sticker-packs/{pack_id}. Costs nothing. */
  getStickerPack(arg: {
    pack_id: string;
  }): Promise<discord.StickerPackResponse>;
  /** Get thread member, GET /channels/{channel_id}/thread-members/{user_id}. Costs nothing. */
  getThreadMember(arg: {
    channel_id: string;
    user_id: string;
    with_member?: boolean;
  }): Promise<discord.ThreadMemberResponse>;
  /** Get user, GET /users/{user_id}. Costs nothing. */
  getUser(arg: {
    user_id: string;
  }): Promise<discord.UserResponse>;
  /** Get voice state, GET /guilds/{guild_id}/voice-states/{user_id}. Costs nothing. */
  getVoiceState(arg: {
    guild_id: string;
    user_id: string;
  }): Promise<discord.VoiceStateResponse>;
  /** Get webhook, GET /webhooks/{webhook_id}. Costs nothing. */
  getWebhook(arg: {
    webhook_id: string;
  }): Promise<discord.ApplicationIncomingWebhookResponse | discord.ChannelFollowerWebhookResponse | discord.GuildIncomingWebhookResponse>;
  /** Guild role member counts, GET /guilds/{guild_id}/roles/member-counts. Costs nothing. */
  guildRoleMemberCounts(arg: {
    guild_id: string;
  }): Promise<Record<string, number>>;
  /** Guild search, GET /guilds/{guild_id}/messages/search. Costs nothing. */
  guildSearch(arg: {
    guild_id: string;
    sort_by?: string;
    sort_order?: string;
    content?: string;
    slop?: number;
    author_id?: (string | number)[];
    author_type?: (string | number)[];
    mentions?: (string | number)[];
    mentions_role_id?: (string | number)[];
    replied_to_user_id?: (string | number)[];
    replied_to_message_id?: (string | number)[];
    mention_everyone?: boolean;
    min_id?: string;
    max_id?: string;
    limit?: number;
    offset?: number;
    has?: (string | number)[];
    link_hostname?: (string | number)[];
    embed_provider?: (string | number)[];
    embed_type?: (string | number)[];
    attachment_extension?: (string | number)[];
    attachment_filename?: (string | number)[];
    pinned?: boolean;
    include_nsfw?: boolean;
    channel_id?: (string | number)[];
  }): Promise<discord.GuildSearchResponse | discord.SearchIndexNotReadyResponse>;
  /** Invite resolve, GET /invites/{code}. Costs nothing. */
  inviteResolve(arg: {
    code: string;
    with_counts?: boolean;
    guild_scheduled_event_id?: string;
    target_channel_id?: string;
    target_message_id?: string;
  }): Promise<discord.FriendInviteResponse | discord.GroupDMInviteResponse | discord.GuildInviteResponse>;
  /** Invite revoke, DELETE /invites/{code}. Costs 1 credit. */
  inviteRevoke(arg: {
    code: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<discord.FriendInviteResponse | discord.GroupDMInviteResponse | discord.GuildInviteResponse>;
  /** Join thread, PUT /channels/{channel_id}/thread-members/@me. Costs 1 credit. */
  joinThread(arg: {
    channel_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<null>;
  /** Leave lobby, DELETE /lobbies/{lobby_id}/members/@me. Costs 1 credit. */
  leaveLobby(arg: {
    lobby_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<null>;
  /** Leave thread, DELETE /channels/{channel_id}/thread-members/@me. Costs 1 credit. */
  leaveThread(arg: {
    channel_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<null>;
  /** List application commands, GET /applications/{application_id}/commands. Costs nothing. */
  listApplicationCommands(arg: {
    application_id: string;
    with_localizations?: boolean;
  }): Promise<discord.ApplicationCommandResponse[]>;
  /** List application emojis, GET /applications/{application_id}/emojis. Costs nothing. */
  listApplicationEmojis(arg: {
    application_id: string;
  }): Promise<discord.ListApplicationEmojisResponse>;
  /** List auto moderation rules, GET /guilds/{guild_id}/auto-moderation/rules. Costs nothing. */
  listAutoModerationRules(arg: {
    guild_id: string;
  }): Promise<(discord.DefaultKeywordRuleResponse | discord.KeywordRuleResponse | discord.MLSpamRuleResponse | discord.MentionSpamRuleResponse | discord.UserProfileRuleResponse | null)[]>;
  /** List channel invites, GET /channels/{channel_id}/invites. Costs nothing. */
  listChannelInvites(arg: {
    channel_id: string;
  }): Promise<(discord.FriendInviteResponse | discord.GroupDMInviteResponse | discord.GuildInviteResponse | null)[]>;
  /** List channel webhooks, GET /channels/{channel_id}/webhooks. Costs nothing. */
  listChannelWebhooks(arg: {
    channel_id: string;
  }): Promise<(discord.ApplicationIncomingWebhookResponse | discord.ChannelFollowerWebhookResponse | discord.GuildIncomingWebhookResponse)[]>;
  /** List guild application command permissions, GET /applications/{application_id}/guilds/{guild_id}/commands/permissions. Costs nothing. */
  listGuildApplicationCommandPermissions(arg: {
    application_id: string;
    guild_id: string;
  }): Promise<discord.CommandPermissionsResponse[]>;
  /** List guild application commands, GET /applications/{application_id}/guilds/{guild_id}/commands. Costs nothing. */
  listGuildApplicationCommands(arg: {
    application_id: string;
    guild_id: string;
    with_localizations?: boolean;
  }): Promise<discord.ApplicationCommandResponse[]>;
  /** List guild audit log entries, GET /guilds/{guild_id}/audit-logs. Costs nothing. */
  listGuildAuditLogEntries(arg: {
    guild_id: string;
    user_id?: string;
    target_id?: string;
    action_type?: number;
    before?: string;
    after?: string;
    limit?: number;
  }): Promise<discord.GuildAuditLogResponse>;
  /** List guild bans, GET /guilds/{guild_id}/bans. Costs nothing. */
  listGuildBans(arg: {
    guild_id: string;
    limit?: number;
    before?: string;
    after?: string;
  }): Promise<discord.GuildBanResponse[]>;
  /** List guild channels, GET /guilds/{guild_id}/channels. Costs nothing. */
  listGuildChannels(arg: {
    guild_id: string;
  }): Promise<(discord.GuildChannelResponse | discord.PrivateChannelResponse | discord.PrivateGroupChannelResponse | discord.ThreadResponse)[]>;
  /** List guild emojis, GET /guilds/{guild_id}/emojis. Costs nothing. */
  listGuildEmojis(arg: {
    guild_id: string;
  }): Promise<discord.EmojiResponse[]>;
  /** List guild integrations, GET /guilds/{guild_id}/integrations. Costs nothing. */
  listGuildIntegrations(arg: {
    guild_id: string;
  }): Promise<(discord.DiscordIntegrationResponse | discord.ExternalConnectionIntegrationResponse | discord.GuildSubscriptionIntegrationResponse)[]>;
  /** List guild invites, GET /guilds/{guild_id}/invites. Costs nothing. */
  listGuildInvites(arg: {
    guild_id: string;
  }): Promise<(discord.FriendInviteResponse | discord.GroupDMInviteResponse | discord.GuildInviteResponse | null)[]>;
  /** List guild members, GET /guilds/{guild_id}/members. Costs nothing. */
  listGuildMembers(arg: {
    guild_id: string;
    limit?: number;
    after?: string;
  }): Promise<discord.GuildMemberResponse[]>;
  /** List guild roles, GET /guilds/{guild_id}/roles. Costs nothing. */
  listGuildRoles(arg: {
    guild_id: string;
  }): Promise<discord.GuildRoleResponse[]>;
  /** List guild scheduled event exception users, GET /guilds/{guild_id}/scheduled-events/{guild_scheduled_event_id}/{guild_scheduled_event_exception_id}/users. Costs nothing. */
  listGuildScheduledEventExceptionUsers(arg: {
    guild_id: string;
    guild_scheduled_event_id: string;
    guild_scheduled_event_exception_id: string;
    with_member?: boolean;
    limit?: number;
    before?: string;
    after?: string;
  }): Promise<discord.ScheduledEventUserResponse[]>;
  /** List guild scheduled event users, GET /guilds/{guild_id}/scheduled-events/{guild_scheduled_event_id}/users. Costs nothing. */
  listGuildScheduledEventUsers(arg: {
    guild_id: string;
    guild_scheduled_event_id: string;
    with_member?: boolean;
    limit?: number;
    before?: string;
    after?: string;
  }): Promise<discord.ScheduledEventUserResponse[]>;
  /** List guild scheduled events, GET /guilds/{guild_id}/scheduled-events. Costs nothing. */
  listGuildScheduledEvents(arg: {
    guild_id: string;
    with_user_count?: boolean;
  }): Promise<(discord.ExternalScheduledEventResponse | discord.StageScheduledEventResponse | discord.VoiceScheduledEventResponse)[]>;
  /** List guild soundboard sounds, GET /guilds/{guild_id}/soundboard-sounds. Costs nothing. */
  listGuildSoundboardSounds(arg: {
    guild_id: string;
  }): Promise<discord.ListGuildSoundboardSoundsResponse>;
  /** List guild stickers, GET /guilds/{guild_id}/stickers. Costs nothing. */
  listGuildStickers(arg: {
    guild_id: string;
  }): Promise<discord.GuildStickerResponse[]>;
  /** List guild templates, GET /guilds/{guild_id}/templates. Costs nothing. */
  listGuildTemplates(arg: {
    guild_id: string;
  }): Promise<discord.GuildTemplateResponse[]>;
  /** List guild voice regions, GET /guilds/{guild_id}/regions. Costs nothing. */
  listGuildVoiceRegions(arg: {
    guild_id: string;
  }): Promise<discord.VoiceRegionResponse[]>;
  /** List message reactions by emoji, GET /channels/{channel_id}/messages/{message_id}/reactions/{emoji_name}. Costs nothing. */
  listMessageReactionsByEmoji(arg: {
    channel_id: string;
    message_id: string;
    emoji_name: string;
    after?: string;
    limit?: number;
    type?: number;
  }): Promise<discord.UserResponse[]>;
  /** List messages, GET /channels/{channel_id}/messages. Costs nothing. */
  listMessages(arg: {
    channel_id: string;
    around?: string;
    before?: string;
    after?: string;
    limit?: number;
  }): Promise<discord.MessageResponse[]>;
  /** List my connections, GET /users/@me/connections. Costs nothing. */
  listMyConnections(): Promise<discord.ConnectedAccountResponse[]>;
  /** List my guilds, GET /users/@me/guilds. Costs nothing. */
  listMyGuilds(arg?: {
    before?: string;
    after?: string;
    limit?: number;
    shard?: number;
    with_counts?: boolean;
  }): Promise<discord.MyGuildResponse[]>;
  /** List my private archived threads, GET /channels/{channel_id}/users/@me/threads/archived/private. Costs nothing. */
  listMyPrivateArchivedThreads(arg: {
    channel_id: string;
    before?: string;
    limit?: number;
  }): Promise<discord.ThreadsResponse>;
  /** List pins, GET /channels/{channel_id}/messages/pins. Costs nothing. */
  listPins(arg: {
    channel_id: string;
    before?: string;
    limit?: number;
  }): Promise<discord.PinnedMessagesResponse>;
  /** List private archived threads, GET /channels/{channel_id}/threads/archived/private. Costs nothing. */
  listPrivateArchivedThreads(arg: {
    channel_id: string;
    before?: string;
    limit?: number;
  }): Promise<discord.ThreadsResponse>;
  /** List public archived threads, GET /channels/{channel_id}/threads/archived/public. Costs nothing. */
  listPublicArchivedThreads(arg: {
    channel_id: string;
    before?: string;
    limit?: number;
  }): Promise<discord.ThreadsResponse>;
  /** List sticker packs, GET /sticker-packs. Costs nothing. */
  listStickerPacks(): Promise<discord.StickerPackCollectionResponse>;
  /** List thread members, GET /channels/{channel_id}/thread-members. Costs nothing. */
  listThreadMembers(arg: {
    channel_id: string;
    with_member?: boolean;
    limit?: number;
    after?: string;
  }): Promise<discord.ThreadMemberResponse[]>;
  /** List voice regions, GET /voice/regions. Costs nothing. */
  listVoiceRegions(): Promise<discord.VoiceRegionResponse[]>;
  /** Partner sdk token, POST /partner-sdk/token. Costs 1 credit. */
  partnerSdkToken(arg: {
    /** The JSON body of the request. */ body: { client_id: string; client_secret?: string | null; external_auth_token: string; external_auth_type: discord.ApplicationIdentityProviderAuthType };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<discord.ProvisionalTokenResponse>;
  /** Partner sdk unmerge provisional account, POST /partner-sdk/provisional-accounts/unmerge. Costs 1 credit. */
  partnerSdkUnmergeProvisionalAccount(arg: {
    /** The JSON body of the request. */ body: { client_id: string; client_secret?: string | null; external_auth_token: string; external_auth_type: discord.ApplicationIdentityProviderAuthType };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<null>;
  /** Poll expire, POST /channels/{channel_id}/polls/{message_id}/expire. Costs 1 credit. */
  pollExpire(arg: {
    channel_id: string;
    message_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<discord.MessageResponse>;
  /** Preview prune guild, GET /guilds/{guild_id}/prune. Costs nothing. */
  previewPruneGuild(arg: {
    guild_id: string;
    days?: number;
    include_roles?: string;
  }): Promise<discord.GuildPruneResponse>;
  /** Put guilds onboarding, PUT /guilds/{guild_id}/onboarding. Costs 1 credit. */
  putGuildsOnboarding(arg: {
    guild_id: string;
    /** The JSON body of the request. */ body?: discord.UpdateGuildOnboardingRequest;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<discord.GuildOnboardingResponse>;
  /** Remove invite target user, DELETE /invites/{code}/target-users/{user_id}. Costs 1 credit. */
  removeInviteTargetUser(arg: {
    code: string;
    user_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<null>;
  /** Search guild members, GET /guilds/{guild_id}/members/search. Costs nothing. */
  searchGuildMembers(arg: {
    guild_id: string;
    limit?: number;
    query: string;
  }): Promise<discord.GuildMemberResponse[]>;
  /** Send soundboard sound, POST /channels/{channel_id}/send-soundboard-sound. Costs 1 credit. */
  sendSoundboardSound(arg: {
    channel_id: string;
    /** The JSON body of the request. */ body: discord.SoundboardSoundSendRequest;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<null>;
  /** Set channel permission overwrite, PUT /channels/{channel_id}/permissions/{overwrite_id}. Costs 1 credit. */
  setChannelPermissionOverwrite(arg: {
    channel_id: string;
    overwrite_id: string;
    /** The JSON body of the request. */ body?: { type?: discord.ChannelPermissionOverwrites | null; allow?: number | null; deny?: number | null };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<null>;
  /** Sync guild template, PUT /guilds/{guild_id}/templates/{code}. Costs 1 credit. */
  syncGuildTemplate(arg: {
    guild_id: string;
    code: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<discord.GuildTemplateResponse>;
  /** Thread search, GET /channels/{channel_id}/threads/search. Costs nothing. */
  threadSearch(arg: {
    channel_id: string;
    name?: string;
    slop?: number;
    min_id?: string;
    max_id?: string;
    tag?: string;
    tag_setting?: string;
    archived?: boolean;
    sort_by?: string;
    sort_order?: string;
    limit?: number;
    offset?: number;
  }): Promise<discord.ThreadSearchResponse | discord.SearchIndexNotReadyResponse>;
  /** Trigger typing indicator, POST /channels/{channel_id}/typing. Costs 1 credit. */
  triggerTypingIndicator(arg: {
    channel_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<discord.TypingIndicatorResponse | null>;
  /** Unban user from guild, DELETE /guilds/{guild_id}/bans/{user_id}. Costs 1 credit. */
  unbanUserFromGuild(arg: {
    guild_id: string;
    user_id: string;
    /** The JSON body of the request. */ body?: discord.UnbanUserFromGuildRequest;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<null>;
  /** Update auto moderation rule, PATCH /guilds/{guild_id}/auto-moderation/rules/{rule_id}. Costs 1 credit. */
  updateAutoModerationRule(arg: {
    guild_id: string;
    rule_id: string;
    /** The JSON body of the request. */ body?: discord.DefaultKeywordListUpsertRequestPartial | discord.KeywordUpsertRequestPartial | discord.MLSpamUpsertRequestPartial | discord.MentionSpamUpsertRequestPartial | discord.UserProfileUpsertRequestPartial;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<discord.DefaultKeywordRuleResponse | discord.KeywordRuleResponse | discord.MLSpamRuleResponse | discord.MentionSpamRuleResponse | discord.UserProfileRuleResponse>;
  /** Update channel, PATCH /channels/{channel_id}. Costs 1 credit. */
  updateChannel(arg: {
    channel_id: string;
    /** The JSON body of the request. */ body?: discord.UpdateDMRequestPartial | discord.UpdateGroupDMRequestPartial | discord.UpdateGuildChannelRequestPartial | discord.UpdateThreadRequestPartial;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<discord.GuildChannelResponse | discord.PrivateChannelResponse | discord.PrivateGroupChannelResponse | discord.ThreadResponse>;
  /** Update guild, PATCH /guilds/{guild_id}. Costs 1 credit. */
  updateGuild(arg: {
    guild_id: string;
    /** The JSON body of the request. */ body?: discord.GuildPatchRequestPartial;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<discord.GuildResponse>;
  /** Update guild emoji, PATCH /guilds/{guild_id}/emojis/{emoji_id}. Costs 1 credit. */
  updateGuildEmoji(arg: {
    guild_id: string;
    emoji_id: string;
    /** The JSON body of the request. */ body?: { name?: string; roles?: (string | null)[] | null };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<discord.EmojiResponse>;
  /** Update guild incident actions, PUT /guilds/{guild_id}/incident-actions. Costs 1 credit. */
  updateGuildIncidentActions(arg: {
    guild_id: string;
    /** The JSON body of the request. */ body?: discord.GuildIncidentActionsRequest;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<discord.GuildIncidentsDataResponse>;
  /** Update guild member, PATCH /guilds/{guild_id}/members/{user_id}. Costs 1 credit. */
  updateGuildMember(arg: {
    guild_id: string;
    user_id: string;
    /** The JSON body of the request. */ body?: { nick?: string | null; roles?: (string | null)[] | null; mute?: boolean | null; deaf?: boolean | null; channel_id?: string | null; communication_disabled_until?: string | null; flags?: number | null };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<discord.GuildMemberResponse | null>;
  /** Update guild role, PATCH /guilds/{guild_id}/roles/{role_id}. Costs 1 credit. */
  updateGuildRole(arg: {
    guild_id: string;
    role_id: string;
    /** The JSON body of the request. */ body?: discord.UpdateRoleRequestPartial;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<discord.GuildRoleResponse>;
  /** Update guild scheduled event, PATCH /guilds/{guild_id}/scheduled-events/{guild_scheduled_event_id}. Costs 1 credit. */
  updateGuildScheduledEvent(arg: {
    guild_id: string;
    guild_scheduled_event_id: string;
    /** The JSON body of the request. */ body?: discord.ExternalScheduledEventPatchRequestPartial | discord.StageScheduledEventPatchRequestPartial | discord.VoiceScheduledEventPatchRequestPartial;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<discord.ExternalScheduledEventResponse | discord.StageScheduledEventResponse | discord.VoiceScheduledEventResponse>;
  /** Update guild scheduled event exception, PATCH /guilds/{guild_id}/scheduled-events/{guild_scheduled_event_id}/exceptions/{exception_id}. Costs 1 credit. */
  updateGuildScheduledEventException(arg: {
    guild_id: string;
    guild_scheduled_event_id: string;
    exception_id: string;
    /** The JSON body of the request. */ body?: discord.GuildScheduledEventExceptionPatchRequestPartial;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<discord.GuildScheduledEventExceptionResponse>;
  /** Update guild soundboard sound, PATCH /guilds/{guild_id}/soundboard-sounds/{sound_id}. Costs 1 credit. */
  updateGuildSoundboardSound(arg: {
    guild_id: string;
    sound_id: string;
    /** The JSON body of the request. */ body?: discord.SoundboardPatchRequestPartial;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<discord.SoundboardSoundResponse>;
  /** Update guild sticker, PATCH /guilds/{guild_id}/stickers/{sticker_id}. Costs 1 credit. */
  updateGuildSticker(arg: {
    guild_id: string;
    sticker_id: string;
    /** The JSON body of the request. */ body?: { name?: string; tags?: string; description?: string | null };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<discord.GuildStickerResponse>;
  /** Update guild template, PATCH /guilds/{guild_id}/templates/{code}. Costs 1 credit. */
  updateGuildTemplate(arg: {
    guild_id: string;
    code: string;
    /** The JSON body of the request. */ body?: { name?: string; description?: string | null };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<discord.GuildTemplateResponse>;
  /** Update guild welcome screen, PATCH /guilds/{guild_id}/welcome-screen. Costs 1 credit. */
  updateGuildWelcomeScreen(arg: {
    guild_id: string;
    /** The JSON body of the request. */ body?: discord.WelcomeScreenPatchRequestPartial;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<discord.GuildWelcomeScreenResponse>;
  /** Update guild widget settings, PATCH /guilds/{guild_id}/widget. Costs 1 credit. */
  updateGuildWidgetSettings(arg: {
    guild_id: string;
    /** The JSON body of the request. */ body?: { channel_id?: string | null; enabled?: boolean | null };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<discord.WidgetSettingsResponse>;
  /** Update lobby message external moderation metadata, PUT /lobbies/{lobby_id}/messages/{message_id}/moderation-metadata. Costs 1 credit. */
  updateLobbyMessageExternalModerationMetadata(arg: {
    lobby_id: string;
    message_id: string;
    /** The JSON body of the request. */ body?: Record<string, string>;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<null>;
  /** Update message, PATCH /channels/{channel_id}/messages/{message_id}. Costs 1 credit. */
  updateMessage(arg: {
    channel_id: string;
    message_id: string;
    /** The JSON body of the request. */ body?: discord.MessageEditRequestPartial;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<discord.MessageResponse>;
  /** Update self voice state, PATCH /guilds/{guild_id}/voice-states/@me. Costs 1 credit. */
  updateSelfVoiceState(arg: {
    guild_id: string;
    /** The JSON body of the request. */ body?: discord.UpdateSelfVoiceStateRequestPartial;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<null>;
  /** Update stage instance, PATCH /stage-instances/{channel_id}. Costs 1 credit. */
  updateStageInstance(arg: {
    channel_id: string;
    /** The JSON body of the request. */ body?: { topic?: string; privacy_level?: discord.StageInstancesPrivacyLevels };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<discord.StageInstanceResponse>;
  /** Update user message external moderation metadata, PUT /partner-sdk/dms/{user_id_1}/{user_id_2}/messages/{message_id}/moderation-metadata. Costs 1 credit. */
  updateUserMessageExternalModerationMetadata(arg: {
    user_id_1: string;
    user_id_2: string;
    message_id: string;
    /** The JSON body of the request. */ body?: Record<string, string>;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<null>;
  /** Update voice channel status, PUT /channels/{channel_id}/voice-status. Costs 1 credit. */
  updateVoiceChannelStatus(arg: {
    channel_id: string;
    /** The JSON body of the request. */ body?: { status?: string | null };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<null>;
  /** Update voice state, PATCH /guilds/{guild_id}/voice-states/{user_id}. Costs 1 credit. */
  updateVoiceState(arg: {
    guild_id: string;
    user_id: string;
    /** The JSON body of the request. */ body?: discord.UpdateVoiceStateRequestPartial;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<null>;
  /** Update webhook, PATCH /webhooks/{webhook_id}. Costs 1 credit. */
  updateWebhook(arg: {
    webhook_id: string;
    /** The JSON body of the request. */ body?: { name?: string; avatar?: string | null; channel_id?: string | null };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<discord.ApplicationIncomingWebhookResponse | discord.ChannelFollowerWebhookResponse | discord.GuildIncomingWebhookResponse>;
};

declare const cookieApi: {
  /** Create transcript, POST /api/transcript. Costs 1 credit. */
  createTranscript(arg: {
    channel_id: string;
    /** The JSON body of the request. */ body: { bot_token: any; title?: any; description?: any; password?: any; gtag?: any; favicon?: any; image_restore?: any; domain?: any };
  }): Promise<any>;
  /** Create card, POST /api/cards/card-builder/build. Costs 1 credit. */
  createCard(arg?: {
    dev?: boolean;
    base64?: boolean;
    /** The JSON body of the request. */ body?: any;
  }): Promise<any>;
  /** Create qr code, POST /api/images/qr-code. Costs 1 credit. */
  createQrCode(arg: {
    /** The JSON body of the request. */ body: { data: any; style?: "standart" | "dots"; border?: number; center_image_url?: any };
  }): Promise<any>;
  /** Create captcha, POST /api/security/captcha/create. Costs 1 credit. */
  createCaptcha(arg: {
    captcha_provider: "Cloudflare" | "Google";
    color?: string;
    font?: string;
    expires?: number;
  }): Promise<any>;
  /** Get captcha, GET /api/security/captcha/get-captcha. Costs 1 credit. */
  getCaptcha(arg: {
    captcha_id: string;
  }): Promise<any>;
  /** Get minecraft user, GET /api/minecraft/get-user. Costs 1 credit. */
  getMinecraftUser(arg: {
    code: string;
  }): Promise<any>;
  /** Get current time, GET /api/time/current-time. Costs 1 credit. */
  getCurrentTime(): Promise<any>;
};

declare const erlc: {
  /** Get server, GET /v2/server. Costs 1 credit. */
  getServer(arg?: {
    Players?: boolean;
    Staff?: boolean;
    JoinLogs?: boolean;
    Queue?: boolean;
    KillLogs?: boolean;
    CommandLogs?: boolean;
    ModCalls?: boolean;
    EmergencyCalls?: boolean;
    Vehicles?: boolean;
  }): Promise<any>;
  /** Run command, POST /v2/server/command. Costs 1 credit. */
  runCommand(arg: {
    /** The JSON body of the request. */ body: { command: any };
  }): Promise<any>;
};

declare const http: {
  /** Send an HTTP request to a URL, like the Send API Request block. {{secrets.NAME}} in the URL, the headers and the body is replaced with the app's secret, which the code never sees. Responses that aren't 2xx resolve too, only failing to get a response rejects. fetch(url, init) calls it. Costs 3 credits. */
  request(arg: {
    /** The URL, starting with http:// or https://. */ url: string;
    /** Defaults to GET. */ method?: "GET" | "POST" | "PUT" | "PATCH" | "DELETE";
    /** Headers by name. */ headers?: Record<string, string>;
    /** The body as text, like JSON.stringify(data). */ body?: string;
  }): Promise<{ status: number; status_text: string; headers: Record<string, string>; body: string }>;
};

declare const variables: {
  /** Get the value of a stored variable, or null if it has none. Costs nothing. */
  get(arg: {
    /** Name of the variable under Stored variables of the block. */ name: string;
    /** Scope of the value, like a user ID for one value per user. Required for scoped variables, ignored for others. */ scope?: string;
  }): Promise<any>;
  /** Set the value of a stored variable, and get it back. Costs 1 credit. */
  set(arg: {
    /** Name of the variable under Stored variables of the block. */ name: string;
    /** Scope of the value, like a user ID for one value per user. Required for scoped variables, ignored for others. */ scope?: string;
    /** The value: text, a number, a list, an object, or null. */ value: any;
  }): Promise<any>;
  /** Add an item to the end of the list in a stored variable, or text to the end of other values, and get the new value. An object or a list starts a list. For a list of text or numbers, set the variable to an empty list first. An empty value adds nothing. Calls at the same time don't overwrite each other. Costs 1 credit. */
  append(arg: {
    /** Name of the variable under Stored variables of the block. */ name: string;
    /** Scope of the value, like a user ID for one value per user. Required for scoped variables, ignored for others. */ scope?: string;
    /** The value: text, a number, a list, an object, or null. */ value: any;
  }): Promise<any>;
  /** Like append, at the start of the list or the text. Costs 1 credit. */
  prepend(arg: {
    /** Name of the variable under Stored variables of the block. */ name: string;
    /** Scope of the value, like a user ID for one value per user. Required for scoped variables, ignored for others. */ scope?: string;
    /** The value: text, a number, a list, an object, or null. */ value: any;
  }): Promise<any>;
  /** Add to the number in a stored variable, which starts at 0, and get the new value. Calls at the same time don't overwrite each other, unlike get and set. Costs 1 credit. */
  increment(arg: {
    /** Name of the variable under Stored variables of the block. */ name: string;
    /** Scope of the value, like a user ID for one value per user. Required for scoped variables, ignored for others. */ scope?: string;
    /** How much to add, negative to subtract. Defaults to 1. */ by?: number;
  }): Promise<number>;
  /** Delete the value of a stored variable. Costs 1 credit. */
  delete(arg: {
    /** Name of the variable under Stored variables of the block. */ name: string;
    /** Scope of the value, like a user ID for one value per user. Required for scoped variables, ignored for others. */ scope?: string;
  }): Promise<null>;
  /** List the values of a stored variable, like a leaderboard of the highest ones. Costs 1 credit. */
  list(arg: {
    /** Name of the variable under Stored variables of the block. */ name: string;
    /** Only values whose scope starts with it, like the server ID for scopes like <server ID>:<user ID>. */ scope_prefix?: string;
    /** Defaults to value_desc, the highest values first. */ sort?: "value_desc" | "value_asc" | "updated_desc" | "updated_asc";
    /** How many values, from 1 to 100. Defaults to 10. */ limit?: number;
    /** How many values to skip first, for the next page. */ skip?: number;
  }): Promise<{ scope: any; value: any; updated_at: number }[]>;
};

/** The objects of Discord's API that its calls take and return. */
declare namespace discord {
  interface AccountResponse {
    id: string;
    name: string | null;
  }
  interface ActionRowComponentForMessageRequest {
    type: 1;
    id?: number | null;
    components: (discord.ButtonComponentForMessageRequest | discord.ChannelSelectComponentForMessageRequest | discord.MentionableSelectComponentForMessageRequest | discord.RoleSelectComponentForMessageRequest | discord.StringSelectComponentForMessageRequest | discord.UserSelectComponentForMessageRequest)[];
  }
  interface ActionRowComponentResponse {
    type: 1;
    id: number;
    components: (discord.ButtonComponentResponse | discord.ChannelSelectComponentResponse | discord.MentionableSelectComponentResponse | discord.RoleSelectComponentResponse | discord.StringSelectComponentResponse | discord.TextInputComponentResponse | discord.UserSelectComponentResponse)[];
  }
  type ActivityActionTypes = 1 | 2 | 3 | 5 | 6;
  type AfkTimeouts = 60 | 300 | 900 | 1800 | 3600;
  type AllowedMentionTypes = "users" | "roles" | "everyone";
  interface ApplicationCommandAttachmentOptionResponse {
    type: 11;
    name: string;
    name_localized?: string;
    name_localizations?: Record<string, string> | null;
    description: string;
    description_localized?: string;
    description_localizations?: Record<string, string> | null;
    required?: boolean;
    file_types?: string[];
  }
  interface ApplicationCommandBooleanOptionResponse {
    type: 5;
    name: string;
    name_localized?: string;
    name_localizations?: Record<string, string> | null;
    description: string;
    description_localized?: string;
    description_localizations?: Record<string, string> | null;
    required?: boolean;
  }
  interface ApplicationCommandChannelOptionResponse {
    type: 7;
    name: string;
    name_localized?: string;
    name_localizations?: Record<string, string> | null;
    description: string;
    description_localized?: string;
    description_localizations?: Record<string, string> | null;
    required?: boolean;
    channel_types?: discord.ChannelTypes[];
  }
  type ApplicationCommandHandler = 1 | 2;
  interface ApplicationCommandIntegerOptionResponse {
    type: 4;
    name: string;
    name_localized?: string;
    name_localizations?: Record<string, string> | null;
    description: string;
    description_localized?: string;
    description_localizations?: Record<string, string> | null;
    required?: boolean;
    autocomplete?: boolean;
    choices?: discord.ApplicationCommandOptionIntegerChoiceResponse[];
    min_value?: number;
    max_value?: number;
  }
  interface ApplicationCommandInteractionMetadataResponse {
    id: string;
    type: 2;
    user?: discord.UserResponse;
    authorizing_integration_owners: Record<string, string>;
    original_response_message_id?: string;
    target_user?: discord.UserResponse;
    target_message_id?: string;
  }
  interface ApplicationCommandMentionableOptionResponse {
    type: 9;
    name: string;
    name_localized?: string;
    name_localizations?: Record<string, string> | null;
    description: string;
    description_localized?: string;
    description_localizations?: Record<string, string> | null;
    required?: boolean;
  }
  interface ApplicationCommandNumberOptionResponse {
    type: 10;
    name: string;
    name_localized?: string;
    name_localizations?: Record<string, string> | null;
    description: string;
    description_localized?: string;
    description_localizations?: Record<string, string> | null;
    required?: boolean;
    autocomplete?: boolean;
    choices?: discord.ApplicationCommandOptionNumberChoiceResponse[];
    min_value?: number;
    max_value?: number;
  }
  interface ApplicationCommandOptionIntegerChoiceResponse {
    name: string;
    name_localized?: string;
    name_localizations?: Record<string, string> | null;
    value: number;
  }
  interface ApplicationCommandOptionNumberChoiceResponse {
    name: string;
    name_localized?: string;
    name_localizations?: Record<string, string> | null;
    value: number;
  }
  interface ApplicationCommandOptionStringChoiceResponse {
    name: string;
    name_localized?: string;
    name_localizations?: Record<string, string> | null;
    value: string;
  }
  type ApplicationCommandPermissionType = 1 | 2 | 3;
  interface ApplicationCommandResponse {
    id: string;
    application_id: string;
    version: string;
    default_member_permissions: string | null;
    type: discord.ApplicationCommandType;
    name: string;
    name_localized?: string;
    name_localizations?: Record<string, string> | null;
    description: string;
    description_localized?: string;
    description_localizations?: Record<string, string> | null;
    guild_id?: string;
    dm_permission?: boolean;
    contexts?: discord.InteractionContextType[] | null;
    integration_types?: discord.ApplicationIntegrationType[];
    options?: (discord.ApplicationCommandAttachmentOptionResponse | discord.ApplicationCommandBooleanOptionResponse | discord.ApplicationCommandChannelOptionResponse | discord.ApplicationCommandIntegerOptionResponse | discord.ApplicationCommandMentionableOptionResponse | discord.ApplicationCommandNumberOptionResponse | discord.ApplicationCommandRoleOptionResponse | discord.ApplicationCommandStringOptionResponse | discord.ApplicationCommandSubcommandGroupOptionResponse | discord.ApplicationCommandSubcommandOptionResponse | discord.ApplicationCommandUserOptionResponse)[];
    nsfw?: boolean;
    /** Determines whether the interaction is handled by the app's interactions handler or by Discord */
    handler?: discord.ApplicationCommandHandler;
  }
  interface ApplicationCommandRoleOptionResponse {
    type: 8;
    name: string;
    name_localized?: string;
    name_localizations?: Record<string, string> | null;
    description: string;
    description_localized?: string;
    description_localizations?: Record<string, string> | null;
    required?: boolean;
  }
  interface ApplicationCommandStringOptionResponse {
    type: 3;
    name: string;
    name_localized?: string;
    name_localizations?: Record<string, string> | null;
    description: string;
    description_localized?: string;
    description_localizations?: Record<string, string> | null;
    required?: boolean;
    autocomplete?: boolean;
    choices?: discord.ApplicationCommandOptionStringChoiceResponse[];
    min_length?: number;
    max_length?: number;
  }
  interface ApplicationCommandSubcommandGroupOptionResponse {
    type: 2;
    name: string;
    name_localized?: string;
    name_localizations?: Record<string, string> | null;
    description: string;
    description_localized?: string;
    description_localizations?: Record<string, string> | null;
    required?: boolean;
    options?: discord.ApplicationCommandSubcommandOptionResponse[];
  }
  interface ApplicationCommandSubcommandOptionResponse {
    type: 1;
    name: string;
    name_localized?: string;
    name_localizations?: Record<string, string> | null;
    description: string;
    description_localized?: string;
    description_localizations?: Record<string, string> | null;
    required?: boolean;
    options?: (discord.ApplicationCommandAttachmentOptionResponse | discord.ApplicationCommandBooleanOptionResponse | discord.ApplicationCommandChannelOptionResponse | discord.ApplicationCommandIntegerOptionResponse | discord.ApplicationCommandMentionableOptionResponse | discord.ApplicationCommandNumberOptionResponse | discord.ApplicationCommandRoleOptionResponse | discord.ApplicationCommandStringOptionResponse | discord.ApplicationCommandUserOptionResponse)[];
  }
  type ApplicationCommandType = 1 | 2 | 3 | 4;
  interface ApplicationCommandUserOptionResponse {
    type: 6;
    name: string;
    name_localized?: string;
    name_localizations?: Record<string, string> | null;
    description: string;
    description_localized?: string;
    description_localizations?: Record<string, string> | null;
    required?: boolean;
  }
  type ApplicationEventWebhooksStatus = 1 | 2 | 3;
  type ApplicationExplicitContentFilterTypes = 0 | 1;
  type ApplicationIdentityProviderAuthType = "OIDC" | "EPIC_ONLINE_SERVICES_ACCESS_TOKEN" | "EPIC_ONLINE_SERVICES_ID_TOKEN" | "STEAM_SESSION_TICKET" | "UNITY_SERVICES_ID_TOKEN" | "DISCORD_BOT_ISSUED_ACCESS_TOKEN" | "APPLE_ID_TOKEN" | "PLAYSTATION_NETWORK_ID_TOKEN";
  interface ApplicationIncomingWebhookResponse {
    application_id: string | null;
    avatar: string | null;
    channel_id: string | null;
    guild_id?: string | null;
    id: string;
    name: string;
    type: 3;
    user?: discord.UserResponse;
  }
  type ApplicationIntegrationType = 0 | 1;
  interface ApplicationIntegrationTypeConfigurationResponse {
    oauth2_install_params?: discord.ApplicationOAuth2InstallParamsResponse;
  }
  interface ApplicationOAuth2InstallParamsResponse {
    scopes: ("applications.commands" | "bot")[];
    permissions: string;
  }
  interface ApplicationResponse {
    id: string;
    name: string;
    icon: string | null;
    description: string;
    type: discord.ApplicationTypes | null;
    cover_image?: string;
    primary_sku_id?: string;
    flags: number;
    flags_new: string;
    bot?: discord.UserResponse;
    slug?: string;
    vibegrations_project_id?: string;
    guild_id?: string;
    rpc_origins?: string[];
    bot_public?: boolean;
    bot_require_code_grant?: boolean;
    terms_of_service_url?: string;
    privacy_policy_url?: string;
    custom_install_url?: string;
    install_params?: discord.ApplicationOAuth2InstallParamsResponse;
    integration_types_config?: Record<string, discord.ApplicationIntegrationTypeConfigurationResponse>;
    verify_key: string;
    max_participants?: number | null;
    tags?: string[];
  }
  interface ApplicationRoleConnectionsMetadataItemResponse {
    type: discord.MetadataItemTypes;
    key: string;
    name: string;
    name_localizations?: Record<string, string> | null;
    description: string;
    description_localizations?: Record<string, string> | null;
  }
  type ApplicationTypes = 4;
  type AuditLogActionTypes = 1 | 10 | 11 | 12 | 13 | 14 | 15 | 20 | 21 | 22 | 23 | 24 | 25 | 26 | 27 | 28 | 30 | 31 | 32 | 40 | 41 | 42 | 50 | 51 | 52 | 60 | 61 | 62 | 72 | 73 | 74 | 75 | 80 | 81 | 82 | 83 | 84 | 85 | 90 | 91 | 92 | 100 | 101 | 102 | 110 | 111 | 112 | 121 | 130 | 131 | 132 | 140 | 141 | 142 | 143 | 144 | 145 | 146 | 150 | 151 | 163 | 164 | 165 | 166 | 167 | 171 | 172 | 180 | 190 | 191 | 192 | 193 | 200 | 201 | 202 | 211;
  interface AuditLogEntryResponse {
    id: string;
    action_type: discord.AuditLogActionTypes;
    user_id: string | null;
    target_id: string | null;
    changes?: discord.AuditLogObjectChangeResponse[];
    options?: Record<string, string>;
    reason?: string;
  }
  interface AuditLogObjectChangeResponse {
    key: string | null;
    new_value?: any;
    old_value?: any;
  }
  type AutomodEventType = 1 | 2;
  type AutomodKeywordPresetType = 1 | 2 | 3;
  type AvailableLocalesEnum = "ar" | "bg" | "cs" | "da" | "de" | "el" | "en-GB" | "en-US" | "es-419" | "es-ES" | "fi" | "fr" | "he" | "hi" | "hr" | "hu" | "id" | "it" | "ja" | "ko" | "lt" | "nl" | "no" | "pl" | "pt-BR" | "ro" | "ru" | "sv-SE" | "th" | "tr" | "uk" | "vi" | "zh-CN" | "zh-TW";
  interface BanUserFromGuildRequest {
    delete_message_seconds?: number | null;
    delete_message_days?: number | null;
  }
  interface BaseCreateMessageCreateRequest {
    content?: string | null;
    embeds?: discord.RichEmbed[] | null;
    allowed_mentions?: discord.MessageAllowedMentionsRequest | null;
    sticker_ids?: string[] | null;
    components?: (discord.ActionRowComponentForMessageRequest | discord.ContainerComponentForMessageRequest | discord.FileComponentForMessageRequest | discord.MediaGalleryComponentForMessageRequest | discord.SectionComponentForMessageRequest | discord.SeparatorComponentForMessageRequest | discord.TextDisplayComponentForMessageRequest)[] | null;
    flags?: number | null;
    attachments?: discord.MessageAttachmentRequest[] | null;
    poll?: discord.PollCreateRequest | null;
    shared_client_theme?: discord.CustomClientThemeShareRequest | null;
  }
  interface BasicApplicationResponseWithBot {
    id: string;
    name: string;
    icon: string | null;
    description: string;
    type: discord.ApplicationTypes | null;
    cover_image?: string;
    primary_sku_id?: string;
    flags: number;
    flags_new: string;
    bot?: discord.UserResponse;
  }
  interface BasicGuildMemberResponse {
    /** the member's guild avatar hash */
    avatar: string | null;
    /** data for the member's guild avatar decoration */
    avatar_decoration_data?: discord.UserAvatarDecorationResponse | null;
    /** the member's guild banner hash */
    banner: string | null;
    /** when the user's timeout will expire and the user will be able to communicate in the guild again, null or a time in the past if the user is not timed out */
    communication_disabled_until: string | null;
    /** guild member flags represented as a bit set, defaults to 0 */
    flags: number;
    /** when the user joined the guild */
    joined_at: string;
    /** this user's guild nickname */
    nick: string | null;
    /** whether the user has not yet passed the guild's Membership Screening requirements */
    pending: boolean;
    /** when the user started boosting the guild */
    premium_since: string | null;
    /** array of role object ids */
    roles: string[];
    /** data for the member's collectibles */
    collectibles?: discord.UserCollectiblesResponse | null;
  }
  interface BasicMessageResponse {
    type: discord.MessageType;
    content: string;
    mentions: discord.UserResponse[];
    mention_roles: string[];
    attachments: discord.MessageAttachmentResponse[];
    embeds: discord.MessageEmbedResponse[];
    timestamp: string;
    edited_timestamp: string | null;
    flags: number;
    components: (discord.ActionRowComponentResponse | discord.ContainerComponentResponse | discord.FileComponentResponse | discord.MediaGalleryComponentResponse | discord.SectionComponentResponse | discord.SeparatorComponentResponse | discord.TextDisplayComponentResponse)[];
    stickers?: (discord.GuildStickerResponse | discord.StandardStickerResponse)[];
    sticker_items?: discord.MessageStickerItemResponse[];
    id: string;
    channel_id: string;
    author: discord.UserResponse;
    pinned: boolean;
    mention_everyone: boolean;
    tts: boolean;
    call?: discord.MessageCallResponse;
    activity?: discord.MessageActivityResponse;
    application?: discord.BasicApplicationResponseWithBot;
    application_id?: string;
    interaction?: discord.MessageInteractionResponse;
    nonce?: number | string;
    webhook_id?: string;
    message_reference?: discord.MessageReferenceResponse;
    thread?: discord.ThreadResponse;
    mention_channels?: discord.MessageMentionChannelResponse[];
    role_subscription_data?: discord.MessageRoleSubscriptionDataResponse;
    purchase_notification?: discord.PurchaseNotificationResponse;
    position?: number;
    resolved?: discord.ResolvedObjectsResponse;
    poll?: discord.PollResponse;
    shared_client_theme?: discord.CustomClientThemeResponse;
    interaction_metadata?: discord.ApplicationCommandInteractionMetadataResponse | discord.MessageComponentInteractionMetadataResponse | discord.ModalSubmitInteractionMetadataResponse;
    message_snapshots?: discord.MessageSnapshotResponse[];
    lobby_member?: discord.MessageLobbyMemberResponse;
  }
  interface BlockMessageAction {
    type: 1;
    metadata?: discord.BlockMessageActionMetadata | null;
  }
  interface BlockMessageActionMetadata {
    custom_message?: string | null;
  }
  interface BlockMessageActionMetadataResponse {
    custom_message?: string;
  }
  interface BlockMessageActionResponse {
    type: 1;
    metadata: discord.BlockMessageActionMetadataResponse;
  }
  interface BotAddGuildMemberRequest {
    nick?: string | null;
    roles?: string[] | null;
    mute?: boolean | null;
    deaf?: boolean | null;
    access_token: string;
    flags?: number | null;
  }
  interface BulkLobbyMemberRequest {
    id: string;
    metadata?: Record<string, string> | null;
    flags?: 1 | null;
    additional_name?: string | null;
    remove_member?: boolean | null;
  }
  interface ButtonComponentForMessageRequest {
    type: 2;
    id?: number | null;
    custom_id?: string | null;
    style: discord.ButtonStyleTypes;
    label?: string | null;
    disabled?: boolean | null;
    url?: string | null;
    sku_id?: string | null;
    emoji?: discord.ComponentEmojiForRequest | null;
  }
  interface ButtonComponentResponse {
    type: 2;
    id: number;
    custom_id?: string;
    style: discord.ButtonStyleTypes;
    label?: string;
    disabled?: boolean;
    emoji?: discord.ComponentEmojiResponse;
    url?: string | null;
    sku_id?: string;
  }
  type ButtonStyleTypes = 1 | 2 | 3 | 4 | 5 | 6;
  interface ByNWeekday {
    /** The week to reoccur on (1-5, where 5 represents the last week) */
    n: number;
    /** The day within the week to reoccur on */
    day: discord.RecurrenceRuleWeekdays;
  }
  interface ByNWeekdayResponse {
    /** The week to reoccur on (1-5, where 5 represents the last week) */
    n: number;
    /** The day within the week to reoccur on */
    day: discord.RecurrenceRuleWeekdays;
  }
  interface ChannelFollowerResponse {
    channel_id: string;
    webhook_id: string;
  }
  interface ChannelFollowerWebhookResponse {
    application_id: string | null;
    avatar: string | null;
    channel_id: string | null;
    guild_id?: string | null;
    id: string;
    name: string;
    type: 2;
    user?: discord.UserResponse;
    source_guild?: discord.WebhookSourceGuildResponse;
    source_channel?: discord.WebhookSourceChannelResponse;
  }
  interface ChannelPermissionOverwriteRequest {
    id: string;
    type?: discord.ChannelPermissionOverwrites | null;
    allow?: number | null;
    deny?: number | null;
  }
  interface ChannelPermissionOverwriteResponse {
    id: string;
    type: discord.ChannelPermissionOverwrites;
    allow: string;
    deny: string;
  }
  type ChannelPermissionOverwrites = 0 | 1;
  interface ChannelSelectComponentForMessageRequest {
    type: 8;
    id?: number | null;
    custom_id: string;
    placeholder?: string | null;
    min_values?: number | null;
    max_values?: number | null;
    disabled?: boolean | null;
    required?: boolean | null;
    default_values?: discord.ChannelSelectDefaultValue[] | null;
    channel_types?: discord.ChannelTypes[] | null;
  }
  interface ChannelSelectComponentResponse {
    type: 8;
    id: number;
    custom_id: string;
    placeholder?: string;
    min_values: number;
    max_values: number;
    disabled?: boolean;
    channel_types?: discord.ChannelTypes[];
    default_values?: discord.ChannelSelectDefaultValueResponse[];
  }
  interface ChannelSelectDefaultValue {
    type: "channel";
    id: string;
  }
  interface ChannelSelectDefaultValueResponse {
    type: "channel";
    id: string;
  }
  type ChannelTypes = 1 | 3 | 0 | 2 | 4 | 5 | 10 | 11 | 12 | 13 | 14 | 15;
  interface CommandPermissionResponse {
    id: string;
    type: discord.ApplicationCommandPermissionType;
    permission: boolean;
  }
  interface CommandPermissionsResponse {
    id: string;
    application_id: string;
    guild_id: string;
    permissions: discord.CommandPermissionResponse[];
  }
  interface ComponentEmojiForRequest {
    id?: string | null;
    name: string;
  }
  interface ComponentEmojiResponse {
    id?: string;
    name: string;
    animated?: boolean;
  }
  interface ConnectedAccountGuildResponse {
    id: string;
    name: string;
    icon: string | null;
  }
  interface ConnectedAccountIntegrationResponse {
    id: string;
    type: discord.IntegrationTypes;
    account: discord.AccountResponse;
    guild: discord.ConnectedAccountGuildResponse;
  }
  type ConnectedAccountProviders = "battlenet" | "bluesky" | "bungie" | "ebay" | "epicgames" | "facebook" | "github" | "instagram" | "mastodon" | "leagueoflegends" | "paypal" | "playstation" | "reddit" | "riotgames" | "roblox" | "skype" | "spotify" | "steam" | "tiktok" | "twitch" | "twitter" | "xbox" | "youtube" | "domain";
  interface ConnectedAccountResponse {
    id: string;
    name: string | null;
    type: discord.ConnectedAccountProviders;
    friend_sync: boolean;
    integrations?: discord.ConnectedAccountIntegrationResponse[];
    show_activity: boolean;
    two_way_link: boolean;
    verified: boolean;
    visibility: discord.ConnectedAccountVisibility;
    revoked?: boolean;
  }
  type ConnectedAccountVisibility = 0 | 1;
  interface ContainerComponentForMessageRequest {
    type: 17;
    id?: number | null;
    accent_color?: number | null;
    components: (discord.ActionRowComponentForMessageRequest | discord.FileComponentForMessageRequest | discord.MediaGalleryComponentForMessageRequest | discord.SectionComponentForMessageRequest | discord.SeparatorComponentForMessageRequest | discord.TextDisplayComponentForMessageRequest)[];
    spoiler?: boolean | null;
  }
  interface ContainerComponentResponse {
    type: 17;
    id: number;
    accent_color: number | null;
    components: (discord.ActionRowComponentResponse | discord.FileComponentResponse | discord.MediaGalleryComponentResponse | discord.SectionComponentResponse | discord.SeparatorComponentResponse | discord.TextDisplayComponentResponse)[];
    spoiler: boolean;
  }
  interface CreateForumThreadRequest {
    name: string;
    auto_archive_duration?: discord.ThreadAutoArchiveDuration | null;
    rate_limit_per_user?: number | null;
    applied_tags?: string[] | null;
    message: discord.BaseCreateMessageCreateRequest;
  }
  interface CreateGroupDMInviteRequest {
    max_age?: number | null;
  }
  interface CreateGuildChannelRequest {
    type?: 0 | 2 | 4 | 5 | 13 | 14 | 15 | null;
    name: string;
    position?: number | null;
    topic?: string | null;
    bitrate?: number | null;
    user_limit?: number | null;
    nsfw?: boolean | null;
    rate_limit_per_user?: number | null;
    parent_id?: string | null;
    permission_overwrites?: discord.ChannelPermissionOverwriteRequest[] | null;
    rtc_region?: string | null;
    video_quality_mode?: discord.VideoQualityModes | null;
    default_auto_archive_duration?: discord.ThreadAutoArchiveDuration | null;
    default_reaction_emoji?: discord.UpdateDefaultReactionEmojiRequest | null;
    default_thread_rate_limit_per_user?: number | null;
    default_sort_order?: discord.ThreadSortOrder | null;
    default_forum_layout?: discord.ForumLayout | null;
    default_tag_setting?: discord.ThreadSearchTagSetting | null;
    available_tags?: (discord.CreateOrUpdateThreadTagRequest | null)[] | null;
  }
  interface CreateGuildInviteRequest {
    max_age?: number | null;
    temporary?: boolean | null;
    max_uses?: number | null;
    unique?: boolean | null;
    target_user_id?: string | null;
    target_application_id?: string | null;
    target_type?: 1 | 2 | null;
    role_ids?: string | string[] | null;
    /** The IDs of the users to target with this invite. */
    target_user_ids?: string | string[] | null;
  }
  interface CreateOrUpdateThreadTagRequest {
    name: string;
    emoji_id?: string | null;
    emoji_name?: string | null;
    moderated?: boolean | null;
  }
  interface CreatePrivateChannelRequest {
    recipient_id?: string | null;
    access_tokens?: string[] | null;
    nicks?: Record<string, string | null> | null;
  }
  interface CreateRoleRequest {
    name?: string | null;
    permissions?: number | null;
    color?: number | null;
    colors?: discord.RoleColors | null;
    hoist?: boolean | null;
    mentionable?: boolean | null;
    icon?: string | null;
    unicode_emoji?: string | null;
  }
  interface CreateTextThreadWithMessageRequest {
    name: string;
    auto_archive_duration?: discord.ThreadAutoArchiveDuration | null;
    rate_limit_per_user?: number | null;
  }
  interface CreateTextThreadWithoutMessageRequest {
    name: string;
    auto_archive_duration?: discord.ThreadAutoArchiveDuration | null;
    rate_limit_per_user?: number | null;
    type?: 10 | 11 | 12 | null;
    invitable?: boolean | null;
  }
  interface CreatedThreadResponse {
    id: string;
    type: 10 | 11 | 12;
    last_message_id?: string | null;
    flags: number;
    last_pin_timestamp?: string | null;
    guild_id: string;
    name: string;
    parent_id?: string | null;
    rate_limit_per_user?: number;
    bitrate?: number;
    user_limit?: number;
    rtc_region?: string | null;
    video_quality_mode?: discord.VideoQualityModes;
    permissions?: string;
    owner_id: string;
    thread_metadata: discord.ThreadMetadataResponse;
    message_count: number;
    member_count: number;
    total_message_sent: number;
    applied_tags?: string[];
    member?: discord.ThreadMemberResponse;
  }
  interface CustomClientThemeResponse {
    colors: string[];
    gradient_angle: number;
    base_mix: number;
    base_theme: discord.MessageShareCustomUserThemeBaseTheme;
  }
  interface CustomClientThemeShareRequest {
    colors: string[];
    gradient_angle: number;
    base_mix: number;
    base_theme?: discord.MessageShareCustomUserThemeBaseTheme | null;
  }
  interface DefaultKeywordListTriggerMetadata {
    allow_list?: string[] | null;
    presets?: discord.AutomodKeywordPresetType[] | null;
  }
  interface DefaultKeywordListTriggerMetadataResponse {
    allow_list: string[];
    presets: discord.AutomodKeywordPresetType[];
  }
  interface DefaultKeywordListUpsertRequest {
    name: string;
    event_type: discord.AutomodEventType;
    actions?: (discord.BlockMessageAction | discord.FlagToChannelAction | discord.QuarantineUserAction | discord.UserCommunicationDisabledAction)[] | null;
    enabled?: boolean | null;
    exempt_roles?: string[] | null;
    exempt_channels?: string[] | null;
    trigger_type: 4;
    trigger_metadata: discord.DefaultKeywordListTriggerMetadata;
  }
  interface DefaultKeywordListUpsertRequestPartial {
    name?: string;
    event_type?: discord.AutomodEventType;
    actions?: (discord.BlockMessageAction | discord.FlagToChannelAction | discord.QuarantineUserAction | discord.UserCommunicationDisabledAction)[] | null;
    enabled?: boolean | null;
    exempt_roles?: string[] | null;
    exempt_channels?: string[] | null;
    trigger_type?: 4;
    trigger_metadata?: discord.DefaultKeywordListTriggerMetadata;
  }
  interface DefaultKeywordRuleResponse {
    id: string;
    guild_id: string;
    creator_id: string;
    name: string;
    event_type: discord.AutomodEventType;
    actions: (discord.BlockMessageActionResponse | discord.FlagToChannelActionResponse | discord.QuarantineUserActionResponse | discord.UserCommunicationDisabledActionResponse)[];
    trigger_type: 4;
    enabled: boolean;
    exempt_roles: string[];
    exempt_channels: string[];
    trigger_metadata: discord.DefaultKeywordListTriggerMetadataResponse;
  }
  interface DefaultReactionEmojiResponse {
    emoji_id: string | null;
    emoji_name: string | null;
  }
  interface DiscordIntegrationResponse {
    type: "discord";
    name: string | null;
    account: discord.AccountResponse;
    enabled: boolean;
    id: string;
    application: discord.IntegrationApplicationResponse;
    scopes: ("applications.commands" | "bot" | "webhook.incoming")[];
    user?: discord.UserResponse;
  }
  interface EmbeddedActivityInstance {
    application_id: string;
    instance_id: string;
    launch_id: string;
    location: discord.GuildChannelLocation | discord.PrivateChannelLocation;
    users: string[];
  }
  interface EmojiResponse {
    id: string;
    name: string;
    user?: discord.UserResponse;
    roles: string[];
    require_colons: boolean;
    managed: boolean;
    animated: boolean;
    available: boolean;
  }
  interface EntitlementResponse {
    id: string;
    sku_id: string;
    application_id: string;
    user_id: string;
    guild_id?: string | null;
    deleted: boolean;
    starts_at: string | null;
    ends_at: string | null;
    type: discord.EntitlementTypes;
    fulfilled_at?: string | null;
    fulfillment_status?: discord.EntitlementTenantFulfillmentStatusResponse | null;
    consumed?: boolean;
    gifter_user_id?: string | null;
    parent_id?: string | null;
  }
  type EntitlementTenantFulfillmentStatusResponse = 0 | 1 | 2 | 3 | 4 | 5 | 6 | 7;
  type EntitlementTypes = 8 | 10;
  interface EntityMetadataExternal {
    location: string;
  }
  interface EntityMetadataExternalResponse {
    /** Location of the external event */
    location: string;
  }
  interface EntityMetadataStageInstance {
  
  }
  interface EntityMetadataStageInstanceResponse {
  
  }
  interface EntityMetadataVoice {
  
  }
  interface EntityMetadataVoiceResponse {
  
  }
  interface ExternalConnectionIntegrationResponse {
    type: "twitch" | "youtube";
    name: string | null;
    account: discord.AccountResponse;
    enabled: boolean;
    id: string;
    user: discord.UserResponse;
    revoked?: boolean;
    expire_behavior?: discord.IntegrationExpireBehaviorTypes;
    expire_grace_period?: discord.IntegrationExpireGracePeriodTypes;
    subscriber_count?: number;
    synced_at?: string;
    role_id?: string | null;
    syncing?: boolean;
    enable_emoticons?: boolean;
  }
  interface ExternalScheduledEventCreateRequest {
    name: string;
    description?: string | null;
    image?: string | null;
    scheduled_start_time: string;
    scheduled_end_time?: string | null;
    privacy_level: discord.GuildScheduledEventPrivacyLevels;
    entity_type: 3;
    channel_id?: string | null;
    /** Recurrence rule for the scheduled event */
    recurrence_rule?: discord.RecurrenceRule | null;
    entity_metadata: discord.EntityMetadataExternal;
  }
  interface ExternalScheduledEventPatchRequestPartial {
    status?: discord.GuildScheduledEventStatuses | null;
    name?: string;
    description?: string | null;
    image?: string | null;
    scheduled_start_time?: string;
    scheduled_end_time?: string | null;
    entity_type?: 3 | null;
    privacy_level?: discord.GuildScheduledEventPrivacyLevels;
    channel_id?: string | null;
    /** Recurrence rule for the scheduled event */
    recurrence_rule?: discord.RecurrenceRule | null;
    entity_metadata?: discord.EntityMetadataExternal;
  }
  interface ExternalScheduledEventResponse {
    /** ID of the scheduled event */
    id: string;
    /** ID of the guild the scheduled event belongs to */
    guild_id: string;
    /** Name of the scheduled event */
    name: string;
    /** Description of the scheduled event */
    description: string | null;
    /** Channel ID in which the scheduled event will be hosted, or null if entity type is EXTERNAL */
    channel_id: string | null;
    /** ID of the user that created the scheduled event */
    creator_id: string | null;
    /** User that created the scheduled event */
    creator?: discord.UserResponse;
    /** Cover image hash of the scheduled event */
    image: string | null;
    /** When the scheduled event will start */
    scheduled_start_time: string;
    /** When the scheduled event will end, or null if no end time */
    scheduled_end_time: string | null;
    /** Status of the scheduled event */
    status: discord.GuildScheduledEventStatuses;
    /** Type of hosting entity associated with the scheduled event */
    entity_type: 3;
    /** ID of the hosting entity associated with the scheduled event */
    entity_id: string | null;
    /** Recurrence rule for the scheduled event, or null if not recurring */
    recurrence_rule: discord.RecurrenceRuleResponse | null;
    /** Number of users subscribed to the scheduled event */
    user_count?: number;
    /** Privacy level of the scheduled event */
    privacy_level: discord.GuildScheduledEventPrivacyLevels;
    user_rsvp?: discord.ScheduledEventUserResponse | null;
    guild_scheduled_event_exceptions: discord.GuildScheduledEventExceptionResponse[];
    entity_metadata: discord.EntityMetadataExternalResponse;
  }
  interface FileComponentForMessageRequest {
    type: 13;
    id?: number | null;
    spoiler?: boolean | null;
    file: discord.UnfurledMediaRequestWithAttachmentReferenceRequired;
  }
  interface FileComponentResponse {
    type: 13;
    id: number;
    file: discord.UnfurledMediaResponse;
    name: string | null;
    size: number | null;
    spoiler: boolean;
  }
  interface FlagToChannelAction {
    type: 2;
    metadata: discord.FlagToChannelActionMetadata;
  }
  interface FlagToChannelActionMetadata {
    channel_id: string;
  }
  interface FlagToChannelActionMetadataResponse {
    channel_id: string;
  }
  interface FlagToChannelActionResponse {
    type: 2;
    metadata: discord.FlagToChannelActionMetadataResponse;
  }
  type ForumLayout = 0 | 1 | 2;
  interface ForumTagResponse {
    id: string;
    name: string;
    moderated: boolean;
    emoji_id: string | null;
    emoji_name: string | null;
  }
  interface FriendInviteResponse {
    type: 2;
    code: string;
    inviter?: discord.UserResponse;
    max_age?: number;
    created_at?: string;
    expires_at: string | null;
    friends_count?: number;
    channel: discord.InviteChannelResponse | null;
    is_contact?: boolean;
    uses?: number;
    max_uses?: number;
    flags?: number;
  }
  interface GatewayBotResponse {
    url: string;
    session_start_limit: discord.GatewayBotSessionStartLimitResponse;
    shards: number;
  }
  interface GatewayBotSessionStartLimitResponse {
    max_concurrency: number;
    remaining: number;
    reset_after: number;
    total: number;
  }
  interface GatewayResponse {
    url: string;
  }
  interface GroupDMInviteResponse {
    type: 1;
    code: string;
    inviter?: discord.UserResponse;
    max_age?: number;
    created_at?: string;
    expires_at: string | null;
    channel: discord.InviteChannelResponse;
    approximate_member_count?: number | null;
  }
  interface GuildAuditLogResponse {
    audit_log_entries: discord.AuditLogEntryResponse[];
    users: discord.UserResponse[];
    integrations: (discord.PartialDiscordIntegrationResponse | discord.PartialExternalConnectionIntegrationResponse | discord.PartialGuildSubscriptionIntegrationResponse)[];
    webhooks: (discord.ApplicationIncomingWebhookResponse | discord.ChannelFollowerWebhookResponse | discord.GuildIncomingWebhookResponse)[];
    guild_scheduled_events: (discord.ExternalScheduledEventResponse | discord.StageScheduledEventResponse | discord.VoiceScheduledEventResponse)[];
    threads: discord.ThreadResponse[];
    application_commands: discord.ApplicationCommandResponse[];
    auto_moderation_rules: (discord.DefaultKeywordRuleResponse | discord.KeywordRuleResponse | discord.MLSpamRuleResponse | discord.MentionSpamRuleResponse | discord.UserProfileRuleResponse | null)[];
  }
  interface GuildBanResponse {
    user: discord.UserResponse;
    reason: string | null;
  }
  interface GuildChannelLocation {
    id: string;
    kind: "gc";
    channel_id: string;
    guild_id: string;
  }
  interface GuildChannelResponse {
    id: string;
    type: 0 | 2 | 4 | 5 | 13 | 14 | 15;
    last_message_id?: string | null;
    flags: number;
    last_pin_timestamp?: string | null;
    guild_id: string;
    name: string;
    parent_id?: string | null;
    rate_limit_per_user?: number;
    bitrate?: number;
    user_limit?: number;
    rtc_region?: string | null;
    video_quality_mode?: discord.VideoQualityModes;
    permissions?: string;
    topic?: string | null;
    default_auto_archive_duration?: discord.ThreadAutoArchiveDuration;
    default_thread_rate_limit_per_user?: number;
    position: number;
    permission_overwrites?: discord.ChannelPermissionOverwriteResponse[];
    nsfw?: boolean;
    available_tags?: discord.ForumTagResponse[];
    default_reaction_emoji?: discord.DefaultReactionEmojiResponse | null;
    default_sort_order?: discord.ThreadSortOrder | null;
    default_forum_layout?: discord.ForumLayout;
    default_tag_setting?: discord.ThreadSearchTagSetting | null;
  }
  type GuildExplicitContentFilterTypes = 0 | 1 | 2;
  type GuildFeatures = "ANIMATED_BANNER" | "ANIMATED_ICON" | "APPLICATION_COMMAND_PERMISSIONS_V2" | "AUTO_MODERATION" | "BANNER" | "COMMUNITY" | "CREATOR_MONETIZABLE_PROVISIONAL" | "CREATOR_STORE_PAGE" | "DEVELOPER_SUPPORT_SERVER" | "DISCOVERABLE" | "FEATURABLE" | "INVITES_DISABLED" | "INVITE_SPLASH" | "MEMBER_VERIFICATION_GATE_ENABLED" | "MORE_STICKERS" | "NEWS" | "PARTNERED" | "PREVIEW_ENABLED" | "RAID_ALERTS_DISABLED" | "PRUNE_REQUIRES_ADMIN" | "ROLE_ICONS" | "ROLE_SUBSCRIPTIONS_AVAILABLE_FOR_PURCHASE" | "ROLE_SUBSCRIPTIONS_ENABLED" | "TICKETED_EVENTS_ENABLED" | "VANITY_URL" | "VERIFIED" | "VIP_REGIONS" | "WELCOME_SCREEN_ENABLED" | "OFFICIAL_GAME_GUILD";
  interface GuildHomeSettingsResponse {
    guild_id: string;
    enabled: boolean;
    welcome_message?: discord.WelcomeMessageResponse;
    new_member_actions: discord.NewMemberActionResponse[];
    resource_channels: discord.ResourceChannelResponse[];
  }
  interface GuildIncidentActionsRequest {
    /** When invites will be enabled again */
    invites_disabled_until?: string | null;
    /** When direct messages will be enabled again */
    dms_disabled_until?: string | null;
  }
  interface GuildIncidentsDataResponse {
    /** When invites get enabled again */
    invites_disabled_until: string | null;
    /** When direct messages get enabled again */
    dms_disabled_until: string | null;
  }
  interface GuildIncomingWebhookResponse {
    application_id: string | null;
    avatar: string | null;
    channel_id: string | null;
    guild_id?: string | null;
    id: string;
    name: string;
    type: 1;
    user?: discord.UserResponse;
    token?: string;
    url?: string;
  }
  interface GuildInviteResponse {
    type: 0;
    code: string;
    inviter?: discord.UserResponse;
    max_age?: number;
    created_at?: string;
    expires_at: string | null;
    is_contact?: boolean;
    flags?: number;
    guild: discord.InviteGuildResponse;
    guild_id: string;
    channel: discord.InviteChannelResponse;
    target_type?: discord.InviteTargetTypes;
    target_user?: discord.UserResponse;
    target_application?: discord.InviteApplicationResponse;
    guild_scheduled_event?: discord.ScheduledEventResponse;
    target_channel_id?: string;
    target_message_id?: string;
    uses?: number;
    max_uses?: number;
    temporary?: boolean;
    approximate_member_count?: number | null;
    approximate_presence_count?: number | null;
    is_nickname_changeable?: boolean;
    roles?: discord.InviteGuildRoleResponse[] | null;
  }
  type GuildJoinRequestApplicationStatus = "STARTED" | "SUBMITTED" | "REJECTED" | "APPROVED";
  interface GuildJoinRequestResponse {
    id: string;
    created_at: string;
    reviewed_at: string | null;
    application_status: discord.GuildJoinRequestApplicationStatus | null;
    /** Reason request was rejected. Only set when application_status is REJECTED */
    rejection_reason: string | null;
    guild_id: string;
    user_id: string;
    user?: discord.UserResponse | null;
    /** Applicant's responses on join request form */
    form_responses?: (discord.MultipleChoiceFormFieldResponse | discord.ParagraphFormFieldResponse | discord.TermsFormFieldResponse | discord.TextInputFormFieldResponse)[] | null;
    actioned_by_user?: discord.UserResponse | null;
  }
  interface GuildJoinRequestsListResponse {
    total?: number;
    guild_join_requests?: discord.GuildJoinRequestResponse[];
  }
  type GuildMFALevel = 0 | 1;
  interface GuildMemberResponse {
    /** the member's guild avatar hash */
    avatar: string | null;
    /** data for the member's guild avatar decoration */
    avatar_decoration_data?: discord.UserAvatarDecorationResponse | null;
    /** the member's guild banner hash */
    banner: string | null;
    /** when the user's timeout will expire and the user will be able to communicate in the guild again, null or a time in the past if the user is not timed out */
    communication_disabled_until: string | null;
    /** guild member flags represented as a bit set, defaults to 0 */
    flags: number;
    /** when the user joined the guild */
    joined_at: string;
    /** this user's guild nickname */
    nick: string | null;
    /** whether the user has not yet passed the guild's Membership Screening requirements */
    pending: boolean;
    /** when the user started boosting the guild */
    premium_since: string | null;
    /** array of role object ids */
    roles: string[];
    /** data for the member's collectibles */
    collectibles?: discord.UserCollectiblesResponse | null;
    /** the user this guild member represents */
    user: discord.UserResponse;
    /** whether the user is muted in voice channels */
    mute: boolean;
    /** whether the user is deafened in voice channels */
    deaf: boolean;
  }
  type GuildNSFWContentLevel = 0 | 1 | 2 | 3;
  type GuildOnboardingMode = 0 | 1;
  interface GuildOnboardingResponse {
    guild_id: string;
    prompts: discord.OnboardingPromptResponse[];
    default_channel_ids: string[];
    enabled: boolean;
    mode: discord.GuildOnboardingMode;
  }
  interface GuildPatchRequestPartial {
    name?: string;
    description?: string | null;
    region?: string | null;
    icon?: string | null;
    verification_level?: discord.VerificationLevels | null;
    default_message_notifications?: discord.UserNotificationSettings | null;
    explicit_content_filter?: discord.GuildExplicitContentFilterTypes | null;
    preferred_locale?: discord.AvailableLocalesEnum | null;
    afk_timeout?: discord.AfkTimeouts | null;
    afk_channel_id?: string | null;
    system_channel_id?: string | null;
    splash?: string | null;
    banner?: string | null;
    system_channel_flags?: number | null;
    features?: (string | null)[] | null;
    discovery_splash?: string | null;
    home_header?: string | null;
    rules_channel_id?: string | null;
    safety_alerts_channel_id?: string | null;
    public_updates_channel_id?: string | null;
    premium_progress_bar_enabled?: boolean | null;
  }
  interface GuildPreviewResponse {
    id: string;
    name: string;
    icon: string | null;
    description: string | null;
    home_header: string | null;
    splash: string | null;
    discovery_splash: string | null;
    features: discord.GuildFeatures[];
    approximate_member_count: number;
    approximate_presence_count: number;
    emojis: discord.EmojiResponse[];
    stickers: discord.GuildStickerResponse[];
  }
  interface GuildProductPurchaseResponse {
    listing_id: string;
    product_name: string;
  }
  interface GuildPruneResponse {
    pruned: number | null;
  }
  interface GuildResponse {
    id: string;
    name: string;
    icon: string | null;
    description: string | null;
    home_header: string | null;
    splash: string | null;
    discovery_splash: string | null;
    features: discord.GuildFeatures[];
    banner: string | null;
    owner_id: string;
    application_id: string | null;
    region: string;
    afk_channel_id: string | null;
    afk_timeout: discord.AfkTimeouts;
    system_channel_id: string | null;
    system_channel_flags: number;
    widget_enabled: boolean;
    widget_channel_id: string | null;
    verification_level: discord.VerificationLevels;
    roles: discord.GuildRoleResponse[];
    default_message_notifications: discord.UserNotificationSettings;
    mfa_level: discord.GuildMFALevel;
    explicit_content_filter: discord.GuildExplicitContentFilterTypes;
    max_presences: number | null;
    max_members: number;
    max_stage_video_channel_users: number;
    max_video_channel_users: number;
    vanity_url_code: string | null;
    premium_tier: discord.PremiumGuildTiers;
    premium_subscription_count: number;
    preferred_locale: discord.AvailableLocalesEnum;
    rules_channel_id: string | null;
    safety_alerts_channel_id: string | null;
    public_updates_channel_id: string | null;
    premium_progress_bar_enabled: boolean;
    premium_progress_bar_enabled_user_updated_at?: string | null;
    nsfw: boolean;
    nsfw_level: discord.GuildNSFWContentLevel;
    emojis: discord.EmojiResponse[];
    stickers: discord.GuildStickerResponse[];
    incidents_data: discord.GuildIncidentsDataResponse | null;
  }
  interface GuildRoleColorsResponse {
    primary_color: number;
    secondary_color: number | null;
    tertiary_color: number | null;
  }
  interface GuildRoleResponse {
    id: string;
    name: string;
    permissions: string;
    position: number;
    color: number;
    colors: discord.GuildRoleColorsResponse;
    hoist: boolean;
    managed: boolean;
    mentionable: boolean;
    icon: string | null;
    unicode_emoji: string | null;
    tags?: discord.GuildRoleTagsResponse;
    flags: number;
  }
  interface GuildRoleTagsResponse {
    premium_subscriber?: null;
    bot_id?: string;
    integration_id?: string;
    subscription_listing_id?: string;
    available_for_purchase?: null;
    guild_connections?: null;
  }
  type GuildScheduledEventEntityTypes = 0 | 1 | 2 | 3;
  interface GuildScheduledEventExceptionCreateRequest {
    /** Overridden start time of this occurrence */
    scheduled_start_time?: string | null;
    /** Overridden end time of this occurrence */
    scheduled_end_time?: string | null;
    /** The original start time of the occurrence to create an exception for */
    original_scheduled_start_time: string;
    /** Whether this occurrence is canceled */
    is_canceled?: boolean | null;
  }
  interface GuildScheduledEventExceptionPatchRequestPartial {
    /** Overridden start time of this occurrence */
    scheduled_start_time?: string | null;
    /** Overridden end time of this occurrence */
    scheduled_end_time?: string | null;
    /** Whether this occurrence is canceled */
    is_canceled?: boolean | null;
  }
  interface GuildScheduledEventExceptionResponse {
    /** ID of the scheduled event this exception belongs to */
    event_id: string;
    /** ID of the event exception */
    event_exception_id: string;
    /** Overridden start time of this occurrence */
    scheduled_start_time: string | null;
    /** Overridden end time of this occurrence */
    scheduled_end_time: string | null;
    /** Whether this occurrence is canceled */
    is_canceled: boolean;
  }
  type GuildScheduledEventPrivacyLevels = 2;
  type GuildScheduledEventStatuses = 1 | 2 | 3 | 4;
  type GuildScheduledEventUserResponses = 0 | 1;
  interface GuildSearchResponse {
    messages: discord.SearchMessageResponse[][];
    doing_deep_historical_index: boolean;
    total_results: number;
    threads?: discord.ThreadResponse[] | null;
    members?: discord.ThreadMemberResponse[] | null;
    documents_indexed?: number | null;
  }
  interface GuildStickerResponse {
    id: string;
    name: string;
    tags: string;
    type: 2;
    format_type: discord.StickerFormatTypes | null;
    description: string | null;
    available: boolean;
    guild_id: string;
    user?: discord.UserResponse;
  }
  interface GuildSubscriptionIntegrationResponse {
    type: "guild_subscription";
    name: string | null;
    account: discord.AccountResponse;
    enabled: boolean;
    id: string;
  }
  interface GuildTemplateChannelResponse {
    id: number | null;
    type: 0 | 2 | 4 | 15;
    name: string | null;
    position: number | null;
    topic: string | null;
    bitrate: number;
    user_limit: number;
    nsfw: boolean;
    rate_limit_per_user: number;
    parent_id: string | null;
    default_auto_archive_duration: discord.ThreadAutoArchiveDuration | null;
    permission_overwrites: discord.ChannelPermissionOverwriteResponse[];
    available_tags: discord.GuildTemplateChannelTags[] | null;
    template: string;
    default_reaction_emoji: discord.DefaultReactionEmojiResponse | null;
    default_thread_rate_limit_per_user: number | null;
    default_sort_order: discord.ThreadSortOrder | null;
    default_forum_layout: discord.ForumLayout | null;
    default_tag_setting: discord.ThreadSearchTagSetting | null;
    theme_color: number | null;
  }
  interface GuildTemplateChannelTags {
    id: number | null;
    name: string;
    emoji_id: string | null;
    emoji_name: string | null;
    moderated: boolean | null;
  }
  interface GuildTemplateResponse {
    code: string;
    name: string;
    description: string | null;
    usage_count: number;
    creator_id: string;
    creator: discord.UserResponse | null;
    created_at: string;
    updated_at: string;
    source_guild_id: string;
    serialized_source_guild: discord.GuildTemplateSnapshotResponse;
    is_dirty: boolean | null;
  }
  interface GuildTemplateRoleColorsResponse {
    primary_color: number;
    secondary_color: number | null;
    tertiary_color: number | null;
  }
  interface GuildTemplateRoleResponse {
    id: number;
    name: string;
    permissions: string;
    color: number;
    colors: discord.GuildTemplateRoleColorsResponse | null;
    hoist: boolean;
    mentionable: boolean;
    icon: string | null;
    unicode_emoji: string | null;
  }
  interface GuildTemplateSnapshotResponse {
    name: string;
    description: string | null;
    region: string | null;
    verification_level: discord.VerificationLevels;
    default_message_notifications: discord.UserNotificationSettings;
    explicit_content_filter: discord.GuildExplicitContentFilterTypes;
    preferred_locale: discord.AvailableLocalesEnum;
    afk_channel_id: string | null;
    afk_timeout: discord.AfkTimeouts;
    system_channel_id: string | null;
    system_channel_flags: number;
    roles: discord.GuildTemplateRoleResponse[];
    channels: discord.GuildTemplateChannelResponse[];
  }
  interface GuildWelcomeChannel {
    channel_id: string;
    description: string;
    emoji_id?: string | null;
    emoji_name?: string | null;
  }
  interface GuildWelcomeScreenChannelResponse {
    channel_id: string;
    description: string;
    emoji_id: string | null;
    emoji_name: string | null;
  }
  interface GuildWelcomeScreenResponse {
    description: string | null;
    welcome_channels: discord.GuildWelcomeScreenChannelResponse[];
  }
  interface GuildWithCountsResponse {
    id: string;
    name: string;
    icon: string | null;
    description: string | null;
    home_header: string | null;
    splash: string | null;
    discovery_splash: string | null;
    features: discord.GuildFeatures[];
    banner: string | null;
    owner_id: string;
    application_id: string | null;
    region: string;
    afk_channel_id: string | null;
    afk_timeout: discord.AfkTimeouts;
    system_channel_id: string | null;
    system_channel_flags: number;
    widget_enabled: boolean;
    widget_channel_id: string | null;
    verification_level: discord.VerificationLevels;
    roles: discord.GuildRoleResponse[];
    default_message_notifications: discord.UserNotificationSettings;
    mfa_level: discord.GuildMFALevel;
    explicit_content_filter: discord.GuildExplicitContentFilterTypes;
    max_presences: number | null;
    max_members: number;
    max_stage_video_channel_users: number;
    max_video_channel_users: number;
    vanity_url_code: string | null;
    premium_tier: discord.PremiumGuildTiers;
    premium_subscription_count: number;
    preferred_locale: discord.AvailableLocalesEnum;
    rules_channel_id: string | null;
    safety_alerts_channel_id: string | null;
    public_updates_channel_id: string | null;
    premium_progress_bar_enabled: boolean;
    premium_progress_bar_enabled_user_updated_at?: string | null;
    nsfw: boolean;
    nsfw_level: discord.GuildNSFWContentLevel;
    emojis: discord.EmojiResponse[];
    stickers: discord.GuildStickerResponse[];
    incidents_data: discord.GuildIncidentsDataResponse | null;
    approximate_member_count?: number | null;
    approximate_presence_count?: number | null;
  }
  interface IntegrationApplicationResponse {
    id: string;
    name: string;
    icon: string | null;
    description: string;
    type: discord.ApplicationTypes | null;
    cover_image?: string;
    primary_sku_id?: string;
    flags: number;
    flags_new: string;
    bot?: discord.UserResponse;
  }
  type IntegrationExpireBehaviorTypes = 0 | 1;
  type IntegrationExpireGracePeriodTypes = 1 | 3 | 7 | 14 | 30;
  type IntegrationTypes = "discord" | "twitch" | "youtube" | "guild_subscription";
  type InteractionContextType = 0 | 1 | 2;
  type InteractionTypes = 1 | 2 | 3 | 4 | 5 | 6;
  interface InviteApplicationResponse {
    id: string;
    name: string;
    icon: string | null;
    description: string;
    type: discord.ApplicationTypes | null;
    cover_image?: string;
    primary_sku_id?: string;
    flags: number;
    flags_new: string;
    bot?: discord.UserResponse;
    slug?: string;
    vibegrations_project_id?: string;
    guild_id?: string;
    rpc_origins?: string[];
    bot_public?: boolean;
    bot_require_code_grant?: boolean;
    terms_of_service_url?: string;
    privacy_policy_url?: string;
    custom_install_url?: string;
    install_params?: discord.ApplicationOAuth2InstallParamsResponse;
    integration_types_config?: Record<string, discord.ApplicationIntegrationTypeConfigurationResponse>;
    verify_key: string;
    max_participants?: number | null;
    tags?: string[];
  }
  interface InviteChannelRecipientResponse {
    id: string;
    username: string;
    avatar: string | null;
  }
  interface InviteChannelResponse {
    id: string;
    type: discord.ChannelTypes;
    name: string | null;
    icon?: string;
    nsfw?: boolean;
    recipients?: discord.InviteChannelRecipientResponse[];
  }
  interface InviteGuildResponse {
    id: string;
    name: string;
    splash: string | null;
    banner: string | null;
    description: string | null;
    icon: string | null;
    features: discord.GuildFeatures[];
    verification_level: discord.VerificationLevels | null;
    vanity_url_code: string | null;
    nsfw_level: discord.GuildNSFWContentLevel | null;
    nsfw: boolean | null;
    premium_subscription_count: number;
  }
  interface InviteGuildRoleResponse {
    id: string;
    name: string;
    position: number;
    color: number;
    colors: discord.GuildRoleColorsResponse;
    icon: string | null;
    unicode_emoji: string | null;
    permissions?: string;
  }
  type InviteTargetTypes = 1 | 2 | 3;
  interface KeywordRuleResponse {
    id: string;
    guild_id: string;
    creator_id: string;
    name: string;
    event_type: discord.AutomodEventType;
    actions: (discord.BlockMessageActionResponse | discord.FlagToChannelActionResponse | discord.QuarantineUserActionResponse | discord.UserCommunicationDisabledActionResponse)[];
    trigger_type: 1;
    enabled: boolean;
    exempt_roles: string[];
    exempt_channels: string[];
    trigger_metadata: discord.KeywordTriggerMetadataResponse;
  }
  interface KeywordTriggerMetadata {
    keyword_filter?: string[] | null;
    regex_patterns?: string[] | null;
    allow_list?: string[] | null;
  }
  interface KeywordTriggerMetadataResponse {
    keyword_filter: string[];
    regex_patterns: string[];
    allow_list: string[];
  }
  interface KeywordUpsertRequest {
    name: string;
    event_type: discord.AutomodEventType;
    actions?: (discord.BlockMessageAction | discord.FlagToChannelAction | discord.QuarantineUserAction | discord.UserCommunicationDisabledAction)[] | null;
    enabled?: boolean | null;
    exempt_roles?: string[] | null;
    exempt_channels?: string[] | null;
    trigger_type: 1;
    trigger_metadata?: discord.KeywordTriggerMetadata | null;
  }
  interface KeywordUpsertRequestPartial {
    name?: string;
    event_type?: discord.AutomodEventType;
    actions?: (discord.BlockMessageAction | discord.FlagToChannelAction | discord.QuarantineUserAction | discord.UserCommunicationDisabledAction)[] | null;
    enabled?: boolean | null;
    exempt_roles?: string[] | null;
    exempt_channels?: string[] | null;
    trigger_type?: 1;
    trigger_metadata?: discord.KeywordTriggerMetadata | null;
  }
  interface ListApplicationEmojisResponse {
    items: discord.EmojiResponse[];
  }
  interface ListGuildSoundboardSoundsResponse {
    items: discord.SoundboardSoundResponse[];
  }
  interface LobbyGuildInviteResponse {
    code: string;
  }
  interface LobbyMemberRequest {
    id: string;
    metadata?: Record<string, string> | null;
    flags?: 1 | null;
    additional_name?: string | null;
  }
  interface LobbyMemberResponse {
    id: string;
    metadata: Record<string, string> | null;
    flags: number;
    additional_name?: string;
  }
  interface LobbyMessageResponse {
    id: string;
    type: discord.MessageType;
    content: string;
    lobby_id: string;
    channel_id: string;
    author: discord.UserResponse;
    lobby_member?: discord.MessageLobbyMemberResponse;
    metadata?: Record<string, string>;
    moderation_metadata?: Record<string, string>;
    flags: number;
    application_id?: string;
  }
  interface LobbyResponse {
    id: string;
    application_id: string;
    metadata: Record<string, string> | null;
    members: discord.LobbyMemberResponse[];
    linked_channel?: discord.GuildChannelResponse;
    flags: number;
    override_event_webhooks_url?: string | null;
  }
  interface MLSpamRuleResponse {
    id: string;
    guild_id: string;
    creator_id: string;
    name: string;
    event_type: discord.AutomodEventType;
    actions: (discord.BlockMessageActionResponse | discord.FlagToChannelActionResponse | discord.QuarantineUserActionResponse | discord.UserCommunicationDisabledActionResponse)[];
    trigger_type: 3;
    enabled: boolean;
    exempt_roles: string[];
    exempt_channels: string[];
    trigger_metadata: discord.MLSpamTriggerMetadataResponse;
  }
  interface MLSpamTriggerMetadata {
  
  }
  interface MLSpamTriggerMetadataResponse {
  
  }
  interface MLSpamUpsertRequest {
    name: string;
    event_type: discord.AutomodEventType;
    actions?: (discord.BlockMessageAction | discord.FlagToChannelAction | discord.QuarantineUserAction | discord.UserCommunicationDisabledAction)[] | null;
    enabled?: boolean | null;
    exempt_roles?: string[] | null;
    exempt_channels?: string[] | null;
    trigger_type: 3;
    trigger_metadata?: discord.MLSpamTriggerMetadata | null;
  }
  interface MLSpamUpsertRequestPartial {
    name?: string;
    event_type?: discord.AutomodEventType;
    actions?: (discord.BlockMessageAction | discord.FlagToChannelAction | discord.QuarantineUserAction | discord.UserCommunicationDisabledAction)[] | null;
    enabled?: boolean | null;
    exempt_roles?: string[] | null;
    exempt_channels?: string[] | null;
    trigger_type?: 3;
    trigger_metadata?: discord.MLSpamTriggerMetadata | null;
  }
  interface MediaGalleryComponentForMessageRequest {
    type: 12;
    id?: number | null;
    items: discord.MediaGalleryItemRequest[];
  }
  interface MediaGalleryComponentResponse {
    type: 12;
    id: number;
    items: discord.MediaGalleryItemResponse[];
  }
  interface MediaGalleryItemRequest {
    description?: string | null;
    spoiler?: boolean | null;
    media: discord.UnfurledMediaRequest;
  }
  interface MediaGalleryItemResponse {
    media: discord.UnfurledMediaResponse;
    description: string | null;
    spoiler: boolean;
  }
  interface MentionSpamRuleResponse {
    id: string;
    guild_id: string;
    creator_id: string;
    name: string;
    event_type: discord.AutomodEventType;
    actions: (discord.BlockMessageActionResponse | discord.FlagToChannelActionResponse | discord.QuarantineUserActionResponse | discord.UserCommunicationDisabledActionResponse)[];
    trigger_type: 5;
    enabled: boolean;
    exempt_roles: string[];
    exempt_channels: string[];
    trigger_metadata: discord.MentionSpamTriggerMetadataResponse;
  }
  interface MentionSpamTriggerMetadata {
    mention_total_limit?: number | null;
    mention_raid_protection_enabled?: boolean | null;
  }
  interface MentionSpamTriggerMetadataResponse {
    mention_total_limit: number;
    mention_raid_protection_enabled: boolean;
  }
  interface MentionSpamUpsertRequest {
    name: string;
    event_type: discord.AutomodEventType;
    actions?: (discord.BlockMessageAction | discord.FlagToChannelAction | discord.QuarantineUserAction | discord.UserCommunicationDisabledAction)[] | null;
    enabled?: boolean | null;
    exempt_roles?: string[] | null;
    exempt_channels?: string[] | null;
    trigger_type: 5;
    trigger_metadata?: discord.MentionSpamTriggerMetadata | null;
  }
  interface MentionSpamUpsertRequestPartial {
    name?: string;
    event_type?: discord.AutomodEventType;
    actions?: (discord.BlockMessageAction | discord.FlagToChannelAction | discord.QuarantineUserAction | discord.UserCommunicationDisabledAction)[] | null;
    enabled?: boolean | null;
    exempt_roles?: string[] | null;
    exempt_channels?: string[] | null;
    trigger_type?: 5;
    trigger_metadata?: discord.MentionSpamTriggerMetadata | null;
  }
  interface MentionableSelectComponentForMessageRequest {
    type: 7;
    id?: number | null;
    custom_id: string;
    placeholder?: string | null;
    min_values?: number | null;
    max_values?: number | null;
    disabled?: boolean | null;
    required?: boolean | null;
    default_values?: (discord.RoleSelectDefaultValue | discord.UserSelectDefaultValue)[] | null;
  }
  interface MentionableSelectComponentResponse {
    type: 7;
    id: number;
    custom_id: string;
    placeholder?: string;
    min_values: number;
    max_values: number;
    disabled?: boolean;
    default_values?: (discord.RoleSelectDefaultValueResponse | discord.UserSelectDefaultValueResponse)[];
  }
  interface MessageActivityResponse {
    type: discord.ActivityActionTypes;
    party_id?: string;
  }
  interface MessageAllowedMentionsRequest {
    parse?: (discord.AllowedMentionTypes | null)[] | null;
    users?: (string | null)[] | null;
    roles?: (string | null)[] | null;
    replied_user?: boolean | null;
  }
  interface MessageAttachmentRequest {
    id: string;
    filename?: string | null;
    description?: string | null;
    duration_secs?: number | null;
    waveform?: string | null;
    title?: string | null;
    is_spoiler?: boolean | null;
  }
  interface MessageAttachmentResponse {
    id: string;
    filename: string;
    size: number;
    url: string;
    proxy_url: string;
    width?: number;
    height?: number;
    duration_secs?: number;
    waveform?: string;
    description?: string;
    content_type?: string;
    ephemeral?: boolean;
    flags?: number;
    placeholder?: string;
    placeholder_version?: number;
    title?: string | null;
    application?: discord.ApplicationResponse;
    clip_created_at?: string;
    clip_participants?: discord.UserResponse[];
  }
  interface MessageCallResponse {
    ended_timestamp?: string | null;
    participants: string[];
  }
  interface MessageComponentInteractionMetadataResponse {
    id: string;
    type: 3;
    user?: discord.UserResponse;
    authorizing_integration_owners: Record<string, string>;
    original_response_message_id?: string;
    interacted_message_id: string;
  }
  type MessageComponentSeparatorSpacingSize = 1 | 2;
  interface MessageCreateRequest {
    content?: string | null;
    embeds?: discord.RichEmbed[] | null;
    allowed_mentions?: discord.MessageAllowedMentionsRequest | null;
    sticker_ids?: string[] | null;
    components?: (discord.ActionRowComponentForMessageRequest | discord.ContainerComponentForMessageRequest | discord.FileComponentForMessageRequest | discord.MediaGalleryComponentForMessageRequest | discord.SectionComponentForMessageRequest | discord.SeparatorComponentForMessageRequest | discord.TextDisplayComponentForMessageRequest)[] | null;
    flags?: number | null;
    attachments?: discord.MessageAttachmentRequest[] | null;
    poll?: discord.PollCreateRequest | null;
    shared_client_theme?: discord.CustomClientThemeShareRequest | null;
    message_reference?: discord.MessageReferenceRequest | null;
    nonce?: number | string | null;
    enforce_nonce?: boolean | null;
    tts?: boolean | null;
  }
  interface MessageEditRequestPartial {
    content?: string | null;
    embeds?: discord.RichEmbed[] | null;
    flags?: number | null;
    allowed_mentions?: discord.MessageAllowedMentionsRequest | null;
    sticker_ids?: string[] | null;
    components?: (discord.ActionRowComponentForMessageRequest | discord.ContainerComponentForMessageRequest | discord.FileComponentForMessageRequest | discord.MediaGalleryComponentForMessageRequest | discord.SectionComponentForMessageRequest | discord.SeparatorComponentForMessageRequest | discord.TextDisplayComponentForMessageRequest)[] | null;
    attachments?: discord.MessageAttachmentRequest[] | null;
  }
  interface MessageEmbedAuthorResponse {
    name: string;
    url?: string;
    icon_url?: string;
    proxy_icon_url?: string;
  }
  interface MessageEmbedFieldResponse {
    name: string;
    value: string;
    inline: boolean;
  }
  interface MessageEmbedFooterResponse {
    text: string;
    icon_url?: string;
    proxy_icon_url?: string;
  }
  interface MessageEmbedImageResponse {
    url?: string;
    proxy_url?: string;
    width?: number;
    height?: number;
    content_type?: string;
    placeholder?: string;
    placeholder_version?: number;
    description?: string;
    flags?: number;
  }
  interface MessageEmbedProviderResponse {
    name: string;
    url?: string;
  }
  interface MessageEmbedResponse {
    type: string;
    url?: string;
    title?: string;
    description?: string;
    color?: number;
    timestamp?: string;
    fields?: discord.MessageEmbedFieldResponse[];
    author?: discord.MessageEmbedAuthorResponse;
    provider?: discord.MessageEmbedProviderResponse;
    image?: discord.MessageEmbedImageResponse;
    thumbnail?: discord.MessageEmbedImageResponse;
    video?: discord.MessageEmbedVideoResponse;
    footer?: discord.MessageEmbedFooterResponse;
    flags?: number | null;
    components?: discord.ContainerComponentResponse[];
  }
  interface MessageEmbedVideoResponse {
    url?: string;
    proxy_url?: string;
    width?: number;
    height?: number;
    content_type?: string;
    placeholder?: string;
    placeholder_version?: number;
    description?: string;
    flags?: number;
  }
  interface MessageInteractionResponse {
    id: string;
    type: discord.InteractionTypes;
    name: string;
    user?: discord.UserResponse;
    name_localized?: string;
  }
  interface MessageLobbyMemberResponse {
    additional_name: string;
  }
  interface MessageMentionChannelResponse {
    id: string;
    name: string;
    type: discord.ChannelTypes;
    guild_id: string;
  }
  interface MessageReactionCountDetailsResponse {
    burst: number;
    normal: number;
  }
  interface MessageReactionEmojiResponse {
    id: string | null;
    name: string | null;
    animated?: boolean;
  }
  interface MessageReactionResponse {
    emoji: discord.MessageReactionEmojiResponse;
    count: number;
    count_details: discord.MessageReactionCountDetailsResponse;
    burst_colors: string[];
    me_burst: boolean;
    me: boolean;
  }
  interface MessageReferenceRequest {
    guild_id?: string | null;
    channel_id?: string | null;
    message_id: string;
    fail_if_not_exists?: boolean | null;
    type?: discord.MessageReferenceType | null;
  }
  interface MessageReferenceResponse {
    type: discord.MessageReferenceType;
    channel_id: string;
    message_id?: string;
    guild_id?: string;
  }
  type MessageReferenceType = 0;
  interface MessageResponse {
    type: discord.MessageType;
    content: string;
    mentions: discord.UserResponse[];
    mention_roles: string[];
    attachments: discord.MessageAttachmentResponse[];
    embeds: discord.MessageEmbedResponse[];
    timestamp: string;
    edited_timestamp: string | null;
    flags: number;
    components: (discord.ActionRowComponentResponse | discord.ContainerComponentResponse | discord.FileComponentResponse | discord.MediaGalleryComponentResponse | discord.SectionComponentResponse | discord.SeparatorComponentResponse | discord.TextDisplayComponentResponse)[];
    stickers?: (discord.GuildStickerResponse | discord.StandardStickerResponse)[];
    sticker_items?: discord.MessageStickerItemResponse[];
    id: string;
    channel_id: string;
    author: discord.UserResponse;
    pinned: boolean;
    mention_everyone: boolean;
    tts: boolean;
    call?: discord.MessageCallResponse;
    activity?: discord.MessageActivityResponse;
    application?: discord.BasicApplicationResponseWithBot;
    application_id?: string;
    interaction?: discord.MessageInteractionResponse;
    nonce?: number | string;
    webhook_id?: string;
    message_reference?: discord.MessageReferenceResponse;
    thread?: discord.ThreadResponse;
    mention_channels?: discord.MessageMentionChannelResponse[];
    role_subscription_data?: discord.MessageRoleSubscriptionDataResponse;
    purchase_notification?: discord.PurchaseNotificationResponse;
    position?: number;
    resolved?: discord.ResolvedObjectsResponse;
    poll?: discord.PollResponse;
    shared_client_theme?: discord.CustomClientThemeResponse;
    interaction_metadata?: discord.ApplicationCommandInteractionMetadataResponse | discord.MessageComponentInteractionMetadataResponse | discord.ModalSubmitInteractionMetadataResponse;
    message_snapshots?: discord.MessageSnapshotResponse[];
    lobby_member?: discord.MessageLobbyMemberResponse;
    reactions?: discord.MessageReactionResponse[];
    referenced_message?: discord.BasicMessageResponse | null;
  }
  interface MessageRoleSubscriptionDataResponse {
    role_subscription_listing_id: string;
    tier_name: string;
    total_months_subscribed: number;
    is_renewal: boolean;
  }
  type MessageShareCustomUserThemeBaseTheme = 0 | 1 | 2 | 3 | 4;
  interface MessageSnapshotResponse {
    message: discord.MinimalContentMessageResponse;
  }
  interface MessageStickerItemResponse {
    id: string;
    name: string;
    format_type: discord.StickerFormatTypes;
  }
  type MessageType = 0 | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9 | 10 | 11 | 12 | 14 | 15 | 16 | 17 | 18 | 19 | 20 | 21 | 22 | 23 | 24 | 25 | 26 | 27 | 28 | 29 | 31 | 32 | 36 | 37 | 38 | 39 | 46 | 55;
  type MetadataItemTypes = 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8;
  interface MinimalContentMessageResponse {
    type: discord.MessageType;
    content: string;
    mentions: discord.UserResponse[];
    mention_roles: string[];
    attachments: discord.MessageAttachmentResponse[];
    embeds: discord.MessageEmbedResponse[];
    timestamp: string;
    edited_timestamp: string | null;
    flags: number;
    components: (discord.ActionRowComponentResponse | discord.ContainerComponentResponse | discord.FileComponentResponse | discord.MediaGalleryComponentResponse | discord.SectionComponentResponse | discord.SeparatorComponentResponse | discord.TextDisplayComponentResponse)[];
    stickers?: (discord.GuildStickerResponse | discord.StandardStickerResponse)[];
    sticker_items?: discord.MessageStickerItemResponse[];
  }
  interface ModalSubmitInteractionMetadataResponse {
    id: string;
    type: 5;
    user?: discord.UserResponse;
    authorizing_integration_owners: Record<string, string>;
    original_response_message_id?: string;
    triggering_interaction_metadata: discord.ApplicationCommandInteractionMetadataResponse | discord.MessageComponentInteractionMetadataResponse;
  }
  interface MultipleChoiceFormFieldResponse {
    /** Type of form field */
    field_type: "MULTIPLE_CHOICE";
    /** Label shown above field */
    label?: string;
    /** Optional helper text shown below label */
    description?: string;
    /** Whether applicant must fill in field */
    required?: boolean;
    /** Choices applicant can select from */
    choices: string[];
    /** Index of choice selected by applicant */
    response?: number;
  }
  interface MyGuildResponse {
    id: string;
    name: string;
    icon: string | null;
    banner: string | null;
    owner: boolean;
    permissions: string;
    features: discord.GuildFeatures[];
    approximate_member_count?: number | null;
    approximate_presence_count?: number | null;
  }
  type NameplatePalette = "crimson" | "berry" | "sky" | "teal" | "forest" | "bubble_gum" | "violet" | "cobalt" | "clover" | "lemon" | "white" | "black";
  interface NewMemberActionResponse {
    channel_id: string;
    action_type: discord.NewMemberActionType;
    title: string;
    description: string;
    emoji?: discord.SettingsEmojiResponse;
    icon?: string;
  }
  type NewMemberActionType = 0 | 1;
  interface OAuth2GetAuthorizationResponse {
    application: discord.ApplicationResponse;
    expires: string;
    scopes: discord.OAuth2Scopes[];
    user?: discord.UserResponse;
  }
  interface OAuth2GetKeys {
    keys: discord.OAuth2Key[];
  }
  interface OAuth2GetOpenIDConnectUserInfoResponse {
    sub: string;
    email?: string | null;
    email_verified?: boolean;
    preferred_username?: string;
    nickname?: string | null;
    picture?: string;
    locale?: string;
  }
  interface OAuth2Key {
    kty: string;
    use: string;
    kid: string;
    n: string;
    e: string;
    alg: string;
  }
  type OAuth2Scopes = "identify" | "email" | "connections" | "guilds" | "guilds.join" | "guilds.members.read" | "gdm.join" | "bot" | "rpc" | "rpc.notifications.read" | "rpc.voice.read" | "rpc.voice.write" | "rpc.video.read" | "rpc.video.write" | "rpc.screenshare.read" | "rpc.screenshare.write" | "rpc.activities.write" | "webhook.incoming" | "messages.read" | "applications.builds.upload" | "applications.builds.read" | "applications.commands" | "applications.commands.permissions.update" | "applications.commands.update" | "applications.store.update" | "applications.entitlements" | "activities.read" | "activities.write" | "activities.invites.write" | "relationships.read" | "voice" | "dm_channels.read" | "role_connections.write" | "openid";
  interface OnboardingPromptOptionRequest {
    id?: string | null;
    title: string;
    description?: string | null;
    emoji_id?: string | null;
    emoji_name?: string | null;
    emoji_animated?: boolean | null;
    role_ids?: string[] | null;
    channel_ids?: string[] | null;
  }
  interface OnboardingPromptOptionResponse {
    id: string;
    title: string;
    description: string;
    emoji: discord.SettingsEmojiResponse;
    role_ids: string[];
    channel_ids: string[];
  }
  interface OnboardingPromptResponse {
    id: string;
    title: string;
    options: discord.OnboardingPromptOptionResponse[];
    single_select: boolean;
    required: boolean;
    in_onboarding: boolean;
    type: discord.OnboardingPromptType;
  }
  type OnboardingPromptType = 0 | 1;
  interface ParagraphFormFieldResponse {
    /** Type of form field */
    field_type: "PARAGRAPH";
    /** Label shown above field */
    label?: string;
    /** Optional helper text shown below label */
    description?: string;
    /** Whether applicant must fill in field */
    required?: boolean;
    /** Placeholder text shown in empty input */
    placeholder?: string;
    /** Applicant's text response */
    response?: string;
  }
  interface PartialDiscordIntegrationResponse {
    id: string;
    type: "discord";
    name: string | null;
    account: discord.AccountResponse;
    application_id: string;
  }
  interface PartialExternalConnectionIntegrationResponse {
    id: string;
    type: "twitch" | "youtube";
    name: string | null;
    account: discord.AccountResponse;
  }
  interface PartialGuildSubscriptionIntegrationResponse {
    id: string;
    type: "guild_subscription";
    name: string | null;
    account: discord.AccountResponse;
  }
  interface PinnedMessageResponse {
    pinned_at: string;
    message: discord.MessageResponse;
  }
  interface PinnedMessagesResponse {
    items: discord.PinnedMessageResponse[];
    has_more: boolean;
  }
  interface PollAnswerCreateRequest {
    /** The data of the answer */
    poll_media: discord.PollMediaCreateRequest;
  }
  interface PollAnswerDetailsResponse {
    users: discord.UserResponse[];
  }
  interface PollAnswerResponse {
    /** The ID of the answer */
    answer_id: number;
    /** The data of the answer */
    poll_media: discord.PollMediaResponse;
  }
  interface PollCreateRequest {
    /** The question of the poll. Only `text` is supported. */
    question: discord.PollMedia;
    /** Each of the answers available in the poll, up to 10 */
    answers: discord.PollAnswerCreateRequest[];
    /** Whether a user can select multiple answers */
    allow_multiselect?: boolean | null;
    /** The layout type of the poll. Defaults to... DEFAULT! */
    layout_type?: discord.PollLayoutTypes | null;
    /** Number of hours the poll should be open for, up to 32 days. Defaults to 24 */
    duration?: number | null;
  }
  interface PollEmoji {
    /** The ID of the custom emoji */
    id?: string | null;
    /** The name of the emoji, or the unicode emoji character */
    name?: string | null;
    /** Whether the emoji is animated */
    animated?: boolean | null;
  }
  interface PollEmojiCreateRequest {
    /** The ID of the custom emoji */
    id?: string | null;
    /** The name of the emoji, or the unicode emoji character */
    name?: string | null;
    /** Whether the emoji is animated */
    animated?: boolean | null;
  }
  type PollLayoutTypes = 1;
  interface PollMedia {
    /** The text of the field */
    text?: string | null;
    /** The emoji of the field */
    emoji?: discord.PollEmoji | null;
  }
  interface PollMediaCreateRequest {
    /** The text of the field */
    text?: string | null;
    /** The emoji of the field */
    emoji?: discord.PollEmojiCreateRequest | null;
  }
  interface PollMediaResponse {
    /** The text of the field */
    text?: string;
    /** The emoji of the field */
    emoji?: discord.MessageReactionEmojiResponse;
  }
  interface PollResponse {
    /** The question of the poll. Only `text` is supported. */
    question: discord.PollMediaResponse;
    /** Each of the answers available in the poll */
    answers: discord.PollAnswerResponse[];
    /** The time when the poll ends */
    expiry: string;
    /** Whether a user can select multiple answers */
    allow_multiselect: boolean;
    /** The layout type of the poll */
    layout_type: discord.PollLayoutTypes;
    /** The results of the poll */
    results: discord.PollResultsResponse;
  }
  interface PollResultsEntryResponse {
    /** The answer_id */
    id: number;
    /** The number of votes for this answer */
    count: number;
    /** Whether the current user voted for this answer */
    me_voted: boolean;
  }
  interface PollResultsResponse {
    /** The counts for each answer */
    answer_counts: discord.PollResultsEntryResponse[];
    /** Whether the votes have been precisely counted */
    is_finalized: boolean;
  }
  type PremiumGuildTiers = 0 | 1 | 2 | 3;
  type PremiumTypes = 0 | 1 | 2 | 3;
  interface PrivateApplicationResponse {
    id: string;
    name: string;
    icon: string | null;
    description: string;
    type: discord.ApplicationTypes | null;
    cover_image?: string;
    primary_sku_id?: string;
    flags: number;
    flags_new: string;
    bot?: discord.UserResponse;
    slug?: string;
    vibegrations_project_id?: string;
    guild_id?: string;
    rpc_origins?: string[];
    bot_public?: boolean;
    bot_require_code_grant?: boolean;
    terms_of_service_url?: string;
    privacy_policy_url?: string;
    custom_install_url?: string;
    install_params?: discord.ApplicationOAuth2InstallParamsResponse;
    integration_types_config?: Record<string, discord.ApplicationIntegrationTypeConfigurationResponse>;
    verify_key: string;
    max_participants?: number | null;
    tags?: string[];
    redirect_uris: string[];
    interactions_endpoint_url: string | null;
    role_connections_verification_url: string | null;
    owner: discord.UserResponse;
    approximate_guild_count: number;
    approximate_user_install_count: number;
    approximate_user_authorization_count: number;
    event_webhooks_url?: string | null;
    event_webhooks_status?: discord.ApplicationEventWebhooksStatus;
    event_webhooks_types?: ("APPLICATION_AUTHORIZED" | "APPLICATION_DEAUTHORIZED" | "ENTITLEMENT_CREATE" | "ENTITLEMENT_DELETE" | "ENTITLEMENT_UPDATE" | "GAME_DIRECT_MESSAGE_CREATE" | "GAME_DIRECT_MESSAGE_DELETE" | "GAME_DIRECT_MESSAGE_UPDATE" | "LOBBY_MESSAGE_CREATE" | "LOBBY_MESSAGE_DELETE" | "LOBBY_MESSAGE_UPDATE" | "QUEST_USER_ENROLLMENT")[];
    explicit_content_filter: discord.ApplicationExplicitContentFilterTypes;
    team: discord.TeamResponse | null;
    eligible_oauth2_scopes: discord.OAuth2Scopes[];
  }
  interface PrivateChannelLocation {
    id: string;
    kind: "pc";
    channel_id: string;
  }
  interface PrivateChannelResponse {
    id: string;
    type: 1;
    last_message_id?: string | null;
    flags: number;
    last_pin_timestamp?: string | null;
    recipients: discord.UserResponse[];
  }
  interface PrivateGroupChannelResponse {
    id: string;
    type: 3;
    last_message_id?: string | null;
    flags: number;
    last_pin_timestamp?: string | null;
    recipients: discord.UserResponse[];
    name: string | null;
    icon: string | null;
    owner_id: string;
    managed?: boolean;
    application_id?: string;
  }
  interface ProvisionalTokenResponse {
    token_type: string;
    access_token: string;
    expires_in: number;
    scope: string;
    id_token: string;
    refresh_token?: string | null;
    scopes?: string[] | null;
    expires_at_s?: number | null;
  }
  interface PurchaseNotificationResponse {
    type: discord.PurchaseType;
    guild_product_purchase?: discord.GuildProductPurchaseResponse;
  }
  type PurchaseType = 0;
  interface QuarantineUserAction {
    type: 4;
    metadata?: discord.QuarantineUserActionMetadata | null;
  }
  interface QuarantineUserActionMetadata {
  
  }
  interface QuarantineUserActionMetadataResponse {
  
  }
  interface QuarantineUserActionResponse {
    type: 4;
    metadata: discord.QuarantineUserActionMetadataResponse;
  }
  interface RecurrenceRule {
    /** Starting time of the recurrence interval */
    start: string;
    /** Ending time of the recurrence interval */
    end?: string | null;
    /** How often the event occurs */
    frequency: discord.RecurrenceRuleFrequencies;
    /** The spacing between events, defined by frequency */
    interval?: number | null;
    /** Set of specific days within a week for the event to recur on */
    by_weekday?: discord.RecurrenceRuleWeekdays[] | null;
    /** List of specific days within a specific week to recur on */
    by_n_weekday?: discord.ByNWeekday[] | null;
    /** Set of specific months to recur on */
    by_month?: discord.RecurrenceRuleMonths[] | null;
    /** Set of specific dates within a month to recur on */
    by_month_day?: number[] | null;
    /** Set of days within a year to recur on (1-364) */
    by_year_day?: number[] | null;
    /** Total number of times the event is allowed to recur */
    count?: number | null;
  }
  type RecurrenceRuleFrequencies = 3 | 2 | 1 | 0;
  type RecurrenceRuleMonths = 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9 | 10 | 11 | 12;
  interface RecurrenceRuleResponse {
    /** Starting time of the recurrence interval */
    start: string;
    /** Ending time of the recurrence interval */
    end?: string | null;
    /** How often the event occurs */
    frequency: discord.RecurrenceRuleFrequencies;
    /** The spacing between events, defined by frequency */
    interval: number;
    /** Set of specific days within a week for the event to recur on */
    by_weekday: discord.RecurrenceRuleWeekdays[] | null;
    /** List of specific days within a specific week to recur on */
    by_n_weekday: discord.ByNWeekdayResponse[] | null;
    /** Set of specific months to recur on */
    by_month: discord.RecurrenceRuleMonths[] | null;
    /** Set of specific dates within a month to recur on */
    by_month_day: number[] | null;
    /** Set of days within a year to recur on (1-364) */
    by_year_day?: number[] | null;
    /** Total number of times the event is allowed to recur */
    count?: number | null;
  }
  type RecurrenceRuleWeekdays = 0 | 1 | 2 | 3 | 4 | 5 | 6;
  interface ResolvedObjectsResponse {
    users?: Record<string, discord.UserResponse> | null;
    members?: Record<string, discord.BasicGuildMemberResponse> | null;
    channels?: Record<string, discord.GuildChannelResponse | discord.PrivateChannelResponse | discord.PrivateGroupChannelResponse | discord.ThreadResponse> | null;
    roles?: Record<string, discord.GuildRoleResponse> | null;
  }
  interface ResourceChannelResponse {
    channel_id: string;
    title: string;
    emoji?: discord.SettingsEmojiResponse;
    icon?: string;
    description: string;
  }
  interface RichEmbed {
    type?: string | null;
    url?: string | null;
    title?: string | null;
    color?: number | null;
    timestamp?: string | null;
    description?: string | null;
    author?: discord.RichEmbedAuthor | null;
    image?: discord.RichEmbedImage | null;
    thumbnail?: discord.RichEmbedThumbnail | null;
    footer?: discord.RichEmbedFooter | null;
    fields?: discord.RichEmbedField[] | null;
    provider?: discord.RichEmbedProvider | null;
    video?: discord.RichEmbedVideo | null;
  }
  interface RichEmbedAuthor {
    name?: string | null;
    url?: string | null;
    icon_url?: string | null;
  }
  interface RichEmbedField {
    name: string;
    value: string;
    inline?: boolean | null;
  }
  interface RichEmbedFooter {
    text?: string | null;
    icon_url?: string | null;
  }
  interface RichEmbedImage {
    url?: string | null;
    width?: number | null;
    height?: number | null;
    placeholder?: string | null;
    placeholder_version?: number | null;
    is_animated?: boolean | null;
    description?: string | null;
  }
  interface RichEmbedProvider {
    name?: string | null;
    url?: string | null;
  }
  interface RichEmbedThumbnail {
    url?: string | null;
    width?: number | null;
    height?: number | null;
    placeholder?: string | null;
    placeholder_version?: number | null;
    is_animated?: boolean | null;
    description?: string | null;
  }
  interface RichEmbedVideo {
    url?: string | null;
    width?: number | null;
    height?: number | null;
    placeholder?: string | null;
    placeholder_version?: number | null;
    is_animated?: boolean | null;
    description?: string | null;
  }
  interface RoleColors {
    primary_color?: number | null;
    secondary_color?: number | null;
    tertiary_color?: number | null;
  }
  interface RoleSelectComponentForMessageRequest {
    type: 6;
    id?: number | null;
    custom_id: string;
    placeholder?: string | null;
    min_values?: number | null;
    max_values?: number | null;
    disabled?: boolean | null;
    required?: boolean | null;
    default_values?: discord.RoleSelectDefaultValue[] | null;
  }
  interface RoleSelectComponentResponse {
    type: 6;
    id: number;
    custom_id: string;
    placeholder?: string;
    min_values: number;
    max_values: number;
    disabled?: boolean;
    default_values?: discord.RoleSelectDefaultValueResponse[];
  }
  interface RoleSelectDefaultValue {
    type: "role";
    id: string;
  }
  interface RoleSelectDefaultValueResponse {
    type: "role";
    id: string;
  }
  interface SDKMessageRequest {
    content?: string | null;
    embeds?: discord.RichEmbed[] | null;
    allowed_mentions?: discord.MessageAllowedMentionsRequest | null;
    sticker_ids?: string[] | null;
    components?: (discord.ActionRowComponentForMessageRequest | discord.ContainerComponentForMessageRequest | discord.FileComponentForMessageRequest | discord.MediaGalleryComponentForMessageRequest | discord.SectionComponentForMessageRequest | discord.SeparatorComponentForMessageRequest | discord.TextDisplayComponentForMessageRequest)[] | null;
    flags?: number | null;
    attachments?: discord.MessageAttachmentRequest[] | null;
    poll?: discord.PollCreateRequest | null;
    shared_client_theme?: discord.CustomClientThemeShareRequest | null;
    message_reference?: discord.MessageReferenceRequest | null;
    nonce?: number | string | null;
    enforce_nonce?: boolean | null;
    tts?: boolean | null;
  }
  interface ScheduledEventResponse {
    /** ID of the scheduled event */
    id: string;
    /** ID of the guild the scheduled event belongs to */
    guild_id: string;
    /** Name of the scheduled event */
    name: string;
    /** Description of the scheduled event */
    description: string | null;
    /** Channel ID in which the scheduled event will be hosted, or null if entity type is EXTERNAL */
    channel_id: string | null;
    /** ID of the user that created the scheduled event */
    creator_id: string | null;
    /** User that created the scheduled event */
    creator?: discord.UserResponse;
    /** Cover image hash of the scheduled event */
    image: string | null;
    /** When the scheduled event will start */
    scheduled_start_time: string;
    /** When the scheduled event will end, or null if no end time */
    scheduled_end_time: string | null;
    /** Status of the scheduled event */
    status: discord.GuildScheduledEventStatuses;
    /** Type of hosting entity associated with the scheduled event */
    entity_type: discord.GuildScheduledEventEntityTypes;
    /** ID of the hosting entity associated with the scheduled event */
    entity_id: string | null;
    /** Recurrence rule for the scheduled event, or null if not recurring */
    recurrence_rule: discord.RecurrenceRuleResponse | null;
    /** Number of users subscribed to the scheduled event */
    user_count?: number;
    /** Privacy level of the scheduled event */
    privacy_level: discord.GuildScheduledEventPrivacyLevels;
    user_rsvp?: discord.ScheduledEventUserResponse | null;
    guild_scheduled_event_exceptions: discord.GuildScheduledEventExceptionResponse[];
  }
  interface ScheduledEventUserCountResponse {
    /** The number of users subscribed to the scheduled event */
    guild_scheduled_event_count: number;
    /** Map of exception IDs to user counts for each exception */
    guild_scheduled_event_exception_counts: Record<string, number>;
  }
  interface ScheduledEventUserResponse {
    /** ID of the scheduled event */
    guild_scheduled_event_id: string;
    /** ID of the scheduled event exception */
    guild_scheduled_event_exception_id?: string | null;
    /** ID of the user */
    user_id: string;
    /** User object for the RSVP user */
    user?: discord.UserResponse;
    /** Guild member object for the RSVP user */
    member?: discord.GuildMemberResponse;
    /** User's RSVP status for the event */
    response: discord.GuildScheduledEventUserResponses;
  }
  interface SearchIndexNotReadyResponse {
    message: string;
    code: number;
    documents_indexed: number;
    retry_after: number;
  }
  interface SearchMessageResponse {
    type: discord.MessageType;
    content: string;
    mentions: discord.UserResponse[];
    mention_roles: string[];
    attachments: discord.MessageAttachmentResponse[];
    embeds: discord.MessageEmbedResponse[];
    timestamp: string;
    edited_timestamp: string | null;
    flags: number;
    components: (discord.ActionRowComponentResponse | discord.ContainerComponentResponse | discord.FileComponentResponse | discord.MediaGalleryComponentResponse | discord.SectionComponentResponse | discord.SeparatorComponentResponse | discord.TextDisplayComponentResponse)[];
    stickers?: (discord.GuildStickerResponse | discord.StandardStickerResponse)[];
    sticker_items?: discord.MessageStickerItemResponse[];
    id: string;
    channel_id: string;
    author: discord.UserResponse;
    pinned: boolean;
    mention_everyone: boolean;
    tts: boolean;
    call?: discord.MessageCallResponse;
    activity?: discord.MessageActivityResponse;
    application?: discord.BasicApplicationResponseWithBot;
    application_id?: string;
    interaction?: discord.MessageInteractionResponse;
    nonce?: number | string;
    webhook_id?: string;
    message_reference?: discord.MessageReferenceResponse;
    thread?: discord.ThreadResponse;
    mention_channels?: discord.MessageMentionChannelResponse[];
    role_subscription_data?: discord.MessageRoleSubscriptionDataResponse;
    purchase_notification?: discord.PurchaseNotificationResponse;
    position?: number;
    resolved?: discord.ResolvedObjectsResponse;
    poll?: discord.PollResponse;
    shared_client_theme?: discord.CustomClientThemeResponse;
    interaction_metadata?: discord.ApplicationCommandInteractionMetadataResponse | discord.MessageComponentInteractionMetadataResponse | discord.ModalSubmitInteractionMetadataResponse;
    message_snapshots?: discord.MessageSnapshotResponse[];
    lobby_member?: discord.MessageLobbyMemberResponse;
    reactions?: discord.MessageReactionResponse[];
    referenced_message?: discord.BasicMessageResponse | null;
    hit: boolean;
    restriction_count?: number;
  }
  interface SectionComponentForMessageRequest {
    type: 9;
    id?: number | null;
    components: discord.TextDisplayComponentForMessageRequest[];
    accessory: discord.ButtonComponentForMessageRequest | discord.ThumbnailComponentForMessageRequest;
  }
  interface SectionComponentResponse {
    type: 9;
    id: number;
    components: discord.TextDisplayComponentResponse[];
    accessory: discord.ButtonComponentResponse | discord.ThumbnailComponentResponse;
  }
  interface SeparatorComponentForMessageRequest {
    type: 14;
    id?: number | null;
    spacing?: discord.MessageComponentSeparatorSpacingSize | null;
    divider?: boolean | null;
  }
  interface SeparatorComponentResponse {
    type: 14;
    id: number;
    spacing: discord.MessageComponentSeparatorSpacingSize;
    divider: boolean;
  }
  interface SettingsEmojiResponse {
    id: string | null;
    name: string | null;
    animated: boolean;
  }
  interface SoundboardCreateRequest {
    name: string;
    volume?: number | null;
    emoji_id?: string | null;
    emoji_name?: string | null;
    sound: string;
  }
  interface SoundboardPatchRequestPartial {
    name?: string;
    volume?: number | null;
    emoji_id?: string | null;
    emoji_name?: string | null;
  }
  interface SoundboardSoundResponse {
    name: string;
    sound_id: string;
    volume: number;
    emoji_id: string | null;
    emoji_name: string | null;
    guild_id?: string;
    available: boolean;
    user?: discord.UserResponse;
  }
  interface SoundboardSoundSendRequest {
    sound_id: string;
    source_guild_id?: string | null;
  }
  interface StageInstanceResponse {
    guild_id: string;
    channel_id: string;
    topic: string;
    privacy_level: discord.StageInstancesPrivacyLevels;
    id: string;
    discoverable_disabled: boolean;
    guild_scheduled_event_id: string | null;
  }
  type StageInstancesPrivacyLevels = 1 | 2;
  interface StageScheduledEventCreateRequest {
    name: string;
    description?: string | null;
    image?: string | null;
    scheduled_start_time: string;
    scheduled_end_time?: string | null;
    privacy_level: discord.GuildScheduledEventPrivacyLevels;
    entity_type: 1;
    channel_id?: string | null;
    /** Recurrence rule for the scheduled event */
    recurrence_rule?: discord.RecurrenceRule | null;
    entity_metadata?: discord.EntityMetadataStageInstance | null;
  }
  interface StageScheduledEventPatchRequestPartial {
    status?: discord.GuildScheduledEventStatuses | null;
    name?: string;
    description?: string | null;
    image?: string | null;
    scheduled_start_time?: string;
    scheduled_end_time?: string | null;
    entity_type?: 1 | null;
    privacy_level?: discord.GuildScheduledEventPrivacyLevels;
    channel_id?: string | null;
    /** Recurrence rule for the scheduled event */
    recurrence_rule?: discord.RecurrenceRule | null;
    entity_metadata?: discord.EntityMetadataStageInstance | null;
  }
  interface StageScheduledEventResponse {
    /** ID of the scheduled event */
    id: string;
    /** ID of the guild the scheduled event belongs to */
    guild_id: string;
    /** Name of the scheduled event */
    name: string;
    /** Description of the scheduled event */
    description: string | null;
    /** Channel ID in which the scheduled event will be hosted, or null if entity type is EXTERNAL */
    channel_id: string | null;
    /** ID of the user that created the scheduled event */
    creator_id: string | null;
    /** User that created the scheduled event */
    creator?: discord.UserResponse;
    /** Cover image hash of the scheduled event */
    image: string | null;
    /** When the scheduled event will start */
    scheduled_start_time: string;
    /** When the scheduled event will end, or null if no end time */
    scheduled_end_time: string | null;
    /** Status of the scheduled event */
    status: discord.GuildScheduledEventStatuses;
    /** Type of hosting entity associated with the scheduled event */
    entity_type: 1;
    /** ID of the hosting entity associated with the scheduled event */
    entity_id: string | null;
    /** Recurrence rule for the scheduled event, or null if not recurring */
    recurrence_rule: discord.RecurrenceRuleResponse | null;
    /** Number of users subscribed to the scheduled event */
    user_count?: number;
    /** Privacy level of the scheduled event */
    privacy_level: discord.GuildScheduledEventPrivacyLevels;
    user_rsvp?: discord.ScheduledEventUserResponse | null;
    guild_scheduled_event_exceptions: discord.GuildScheduledEventExceptionResponse[];
    entity_metadata: discord.EntityMetadataStageInstanceResponse | null;
  }
  interface StandardStickerResponse {
    id: string;
    name: string;
    tags: string;
    type: 1;
    format_type: discord.StickerFormatTypes | null;
    description: string | null;
    pack_id: string;
    sort_value: number;
  }
  type StickerFormatTypes = 1 | 2 | 3 | 4;
  interface StickerPackCollectionResponse {
    sticker_packs: discord.StickerPackResponse[];
  }
  interface StickerPackResponse {
    id: string;
    sku_id: string;
    name: string;
    description: string | null;
    stickers: discord.StandardStickerResponse[];
    cover_sticker_id?: string;
    banner_asset_id?: string;
  }
  interface StringSelectComponentForMessageRequest {
    type: 3;
    id?: number | null;
    custom_id: string;
    placeholder?: string | null;
    min_values?: number | null;
    max_values?: number | null;
    disabled?: boolean | null;
    required?: boolean | null;
    options: discord.StringSelectOptionForRequest[];
  }
  interface StringSelectComponentResponse {
    type: 3;
    id: number;
    custom_id: string;
    placeholder?: string;
    min_values: number;
    max_values: number;
    disabled?: boolean;
    options: discord.StringSelectOptionResponse[];
  }
  interface StringSelectOptionForRequest {
    label: string;
    value: string;
    description?: string | null;
    default?: boolean | null;
    emoji?: discord.ComponentEmojiForRequest | null;
  }
  interface StringSelectOptionResponse {
    label: string;
    value: string;
    description?: string;
    emoji?: discord.ComponentEmojiResponse;
    default?: boolean;
  }
  interface SubscriptionResponse {
    /** ID of the subscription */
    id: string;
    /** ID of the user who is subscribed */
    user_id: string;
    /** List of SKUs subscribed to */
    sku_ids: string[];
    /** List of SKUs that this user will be subscribed to at renewal */
    renewal_sku_ids: string[] | null;
    /** List of entitlements granted for this subscription */
    entitlement_ids: string[];
    /** Start of the current subscription period */
    current_period_start: string;
    /** End of the current subscription period */
    current_period_end: string;
    /** Current status of the subscription */
    status: discord.SubscriptionResponseStatusType;
    /** When the subscription was canceled */
    canceled_at: string | null;
    /** ISO3166-1 alpha-2 country code of the payment source used to purchase the subscription */
    country?: string | null;
  }
  type SubscriptionResponseStatusType = 0 | 1 | 2;
  interface TargetUsersJobStatusResponse {
    /** The status of the job processing the target users. */
    status: discord.TargetUsersJobStatusTypes;
    /** The total number of users in the provided list. */
    total_users: number;
    /** The number of users processed so far. */
    processed_users: number;
    /** The timestamp when the job was created. */
    created_at: string | null;
    /** The timestamp when the job was successfully completed. */
    completed_at: string | null;
    /** The error message if the job failed. */
    error_message: string | null;
  }
  type TargetUsersJobStatusTypes = 0 | 1 | 2 | 3;
  interface TeamMemberResponse {
    user: discord.UserResponse;
    team_id: string;
    membership_state: discord.TeamMembershipStates;
    role: discord.TeamMemberRoles;
    permissions: string[];
  }
  type TeamMemberRoles = "admin" | "developer" | "read_only";
  type TeamMembershipStates = 1 | 2;
  interface TeamResponse {
    id: string;
    icon: string | null;
    name: string;
    owner_user_id: string;
    members: discord.TeamMemberResponse[];
  }
  interface TermsFormFieldResponse {
    /** Type of form field */
    field_type: "TERMS";
    /** Label shown above field */
    label?: string;
    /** Optional helper text shown below label */
    description?: string;
    /** Whether applicant must fill in field */
    required?: boolean;
    /** Terms applicant must acknowledge */
    values: string[];
    /** Whether applicant accepted terms */
    response?: boolean;
  }
  interface TextDisplayComponentForMessageRequest {
    type: 10;
    id?: number | null;
    content: string;
  }
  interface TextDisplayComponentResponse {
    type: 10;
    id: number;
    content: string;
  }
  interface TextInputComponentResponse {
    type: 4;
    id: number;
    custom_id: string;
    style: discord.TextInputStyleTypes;
    label: string | null;
    value?: string;
    placeholder?: string;
    required?: boolean;
    min_length: number | null;
    max_length: number | null;
  }
  interface TextInputFormFieldResponse {
    /** Type of form field */
    field_type: "TEXT_INPUT";
    /** Label shown above field */
    label?: string;
    /** Optional helper text shown below label */
    description?: string;
    /** Whether applicant must fill in field */
    required?: boolean;
    /** Placeholder text shown in empty input */
    placeholder?: string;
    /** Applicant's text response */
    response?: string;
  }
  type TextInputStyleTypes = 1 | 2;
  type ThreadAutoArchiveDuration = 60 | 1440 | 4320 | 10080;
  interface ThreadMemberResponse {
    id: string;
    user_id: string;
    join_timestamp: string;
    flags: number;
    member?: discord.GuildMemberResponse;
  }
  interface ThreadMetadataResponse {
    archived: boolean;
    archive_timestamp: string | null;
    auto_archive_duration: discord.ThreadAutoArchiveDuration;
    locked: boolean;
    create_timestamp?: string;
    invitable?: boolean;
  }
  interface ThreadResponse {
    id: string;
    type: 10 | 11 | 12;
    last_message_id?: string | null;
    flags: number;
    last_pin_timestamp?: string | null;
    guild_id: string;
    name: string;
    parent_id?: string | null;
    rate_limit_per_user?: number;
    bitrate?: number;
    user_limit?: number;
    rtc_region?: string | null;
    video_quality_mode?: discord.VideoQualityModes;
    permissions?: string;
    owner_id: string;
    thread_metadata: discord.ThreadMetadataResponse;
    message_count: number;
    member_count: number;
    total_message_sent: number;
    applied_tags?: string[];
    member?: discord.ThreadMemberResponse;
  }
  interface ThreadSearchResponse {
    threads: discord.ThreadResponse[];
    members: discord.ThreadMemberResponse[];
    has_more: boolean;
    first_messages?: discord.MessageResponse[];
    total_results: number;
  }
  type ThreadSearchTagSetting = "match_all" | "match_some";
  type ThreadSortOrder = 0 | 1;
  interface ThreadsResponse {
    threads: discord.ThreadResponse[];
    members: discord.ThreadMemberResponse[];
    has_more: boolean;
    first_messages?: discord.MessageResponse[];
  }
  interface ThumbnailComponentForMessageRequest {
    type: 11;
    id?: number | null;
    description?: string | null;
    spoiler?: boolean | null;
    media: discord.UnfurledMediaRequest;
  }
  interface ThumbnailComponentResponse {
    type: 11;
    id: number;
    media: discord.UnfurledMediaResponse;
    description: string | null;
    spoiler: boolean;
  }
  interface TypingIndicatorResponse {
  
  }
  interface UnbanUserFromGuildRequest {
  
  }
  interface UnfurledMediaRequest {
    url: string;
  }
  interface UnfurledMediaRequestWithAttachmentReferenceRequired {
    url: string;
  }
  interface UnfurledMediaResponse {
    id: string;
    url: string;
    proxy_url: string;
    width?: number | null;
    height?: number | null;
    content_type?: string | null;
    attachment_id?: string;
  }
  interface UpdateDMRequestPartial {
    name?: string | null;
  }
  interface UpdateDefaultReactionEmojiRequest {
    emoji_id?: string | null;
    emoji_name?: string | null;
  }
  interface UpdateGroupDMRequestPartial {
    name?: string | null;
    icon?: string | null;
  }
  interface UpdateGuildChannelRequestPartial {
    type?: 0 | 2 | 4 | 5 | 13 | 14 | 15 | null;
    name?: string;
    position?: number | null;
    topic?: string | null;
    bitrate?: number | null;
    user_limit?: number | null;
    nsfw?: boolean | null;
    rate_limit_per_user?: number | null;
    parent_id?: string | null;
    permission_overwrites?: discord.ChannelPermissionOverwriteRequest[] | null;
    rtc_region?: string | null;
    video_quality_mode?: discord.VideoQualityModes | null;
    default_auto_archive_duration?: discord.ThreadAutoArchiveDuration | null;
    default_reaction_emoji?: discord.UpdateDefaultReactionEmojiRequest | null;
    default_thread_rate_limit_per_user?: number | null;
    default_sort_order?: discord.ThreadSortOrder | null;
    default_forum_layout?: discord.ForumLayout | null;
    default_tag_setting?: discord.ThreadSearchTagSetting | null;
    flags?: number | null;
    available_tags?: discord.UpdateThreadTagRequest[] | null;
  }
  interface UpdateGuildOnboardingRequest {
    prompts?: discord.UpdateOnboardingPromptRequest[] | null;
    enabled?: boolean | null;
    default_channel_ids?: string[] | null;
    mode?: discord.GuildOnboardingMode | null;
  }
  interface UpdateOnboardingPromptRequest {
    title: string;
    options: discord.OnboardingPromptOptionRequest[];
    single_select?: boolean | null;
    required?: boolean | null;
    in_onboarding?: boolean | null;
    type?: discord.OnboardingPromptType | null;
    id: string;
  }
  interface UpdateRolePositionsRequest {
    id?: string | null;
    position?: number | null;
  }
  interface UpdateRoleRequestPartial {
    name?: string | null;
    permissions?: number | null;
    color?: number | null;
    colors?: discord.RoleColors | null;
    hoist?: boolean | null;
    mentionable?: boolean | null;
    icon?: string | null;
    unicode_emoji?: string | null;
  }
  interface UpdateSelfVoiceStateRequestPartial {
    request_to_speak_timestamp?: string | null;
    suppress?: boolean | null;
    channel_id?: string | null;
  }
  interface UpdateThreadRequestPartial {
    name?: string | null;
    archived?: boolean | null;
    locked?: boolean | null;
    invitable?: boolean | null;
    auto_archive_duration?: discord.ThreadAutoArchiveDuration | null;
    rate_limit_per_user?: number | null;
    flags?: number | null;
    applied_tags?: string[] | null;
    bitrate?: number | null;
    user_limit?: number | null;
    rtc_region?: string | null;
    video_quality_mode?: discord.VideoQualityModes | null;
  }
  interface UpdateThreadTagRequest {
    name: string;
    emoji_id?: string | null;
    emoji_name?: string | null;
    moderated?: boolean | null;
    id?: string | null;
  }
  interface UpdateVoiceStateRequestPartial {
    suppress?: boolean | null;
    channel_id?: string | null;
  }
  interface UserAvatarDecorationResponse {
    /** the avatar decoration hash */
    asset: string;
    /** id of the avatar decoration's SKU */
    sku_id: string | null;
  }
  interface UserCollectiblesResponse {
    /** Object mapping of nameplate data */
    nameplate: discord.UserNameplateResponse | null;
  }
  interface UserCommunicationDisabledAction {
    type: 3;
    metadata: discord.UserCommunicationDisabledActionMetadata;
  }
  interface UserCommunicationDisabledActionMetadata {
    duration_seconds?: number | null;
  }
  interface UserCommunicationDisabledActionMetadataResponse {
    duration_seconds: number;
  }
  interface UserCommunicationDisabledActionResponse {
    type: 3;
    metadata: discord.UserCommunicationDisabledActionMetadataResponse;
  }
  interface UserGuildOnboardingResponse {
    guild_id: string;
    prompts: discord.OnboardingPromptResponse[];
    default_channel_ids: string[];
    enabled: boolean;
    mode: discord.GuildOnboardingMode;
  }
  interface UserNameplateResponse {
    /** ID of the nameplate SKU */
    sku_id: string | null;
    /** Path to the nameplate asset */
    asset: string;
    /** The label of this nameplate. Currently unused */
    label: string;
    /** Background color of the nameplate */
    palette: discord.NameplatePalette;
  }
  type UserNotificationSettings = 0 | 1;
  interface UserPIIResponse {
    /** the user's id */
    id: string;
    /** the user's username, not unique across the platform */
    username: string;
    /** the user's avatar hash */
    avatar: string | null;
    /** the user's Discord-tag */
    discriminator: string;
    /** the public flags on a user's account */
    public_flags: number;
    /** the flags on a user's account */
    flags: number;
    /** whether the user belongs to an OAuth2 application */
    bot?: boolean;
    /** whether the user is an Official Discord System user (part of the urgent message system) */
    system?: boolean;
    /** the user's banner hash */
    banner?: string | null;
    /** the user's banner color encoded as an integer representation of hexadecimal color code */
    accent_color?: number | null;
    /** the user's display name, if it is set */
    global_name: string | null;
    /** data for the user's avatar decoration */
    avatar_decoration_data?: discord.UserAvatarDecorationResponse | null;
    /** data for the user's collectibles */
    collectibles?: discord.UserCollectiblesResponse | null;
    primary_guild?: discord.UserPrimaryGuildResponse | null;
    mfa_enabled: boolean;
    locale: discord.AvailableLocalesEnum;
    premium_type?: discord.PremiumTypes;
    email?: string | null;
    verified?: boolean;
  }
  interface UserPrimaryGuildResponse {
    /** the id of the user's primary guild */
    identity_guild_id: string | null;
    /** whether the user is displaying the primary guild's server tag */
    identity_enabled: boolean | null;
    /** the text of the user's server tag, limited to 4 characters */
    tag: string | null;
    /** the server tag badge hash */
    badge: string | null;
  }
  interface UserProfileMetadata {
    keyword_filter?: string[] | null;
    regex_patterns?: string[] | null;
    allow_list?: string[] | null;
  }
  interface UserProfileMetadataResponse {
    keyword_filter: string[];
    regex_patterns: string[];
    allow_list: string[];
  }
  interface UserProfileRuleResponse {
    id: string;
    guild_id: string;
    creator_id: string;
    name: string;
    event_type: discord.AutomodEventType;
    actions: (discord.BlockMessageActionResponse | discord.FlagToChannelActionResponse | discord.QuarantineUserActionResponse | discord.UserCommunicationDisabledActionResponse)[];
    trigger_type: 6;
    enabled: boolean;
    exempt_roles: string[];
    exempt_channels: string[];
    trigger_metadata: discord.UserProfileMetadataResponse;
  }
  interface UserProfileUpsertRequest {
    name: string;
    event_type: discord.AutomodEventType;
    actions?: (discord.BlockMessageAction | discord.FlagToChannelAction | discord.QuarantineUserAction | discord.UserCommunicationDisabledAction)[] | null;
    enabled?: boolean | null;
    exempt_roles?: string[] | null;
    exempt_channels?: string[] | null;
    trigger_type: 6;
    trigger_metadata: discord.UserProfileMetadata;
  }
  interface UserProfileUpsertRequestPartial {
    name?: string;
    event_type?: discord.AutomodEventType;
    actions?: (discord.BlockMessageAction | discord.FlagToChannelAction | discord.QuarantineUserAction | discord.UserCommunicationDisabledAction)[] | null;
    enabled?: boolean | null;
    exempt_roles?: string[] | null;
    exempt_channels?: string[] | null;
    trigger_type?: 6;
    trigger_metadata?: discord.UserProfileMetadata;
  }
  interface UserResponse {
    /** the user's id */
    id: string;
    /** the user's username, not unique across the platform */
    username: string;
    /** the user's avatar hash */
    avatar: string | null;
    /** the user's Discord-tag */
    discriminator: string;
    /** the public flags on a user's account */
    public_flags: number;
    /** the flags on a user's account */
    flags: number;
    /** whether the user belongs to an OAuth2 application */
    bot?: boolean;
    /** whether the user is an Official Discord System user (part of the urgent message system) */
    system?: boolean;
    /** the user's banner hash */
    banner?: string | null;
    /** the user's banner color encoded as an integer representation of hexadecimal color code */
    accent_color?: number | null;
    /** the user's display name, if it is set */
    global_name: string | null;
    /** data for the user's avatar decoration */
    avatar_decoration_data?: discord.UserAvatarDecorationResponse | null;
    /** data for the user's collectibles */
    collectibles?: discord.UserCollectiblesResponse | null;
    /** the user's primary guild */
    primary_guild: discord.UserPrimaryGuildResponse | null;
  }
  interface UserSelectComponentForMessageRequest {
    type: 5;
    id?: number | null;
    custom_id: string;
    placeholder?: string | null;
    min_values?: number | null;
    max_values?: number | null;
    disabled?: boolean | null;
    required?: boolean | null;
    default_values?: discord.UserSelectDefaultValue[] | null;
  }
  interface UserSelectComponentResponse {
    type: 5;
    id: number;
    custom_id: string;
    placeholder?: string;
    min_values: number;
    max_values: number;
    disabled?: boolean;
    default_values?: discord.UserSelectDefaultValueResponse[];
  }
  interface UserSelectDefaultValue {
    type: "user";
    id: string;
  }
  interface UserSelectDefaultValueResponse {
    type: "user";
    id: string;
  }
  interface VanityURLErrorResponse {
    message: string;
    code: number;
  }
  interface VanityURLResponse {
    code: string | null;
    uses: number;
    error?: discord.VanityURLErrorResponse | null;
  }
  type VerificationLevels = 0 | 1 | 2 | 3 | 4;
  type VideoQualityModes = 1 | 2;
  interface VoiceRegionResponse {
    id: string;
    name: string;
    custom: boolean;
    deprecated: boolean;
    optimal: boolean;
  }
  interface VoiceScheduledEventCreateRequest {
    name: string;
    description?: string | null;
    image?: string | null;
    scheduled_start_time: string;
    scheduled_end_time?: string | null;
    privacy_level: discord.GuildScheduledEventPrivacyLevels;
    entity_type: 2;
    channel_id?: string | null;
    /** Recurrence rule for the scheduled event */
    recurrence_rule?: discord.RecurrenceRule | null;
    entity_metadata?: discord.EntityMetadataVoice | null;
  }
  interface VoiceScheduledEventPatchRequestPartial {
    status?: discord.GuildScheduledEventStatuses | null;
    name?: string;
    description?: string | null;
    image?: string | null;
    scheduled_start_time?: string;
    scheduled_end_time?: string | null;
    entity_type?: 2 | null;
    privacy_level?: discord.GuildScheduledEventPrivacyLevels;
    channel_id?: string | null;
    /** Recurrence rule for the scheduled event */
    recurrence_rule?: discord.RecurrenceRule | null;
    entity_metadata?: discord.EntityMetadataVoice | null;
  }
  interface VoiceScheduledEventResponse {
    /** ID of the scheduled event */
    id: string;
    /** ID of the guild the scheduled event belongs to */
    guild_id: string;
    /** Name of the scheduled event */
    name: string;
    /** Description of the scheduled event */
    description: string | null;
    /** Channel ID in which the scheduled event will be hosted, or null if entity type is EXTERNAL */
    channel_id: string | null;
    /** ID of the user that created the scheduled event */
    creator_id: string | null;
    /** User that created the scheduled event */
    creator?: discord.UserResponse;
    /** Cover image hash of the scheduled event */
    image: string | null;
    /** When the scheduled event will start */
    scheduled_start_time: string;
    /** When the scheduled event will end, or null if no end time */
    scheduled_end_time: string | null;
    /** Status of the scheduled event */
    status: discord.GuildScheduledEventStatuses;
    /** Type of hosting entity associated with the scheduled event */
    entity_type: 2;
    /** ID of the hosting entity associated with the scheduled event */
    entity_id: string | null;
    /** Recurrence rule for the scheduled event, or null if not recurring */
    recurrence_rule: discord.RecurrenceRuleResponse | null;
    /** Number of users subscribed to the scheduled event */
    user_count?: number;
    /** Privacy level of the scheduled event */
    privacy_level: discord.GuildScheduledEventPrivacyLevels;
    user_rsvp?: discord.ScheduledEventUserResponse | null;
    guild_scheduled_event_exceptions: discord.GuildScheduledEventExceptionResponse[];
    entity_metadata: discord.EntityMetadataVoiceResponse | null;
  }
  interface VoiceStateResponse {
    channel_id: string | null;
    deaf: boolean;
    guild_id: string | null;
    member?: discord.GuildMemberResponse;
    mute: boolean;
    request_to_speak_timestamp: string | null;
    suppress: boolean;
    self_stream: boolean | null;
    self_deaf: boolean;
    self_mute: boolean;
    self_video: boolean;
    session_id: string;
    user_id: string;
  }
  interface WebhookSourceChannelResponse {
    id: string;
    name: string;
  }
  interface WebhookSourceGuildResponse {
    id: string;
    icon: string | null;
    name: string;
  }
  interface WelcomeMessageResponse {
    author_ids: string[];
    message: string;
  }
  interface WelcomeScreenPatchRequestPartial {
    description?: string | null;
    welcome_channels?: discord.GuildWelcomeChannel[] | null;
    enabled?: boolean | null;
  }
  interface WidgetActivity {
    name: string;
  }
  interface WidgetChannel {
    id: string;
    name: string;
    position: number;
  }
  interface WidgetMember {
    id: string;
    username: string;
    discriminator: discord.WidgetUserDiscriminator;
    avatar: null;
    status: string;
    avatar_url: string;
    activity?: discord.WidgetActivity;
    deaf?: boolean;
    mute?: boolean;
    self_deaf?: boolean;
    self_mute?: boolean;
    suppress?: boolean;
    channel_id?: string;
  }
  interface WidgetResponse {
    id: string;
    name: string;
    instant_invite: string | null;
    channels: discord.WidgetChannel[];
    members: discord.WidgetMember[];
    presence_count: number;
  }
  interface WidgetSettingsResponse {
    enabled: boolean;
    channel_id: string | null;
  }
  type WidgetUserDiscriminator = "0000";
}
