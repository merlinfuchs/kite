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
    /** The JSON body of the request. */ body?: { action?: any; rejection_reason?: any };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Add group dm user, PUT /channels/{channel_id}/recipients/{user_id}. Costs 1 credit. */
  addGroupDmUser(arg: {
    channel_id: string;
    user_id: string;
    /** The JSON body of the request. */ body?: { access_token?: any; nick?: any };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Add guild member, PUT /guilds/{guild_id}/members/{user_id}. Costs 1 credit. */
  addGuildMember(arg: {
    guild_id: string;
    user_id: string;
    /** The JSON body of the request. */ body: { nick?: any; roles?: any; mute?: boolean; deaf?: boolean; access_token: any; flags?: number };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Add guild member role, PUT /guilds/{guild_id}/members/{user_id}/roles/{role_id}. Costs 1 credit. */
  addGuildMemberRole(arg: {
    guild_id: string;
    user_id: string;
    role_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Add invite target user, PUT /invites/{code}/target-users/{user_id}. Costs 1 credit. */
  addInviteTargetUser(arg: {
    code: string;
    user_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Add lobby member, PUT /lobbies/{lobby_id}/members/{user_id}. Costs 1 credit. */
  addLobbyMember(arg: {
    lobby_id: string;
    user_id: string;
    /** The JSON body of the request. */ body?: { metadata?: any; flags?: number; additional_name?: any };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Add my message reaction, PUT /channels/{channel_id}/messages/{message_id}/reactions/{emoji_name}/@me. Costs 1 credit. */
  addMyMessageReaction(arg: {
    channel_id: string;
    message_id: string;
    emoji_name: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Add thread member, PUT /channels/{channel_id}/thread-members/{user_id}. Costs 1 credit. */
  addThreadMember(arg: {
    channel_id: string;
    user_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Applications get activity instance, GET /applications/{application_id}/activity-instances/{instance_id}. Costs nothing. */
  applicationsGetActivityInstance(arg: {
    application_id: string;
    instance_id: string;
  }): Promise<any>;
  /** Ban user from guild, PUT /guilds/{guild_id}/bans/{user_id}. Costs 1 credit. */
  banUserFromGuild(arg: {
    guild_id: string;
    user_id: string;
    /** The JSON body of the request. */ body?: { delete_message_seconds?: number; delete_message_days?: number };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Bot partner sdk token, POST /partner-sdk/token/bot. Costs 1 credit. */
  botPartnerSdkToken(arg: {
    /** The JSON body of the request. */ body: { provisional_user_id?: string; external_user_id: any; preferred_global_name?: any };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Bot partner sdk unmerge provisional account, POST /partner-sdk/provisional-accounts/unmerge/bot. Costs 1 credit. */
  botPartnerSdkUnmergeProvisionalAccount(arg: {
    /** The JSON body of the request. */ body: { external_user_id: any };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Bulk add invite target users, POST /invites/{code}/target-users/bulk-add. Costs 1 credit. */
  bulkAddInviteTargetUsers(arg: {
    code: string;
    /** The JSON body of the request. */ body: { user_ids: any };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Bulk delete messages, POST /channels/{channel_id}/messages/bulk-delete. Costs 1 credit. */
  bulkDeleteMessages(arg: {
    channel_id: string;
    /** The JSON body of the request. */ body: { messages: any };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Bulk remove invite target users, POST /invites/{code}/target-users/bulk-delete. Costs 1 credit. */
  bulkRemoveInviteTargetUsers(arg: {
    code: string;
    /** The JSON body of the request. */ body: { user_ids: any };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Bulk update guild channels, PATCH /guilds/{guild_id}/channels. Costs 1 credit. */
  bulkUpdateGuildChannels(arg: {
    guild_id: string;
    /** The JSON body of the request. */ body?: any;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Bulk update guild roles, PATCH /guilds/{guild_id}/roles. Costs 1 credit. */
  bulkUpdateGuildRoles(arg: {
    guild_id: string;
    /** The JSON body of the request. */ body?: any;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Bulk update lobby members, POST /lobbies/{lobby_id}/members/bulk. Costs 1 credit. */
  bulkUpdateLobbyMembers(arg: {
    lobby_id: string;
    /** The JSON body of the request. */ body?: any;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Count guild scheduled event users, GET /guilds/{guild_id}/scheduled-events/{guild_scheduled_event_id}/users/counts. Costs nothing. */
  countGuildScheduledEventUsers(arg: {
    guild_id: string;
    guild_scheduled_event_id: string;
    guild_scheduled_event_exception_ids?: (string | number)[];
  }): Promise<any>;
  /** Create auto moderation rule, POST /guilds/{guild_id}/auto-moderation/rules. Costs 1 credit. */
  createAutoModerationRule(arg: {
    guild_id: string;
    /** The JSON body of the request. */ body: { name: any; event_type: number; actions?: any; enabled?: boolean; exempt_roles?: any; exempt_channels?: any; trigger_type: number; trigger_metadata?: any };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Create channel invite, POST /channels/{channel_id}/invites. Costs 1 credit. */
  createChannelInvite(arg: {
    channel_id: string;
    /** The JSON body of the request. */ body?: { max_age?: number; temporary?: boolean; max_uses?: number; unique?: boolean; target_user_id?: string; target_application_id?: string; target_type?: number; role_ids?: any; target_user_ids?: any };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Create dm, POST /users/@me/channels. Costs 1 credit. */
  createDm(arg?: {
    /** The JSON body of the request. */ body?: { recipient_id?: string; access_tokens?: any; nicks?: any };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Create guild channel, POST /guilds/{guild_id}/channels. Costs 1 credit. */
  createGuildChannel(arg: {
    guild_id: string;
    /** The JSON body of the request. */ body: { type?: number; name: any; position?: number; topic?: any; bitrate?: number; user_limit?: number; nsfw?: boolean; rate_limit_per_user?: number; parent_id?: string; permission_overwrites?: any; rtc_region?: any; video_quality_mode?: number; default_auto_archive_duration?: number; default_reaction_emoji?: any; default_thread_rate_limit_per_user?: number; default_sort_order?: number; default_forum_layout?: number; default_tag_setting?: any; available_tags?: any };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Create guild emoji, POST /guilds/{guild_id}/emojis. Costs 1 credit. */
  createGuildEmoji(arg: {
    guild_id: string;
    /** The JSON body of the request. */ body: { name: any; image: any; roles?: any };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Create guild role, POST /guilds/{guild_id}/roles. Costs 1 credit. */
  createGuildRole(arg: {
    guild_id: string;
    /** The JSON body of the request. */ body?: { name?: any; permissions?: number; color?: number; colors?: any; hoist?: boolean; mentionable?: boolean; icon?: any; unicode_emoji?: any };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Create guild scheduled event, POST /guilds/{guild_id}/scheduled-events. Costs 1 credit. */
  createGuildScheduledEvent(arg: {
    guild_id: string;
    /** The JSON body of the request. */ body: { name: any; description?: any; image?: any; scheduled_start_time: any; scheduled_end_time?: any; privacy_level: number; entity_type: number; channel_id?: string; recurrence_rule?: any; entity_metadata?: any };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Create guild scheduled event exception, POST /guilds/{guild_id}/scheduled-events/{guild_scheduled_event_id}/exceptions. Costs 1 credit. */
  createGuildScheduledEventException(arg: {
    guild_id: string;
    guild_scheduled_event_id: string;
    /** The JSON body of the request. */ body: { scheduled_start_time?: any; scheduled_end_time?: any; original_scheduled_start_time: any; is_canceled?: boolean };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Create guild soundboard sound, POST /guilds/{guild_id}/soundboard-sounds. Costs 1 credit. */
  createGuildSoundboardSound(arg: {
    guild_id: string;
    /** The JSON body of the request. */ body: { name: any; volume?: number; emoji_id?: string; emoji_name?: any; sound: any };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Create guild template, POST /guilds/{guild_id}/templates. Costs 1 credit. */
  createGuildTemplate(arg: {
    guild_id: string;
    /** The JSON body of the request. */ body: { name: any; description?: any };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Create linked lobby guild invite for self, POST /lobbies/{lobby_id}/members/@me/invites. Costs 1 credit. */
  createLinkedLobbyGuildInviteForSelf(arg: {
    lobby_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Create linked lobby guild invite for user, POST /lobbies/{lobby_id}/members/{user_id}/invites. Costs 1 credit. */
  createLinkedLobbyGuildInviteForUser(arg: {
    lobby_id: string;
    user_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Create lobby, POST /lobbies. Costs 1 credit. */
  createLobby(arg?: {
    /** The JSON body of the request. */ body?: { idle_timeout_seconds?: number; members?: any; metadata?: any; flags?: number; override_event_webhooks_url?: any };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Create lobby message, POST /lobbies/{lobby_id}/messages. Costs 1 credit. */
  createLobbyMessage(arg: {
    lobby_id: string;
    /** The JSON body of the request. */ body?: { content?: any; embeds?: any; allowed_mentions?: any; sticker_ids?: any; components?: any; flags?: number; attachments?: any; poll?: any; shared_client_theme?: any; message_reference?: any; nonce?: number; enforce_nonce?: boolean; tts?: boolean };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Create message, POST /channels/{channel_id}/messages. Costs 1 credit. */
  createMessage(arg: {
    channel_id: string;
    /** The JSON body of the request. */ body?: { content?: any; embeds?: any; allowed_mentions?: any; sticker_ids?: any; components?: any; flags?: number; attachments?: any; poll?: any; shared_client_theme?: any; message_reference?: any; nonce?: number; enforce_nonce?: boolean; tts?: boolean };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Create or join lobby, PUT /lobbies. Costs 1 credit. */
  createOrJoinLobby(arg: {
    /** The JSON body of the request. */ body: { idle_timeout_seconds?: number; lobby_metadata?: any; member_metadata?: any; secret: any; flags?: number };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Create pin, PUT /channels/{channel_id}/messages/pins/{message_id}. Costs 1 credit. */
  createPin(arg: {
    channel_id: string;
    message_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Create stage instance, POST /stage-instances. Costs 1 credit. */
  createStageInstance(arg: {
    /** The JSON body of the request. */ body: { topic: any; channel_id: string; privacy_level?: number; guild_scheduled_event_id?: string; send_start_notification?: boolean };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Create thread, POST /channels/{channel_id}/threads. Costs 1 credit. */
  createThread(arg: {
    channel_id: string;
    /** The JSON body of the request. */ body: { name: any; auto_archive_duration?: number; rate_limit_per_user?: number; applied_tags?: any; message?: any; type?: number; invitable?: boolean };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Create thread from message, POST /channels/{channel_id}/messages/{message_id}/threads. Costs 1 credit. */
  createThreadFromMessage(arg: {
    channel_id: string;
    message_id: string;
    /** The JSON body of the request. */ body: { name: any; auto_archive_duration?: number; rate_limit_per_user?: number };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Create webhook, POST /channels/{channel_id}/webhooks. Costs 1 credit. */
  createWebhook(arg: {
    channel_id: string;
    /** The JSON body of the request. */ body: { name: any; avatar?: any };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Crosspost message, POST /channels/{channel_id}/messages/{message_id}/crosspost. Costs 1 credit. */
  crosspostMessage(arg: {
    channel_id: string;
    message_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Delete all message reactions, DELETE /channels/{channel_id}/messages/{message_id}/reactions. Costs 1 credit. */
  deleteAllMessageReactions(arg: {
    channel_id: string;
    message_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Delete all message reactions by emoji, DELETE /channels/{channel_id}/messages/{message_id}/reactions/{emoji_name}. Costs 1 credit. */
  deleteAllMessageReactionsByEmoji(arg: {
    channel_id: string;
    message_id: string;
    emoji_name: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Delete auto moderation rule, DELETE /guilds/{guild_id}/auto-moderation/rules/{rule_id}. Costs 1 credit. */
  deleteAutoModerationRule(arg: {
    guild_id: string;
    rule_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Delete channel, DELETE /channels/{channel_id}. Costs 1 credit. */
  deleteChannel(arg: {
    channel_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Delete channel permission overwrite, DELETE /channels/{channel_id}/permissions/{overwrite_id}. Costs 1 credit. */
  deleteChannelPermissionOverwrite(arg: {
    channel_id: string;
    overwrite_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Delete group dm user, DELETE /channels/{channel_id}/recipients/{user_id}. Costs 1 credit. */
  deleteGroupDmUser(arg: {
    channel_id: string;
    user_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Delete guild emoji, DELETE /guilds/{guild_id}/emojis/{emoji_id}. Costs 1 credit. */
  deleteGuildEmoji(arg: {
    guild_id: string;
    emoji_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Delete guild integration, DELETE /guilds/{guild_id}/integrations/{integration_id}. Costs 1 credit. */
  deleteGuildIntegration(arg: {
    guild_id: string;
    integration_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Delete guild member, DELETE /guilds/{guild_id}/members/{user_id}. Costs 1 credit. */
  deleteGuildMember(arg: {
    guild_id: string;
    user_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Delete guild member role, DELETE /guilds/{guild_id}/members/{user_id}/roles/{role_id}. Costs 1 credit. */
  deleteGuildMemberRole(arg: {
    guild_id: string;
    user_id: string;
    role_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Delete guild role, DELETE /guilds/{guild_id}/roles/{role_id}. Costs 1 credit. */
  deleteGuildRole(arg: {
    guild_id: string;
    role_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Delete guild scheduled event, DELETE /guilds/{guild_id}/scheduled-events/{guild_scheduled_event_id}. Costs 1 credit. */
  deleteGuildScheduledEvent(arg: {
    guild_id: string;
    guild_scheduled_event_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Delete guild scheduled event exception, DELETE /guilds/{guild_id}/scheduled-events/{guild_scheduled_event_id}/exceptions/{exception_id}. Costs 1 credit. */
  deleteGuildScheduledEventException(arg: {
    guild_id: string;
    guild_scheduled_event_id: string;
    exception_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Delete guild soundboard sound, DELETE /guilds/{guild_id}/soundboard-sounds/{sound_id}. Costs 1 credit. */
  deleteGuildSoundboardSound(arg: {
    guild_id: string;
    sound_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Delete guild sticker, DELETE /guilds/{guild_id}/stickers/{sticker_id}. Costs 1 credit. */
  deleteGuildSticker(arg: {
    guild_id: string;
    sticker_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Delete guild template, DELETE /guilds/{guild_id}/templates/{code}. Costs 1 credit. */
  deleteGuildTemplate(arg: {
    guild_id: string;
    code: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Delete lobby, DELETE /lobbies/{lobby_id}. Costs 1 credit. */
  deleteLobby(arg: {
    lobby_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Delete lobby member, DELETE /lobbies/{lobby_id}/members/{user_id}. Costs 1 credit. */
  deleteLobbyMember(arg: {
    lobby_id: string;
    user_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Delete message, DELETE /channels/{channel_id}/messages/{message_id}. Costs 1 credit. */
  deleteMessage(arg: {
    channel_id: string;
    message_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Delete my message reaction, DELETE /channels/{channel_id}/messages/{message_id}/reactions/{emoji_name}/@me. Costs 1 credit. */
  deleteMyMessageReaction(arg: {
    channel_id: string;
    message_id: string;
    emoji_name: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Delete pin, DELETE /channels/{channel_id}/messages/pins/{message_id}. Costs 1 credit. */
  deletePin(arg: {
    channel_id: string;
    message_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Delete stage instance, DELETE /stage-instances/{channel_id}. Costs 1 credit. */
  deleteStageInstance(arg: {
    channel_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Delete thread member, DELETE /channels/{channel_id}/thread-members/{user_id}. Costs 1 credit. */
  deleteThreadMember(arg: {
    channel_id: string;
    user_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Delete user message reaction, DELETE /channels/{channel_id}/messages/{message_id}/reactions/{emoji_name}/{user_id}. Costs 1 credit. */
  deleteUserMessageReaction(arg: {
    channel_id: string;
    message_id: string;
    emoji_name: string;
    user_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Delete webhook, DELETE /webhooks/{webhook_id}. Costs 1 credit. */
  deleteWebhook(arg: {
    webhook_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Deprecated create pin, PUT /channels/{channel_id}/pins/{message_id}. Costs 1 credit. */
  deprecatedCreatePin(arg: {
    channel_id: string;
    message_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Deprecated delete pin, DELETE /channels/{channel_id}/pins/{message_id}. Costs 1 credit. */
  deprecatedDeletePin(arg: {
    channel_id: string;
    message_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Deprecated list pins, GET /channels/{channel_id}/pins. Costs nothing. */
  deprecatedListPins(arg: {
    channel_id: string;
  }): Promise<any>;
  /** Edit lobby, PATCH /lobbies/{lobby_id}. Costs 1 credit. */
  editLobby(arg: {
    lobby_id: string;
    /** The JSON body of the request. */ body?: { idle_timeout_seconds?: number; metadata?: any; members?: any; flags?: number; override_event_webhooks_url?: any };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Edit lobby channel link, PATCH /lobbies/{lobby_id}/channel-linking. Costs 1 credit. */
  editLobbyChannelLink(arg: {
    lobby_id: string;
    /** The JSON body of the request. */ body?: { channel_id?: string };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Follow channel, POST /channels/{channel_id}/followers. Costs 1 credit. */
  followChannel(arg: {
    channel_id: string;
    /** The JSON body of the request. */ body: { webhook_channel_id: string };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Get active guild threads, GET /guilds/{guild_id}/threads/active. Costs nothing. */
  getActiveGuildThreads(arg: {
    guild_id: string;
  }): Promise<any>;
  /** Get answer voters, GET /channels/{channel_id}/polls/{message_id}/answers/{answer_id}. Costs nothing. */
  getAnswerVoters(arg: {
    channel_id: string;
    message_id: string;
    answer_id: number;
    after?: string;
    limit?: number;
  }): Promise<any>;
  /** Get application, GET /applications/{application_id}. Costs nothing. */
  getApplication(arg: {
    application_id: string;
  }): Promise<any>;
  /** Get application command, GET /applications/{application_id}/commands/{command_id}. Costs nothing. */
  getApplicationCommand(arg: {
    application_id: string;
    command_id: string;
  }): Promise<any>;
  /** Get application emoji, GET /applications/{application_id}/emojis/{emoji_id}. Costs nothing. */
  getApplicationEmoji(arg: {
    application_id: string;
    emoji_id: string;
  }): Promise<any>;
  /** Get application role connections metadata, GET /applications/{application_id}/role-connections/metadata. Costs nothing. */
  getApplicationRoleConnectionsMetadata(arg: {
    application_id: string;
  }): Promise<any>;
  /** Get auto moderation rule, GET /guilds/{guild_id}/auto-moderation/rules/{rule_id}. Costs nothing. */
  getAutoModerationRule(arg: {
    guild_id: string;
    rule_id: string;
  }): Promise<any>;
  /** Get bot gateway, GET /gateway/bot. Costs nothing. */
  getBotGateway(): Promise<any>;
  /** Get channel, GET /channels/{channel_id}. Costs nothing. */
  getChannel(arg: {
    channel_id: string;
  }): Promise<any>;
  /** Get entitlement, GET /applications/{application_id}/entitlements/{entitlement_id}. Costs nothing. */
  getEntitlement(arg: {
    application_id: string;
    entitlement_id: string;
  }): Promise<any>;
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
  }): Promise<any>;
  /** Get gateway, GET /gateway. Costs nothing. */
  getGateway(): Promise<any>;
  /** Get guild, GET /guilds/{guild_id}. Costs nothing. */
  getGuild(arg: {
    guild_id: string;
    with_counts?: boolean;
  }): Promise<any>;
  /** Get guild application command, GET /applications/{application_id}/guilds/{guild_id}/commands/{command_id}. Costs nothing. */
  getGuildApplicationCommand(arg: {
    application_id: string;
    guild_id: string;
    command_id: string;
  }): Promise<any>;
  /** Get guild application command permissions, GET /applications/{application_id}/guilds/{guild_id}/commands/{command_id}/permissions. Costs nothing. */
  getGuildApplicationCommandPermissions(arg: {
    application_id: string;
    guild_id: string;
    command_id: string;
  }): Promise<any>;
  /** Get guild ban, GET /guilds/{guild_id}/bans/{user_id}. Costs nothing. */
  getGuildBan(arg: {
    guild_id: string;
    user_id: string;
  }): Promise<any>;
  /** Get guild emoji, GET /guilds/{guild_id}/emojis/{emoji_id}. Costs nothing. */
  getGuildEmoji(arg: {
    guild_id: string;
    emoji_id: string;
  }): Promise<any>;
  /** Get guild join requests, GET /guilds/{guild_id}/requests. Costs nothing. */
  getGuildJoinRequests(arg: {
    guild_id: string;
    status?: string;
    limit?: number;
    before?: string;
    after?: string;
  }): Promise<any>;
  /** Get guild member, GET /guilds/{guild_id}/members/{user_id}. Costs nothing. */
  getGuildMember(arg: {
    guild_id: string;
    user_id: string;
  }): Promise<any>;
  /** Get guild new member welcome, GET /guilds/{guild_id}/new-member-welcome. Costs nothing. */
  getGuildNewMemberWelcome(arg: {
    guild_id: string;
  }): Promise<any>;
  /** Get guild preview, GET /guilds/{guild_id}/preview. Costs nothing. */
  getGuildPreview(arg: {
    guild_id: string;
  }): Promise<any>;
  /** Get guild role, GET /guilds/{guild_id}/roles/{role_id}. Costs nothing. */
  getGuildRole(arg: {
    guild_id: string;
    role_id: string;
  }): Promise<any>;
  /** Get guild scheduled event, GET /guilds/{guild_id}/scheduled-events/{guild_scheduled_event_id}. Costs nothing. */
  getGuildScheduledEvent(arg: {
    guild_id: string;
    guild_scheduled_event_id: string;
    with_user_count?: boolean;
  }): Promise<any>;
  /** Get guild soundboard sound, GET /guilds/{guild_id}/soundboard-sounds/{sound_id}. Costs nothing. */
  getGuildSoundboardSound(arg: {
    guild_id: string;
    sound_id: string;
  }): Promise<any>;
  /** Get guild sticker, GET /guilds/{guild_id}/stickers/{sticker_id}. Costs nothing. */
  getGuildSticker(arg: {
    guild_id: string;
    sticker_id: string;
  }): Promise<any>;
  /** Get guild template, GET /guilds/templates/{code}. Costs nothing. */
  getGuildTemplate(arg: {
    code: string;
  }): Promise<any>;
  /** Get guild vanity url, GET /guilds/{guild_id}/vanity-url. Costs nothing. */
  getGuildVanityUrl(arg: {
    guild_id: string;
  }): Promise<any>;
  /** Get guild webhooks, GET /guilds/{guild_id}/webhooks. Costs nothing. */
  getGuildWebhooks(arg: {
    guild_id: string;
  }): Promise<any>;
  /** Get guild welcome screen, GET /guilds/{guild_id}/welcome-screen. Costs nothing. */
  getGuildWelcomeScreen(arg: {
    guild_id: string;
  }): Promise<any>;
  /** Get guild widget, GET /guilds/{guild_id}/widget.json. Costs nothing. */
  getGuildWidget(arg: {
    guild_id: string;
  }): Promise<any>;
  /** Get guild widget png, GET /guilds/{guild_id}/widget.png. Costs nothing. */
  getGuildWidgetPng(arg: {
    guild_id: string;
    style?: string;
  }): Promise<any>;
  /** Get guild widget settings, GET /guilds/{guild_id}/widget. Costs nothing. */
  getGuildWidgetSettings(arg: {
    guild_id: string;
  }): Promise<any>;
  /** Get guilds onboarding, GET /guilds/{guild_id}/onboarding. Costs nothing. */
  getGuildsOnboarding(arg: {
    guild_id: string;
  }): Promise<any>;
  /** Get invite target users, GET /invites/{code}/target-users. Costs nothing. */
  getInviteTargetUsers(arg: {
    code: string;
  }): Promise<any>;
  /** Get invite target users job status, GET /invites/{code}/target-users/job-status. Costs nothing. */
  getInviteTargetUsersJobStatus(arg: {
    code: string;
  }): Promise<any>;
  /** Get lobby, GET /lobbies/{lobby_id}. Costs nothing. */
  getLobby(arg: {
    lobby_id: string;
  }): Promise<any>;
  /** Get lobby messages, GET /lobbies/{lobby_id}/messages. Costs nothing. */
  getLobbyMessages(arg: {
    lobby_id: string;
    limit?: number;
  }): Promise<any>;
  /** Get message, GET /channels/{channel_id}/messages/{message_id}. Costs nothing. */
  getMessage(arg: {
    channel_id: string;
    message_id: string;
  }): Promise<any>;
  /** Get my application, GET /applications/@me. Costs nothing. */
  getMyApplication(): Promise<any>;
  /** Get my oauth2 application, GET /oauth2/applications/@me. Costs nothing. */
  getMyOauth2Application(): Promise<any>;
  /** Get my oauth2 authorization, GET /oauth2/@me. Costs nothing. */
  getMyOauth2Authorization(): Promise<any>;
  /** Get my user, GET /users/@me. Costs nothing. */
  getMyUser(): Promise<any>;
  /** Get openid connect userinfo, GET /oauth2/userinfo. Costs nothing. */
  getOpenidConnectUserinfo(): Promise<any>;
  /** Get public keys, GET /oauth2/keys. Costs nothing. */
  getPublicKeys(): Promise<any>;
  /** Get self voice state, GET /guilds/{guild_id}/voice-states/@me. Costs nothing. */
  getSelfVoiceState(arg: {
    guild_id: string;
  }): Promise<any>;
  /** Get sku subscription, GET /skus/{sku_id}/subscriptions/{subscription_id}. Costs nothing. */
  getSkuSubscription(arg: {
    sku_id: string;
    subscription_id: string;
    user_id?: string;
  }): Promise<any>;
  /** Get sku subscriptions, GET /skus/{sku_id}/subscriptions. Costs nothing. */
  getSkuSubscriptions(arg: {
    sku_id: string;
    before?: string;
    after?: string;
    limit?: number;
    user_id?: string;
  }): Promise<any>;
  /** Get soundboard default sounds, GET /soundboard-default-sounds. Costs nothing. */
  getSoundboardDefaultSounds(): Promise<any>;
  /** Get stage instance, GET /stage-instances/{channel_id}. Costs nothing. */
  getStageInstance(arg: {
    channel_id: string;
  }): Promise<any>;
  /** Get sticker, GET /stickers/{sticker_id}. Costs nothing. */
  getSticker(arg: {
    sticker_id: string;
  }): Promise<any>;
  /** Get sticker pack, GET /sticker-packs/{pack_id}. Costs nothing. */
  getStickerPack(arg: {
    pack_id: string;
  }): Promise<any>;
  /** Get thread member, GET /channels/{channel_id}/thread-members/{user_id}. Costs nothing. */
  getThreadMember(arg: {
    channel_id: string;
    user_id: string;
    with_member?: boolean;
  }): Promise<any>;
  /** Get user, GET /users/{user_id}. Costs nothing. */
  getUser(arg: {
    user_id: string;
  }): Promise<any>;
  /** Get voice state, GET /guilds/{guild_id}/voice-states/{user_id}. Costs nothing. */
  getVoiceState(arg: {
    guild_id: string;
    user_id: string;
  }): Promise<any>;
  /** Get webhook, GET /webhooks/{webhook_id}. Costs nothing. */
  getWebhook(arg: {
    webhook_id: string;
  }): Promise<any>;
  /** Guild role member counts, GET /guilds/{guild_id}/roles/member-counts. Costs nothing. */
  guildRoleMemberCounts(arg: {
    guild_id: string;
  }): Promise<any>;
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
  }): Promise<any>;
  /** Invite resolve, GET /invites/{code}. Costs nothing. */
  inviteResolve(arg: {
    code: string;
    with_counts?: boolean;
    guild_scheduled_event_id?: string;
    target_channel_id?: string;
    target_message_id?: string;
  }): Promise<any>;
  /** Invite revoke, DELETE /invites/{code}. Costs 1 credit. */
  inviteRevoke(arg: {
    code: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Join thread, PUT /channels/{channel_id}/thread-members/@me. Costs 1 credit. */
  joinThread(arg: {
    channel_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Leave lobby, DELETE /lobbies/{lobby_id}/members/@me. Costs 1 credit. */
  leaveLobby(arg: {
    lobby_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Leave thread, DELETE /channels/{channel_id}/thread-members/@me. Costs 1 credit. */
  leaveThread(arg: {
    channel_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** List application commands, GET /applications/{application_id}/commands. Costs nothing. */
  listApplicationCommands(arg: {
    application_id: string;
    with_localizations?: boolean;
  }): Promise<any>;
  /** List application emojis, GET /applications/{application_id}/emojis. Costs nothing. */
  listApplicationEmojis(arg: {
    application_id: string;
  }): Promise<any>;
  /** List auto moderation rules, GET /guilds/{guild_id}/auto-moderation/rules. Costs nothing. */
  listAutoModerationRules(arg: {
    guild_id: string;
  }): Promise<any>;
  /** List channel invites, GET /channels/{channel_id}/invites. Costs nothing. */
  listChannelInvites(arg: {
    channel_id: string;
  }): Promise<any>;
  /** List channel webhooks, GET /channels/{channel_id}/webhooks. Costs nothing. */
  listChannelWebhooks(arg: {
    channel_id: string;
  }): Promise<any>;
  /** List guild application command permissions, GET /applications/{application_id}/guilds/{guild_id}/commands/permissions. Costs nothing. */
  listGuildApplicationCommandPermissions(arg: {
    application_id: string;
    guild_id: string;
  }): Promise<any>;
  /** List guild application commands, GET /applications/{application_id}/guilds/{guild_id}/commands. Costs nothing. */
  listGuildApplicationCommands(arg: {
    application_id: string;
    guild_id: string;
    with_localizations?: boolean;
  }): Promise<any>;
  /** List guild audit log entries, GET /guilds/{guild_id}/audit-logs. Costs nothing. */
  listGuildAuditLogEntries(arg: {
    guild_id: string;
    user_id?: string;
    target_id?: string;
    action_type?: number;
    before?: string;
    after?: string;
    limit?: number;
  }): Promise<any>;
  /** List guild bans, GET /guilds/{guild_id}/bans. Costs nothing. */
  listGuildBans(arg: {
    guild_id: string;
    limit?: number;
    before?: string;
    after?: string;
  }): Promise<any>;
  /** List guild channels, GET /guilds/{guild_id}/channels. Costs nothing. */
  listGuildChannels(arg: {
    guild_id: string;
  }): Promise<any>;
  /** List guild emojis, GET /guilds/{guild_id}/emojis. Costs nothing. */
  listGuildEmojis(arg: {
    guild_id: string;
  }): Promise<any>;
  /** List guild integrations, GET /guilds/{guild_id}/integrations. Costs nothing. */
  listGuildIntegrations(arg: {
    guild_id: string;
  }): Promise<any>;
  /** List guild invites, GET /guilds/{guild_id}/invites. Costs nothing. */
  listGuildInvites(arg: {
    guild_id: string;
  }): Promise<any>;
  /** List guild members, GET /guilds/{guild_id}/members. Costs nothing. */
  listGuildMembers(arg: {
    guild_id: string;
    limit?: number;
    after?: string;
  }): Promise<any>;
  /** List guild roles, GET /guilds/{guild_id}/roles. Costs nothing. */
  listGuildRoles(arg: {
    guild_id: string;
  }): Promise<any>;
  /** List guild scheduled event exception users, GET /guilds/{guild_id}/scheduled-events/{guild_scheduled_event_id}/{guild_scheduled_event_exception_id}/users. Costs nothing. */
  listGuildScheduledEventExceptionUsers(arg: {
    guild_id: string;
    guild_scheduled_event_id: string;
    guild_scheduled_event_exception_id: string;
    with_member?: boolean;
    limit?: number;
    before?: string;
    after?: string;
  }): Promise<any>;
  /** List guild scheduled event users, GET /guilds/{guild_id}/scheduled-events/{guild_scheduled_event_id}/users. Costs nothing. */
  listGuildScheduledEventUsers(arg: {
    guild_id: string;
    guild_scheduled_event_id: string;
    with_member?: boolean;
    limit?: number;
    before?: string;
    after?: string;
  }): Promise<any>;
  /** List guild scheduled events, GET /guilds/{guild_id}/scheduled-events. Costs nothing. */
  listGuildScheduledEvents(arg: {
    guild_id: string;
    with_user_count?: boolean;
  }): Promise<any>;
  /** List guild soundboard sounds, GET /guilds/{guild_id}/soundboard-sounds. Costs nothing. */
  listGuildSoundboardSounds(arg: {
    guild_id: string;
  }): Promise<any>;
  /** List guild stickers, GET /guilds/{guild_id}/stickers. Costs nothing. */
  listGuildStickers(arg: {
    guild_id: string;
  }): Promise<any>;
  /** List guild templates, GET /guilds/{guild_id}/templates. Costs nothing. */
  listGuildTemplates(arg: {
    guild_id: string;
  }): Promise<any>;
  /** List guild voice regions, GET /guilds/{guild_id}/regions. Costs nothing. */
  listGuildVoiceRegions(arg: {
    guild_id: string;
  }): Promise<any>;
  /** List message reactions by emoji, GET /channels/{channel_id}/messages/{message_id}/reactions/{emoji_name}. Costs nothing. */
  listMessageReactionsByEmoji(arg: {
    channel_id: string;
    message_id: string;
    emoji_name: string;
    after?: string;
    limit?: number;
    type?: number;
  }): Promise<any>;
  /** List messages, GET /channels/{channel_id}/messages. Costs nothing. */
  listMessages(arg: {
    channel_id: string;
    around?: string;
    before?: string;
    after?: string;
    limit?: number;
  }): Promise<any>;
  /** List my connections, GET /users/@me/connections. Costs nothing. */
  listMyConnections(): Promise<any>;
  /** List my guilds, GET /users/@me/guilds. Costs nothing. */
  listMyGuilds(arg?: {
    before?: string;
    after?: string;
    limit?: number;
    shard?: number;
    with_counts?: boolean;
  }): Promise<any>;
  /** List my private archived threads, GET /channels/{channel_id}/users/@me/threads/archived/private. Costs nothing. */
  listMyPrivateArchivedThreads(arg: {
    channel_id: string;
    before?: string;
    limit?: number;
  }): Promise<any>;
  /** List pins, GET /channels/{channel_id}/messages/pins. Costs nothing. */
  listPins(arg: {
    channel_id: string;
    before?: string;
    limit?: number;
  }): Promise<any>;
  /** List private archived threads, GET /channels/{channel_id}/threads/archived/private. Costs nothing. */
  listPrivateArchivedThreads(arg: {
    channel_id: string;
    before?: string;
    limit?: number;
  }): Promise<any>;
  /** List public archived threads, GET /channels/{channel_id}/threads/archived/public. Costs nothing. */
  listPublicArchivedThreads(arg: {
    channel_id: string;
    before?: string;
    limit?: number;
  }): Promise<any>;
  /** List sticker packs, GET /sticker-packs. Costs nothing. */
  listStickerPacks(): Promise<any>;
  /** List thread members, GET /channels/{channel_id}/thread-members. Costs nothing. */
  listThreadMembers(arg: {
    channel_id: string;
    with_member?: boolean;
    limit?: number;
    after?: string;
  }): Promise<any>;
  /** List voice regions, GET /voice/regions. Costs nothing. */
  listVoiceRegions(): Promise<any>;
  /** Partner sdk token, POST /partner-sdk/token. Costs 1 credit. */
  partnerSdkToken(arg: {
    /** The JSON body of the request. */ body: { client_id: string; client_secret?: any; external_auth_token: any; external_auth_type: any };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Partner sdk unmerge provisional account, POST /partner-sdk/provisional-accounts/unmerge. Costs 1 credit. */
  partnerSdkUnmergeProvisionalAccount(arg: {
    /** The JSON body of the request. */ body: { client_id: string; client_secret?: any; external_auth_token: any; external_auth_type: any };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Poll expire, POST /channels/{channel_id}/polls/{message_id}/expire. Costs 1 credit. */
  pollExpire(arg: {
    channel_id: string;
    message_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Preview prune guild, GET /guilds/{guild_id}/prune. Costs nothing. */
  previewPruneGuild(arg: {
    guild_id: string;
    days?: number;
    include_roles?: string;
  }): Promise<any>;
  /** Put guilds onboarding, PUT /guilds/{guild_id}/onboarding. Costs 1 credit. */
  putGuildsOnboarding(arg: {
    guild_id: string;
    /** The JSON body of the request. */ body?: { prompts?: any; enabled?: boolean; default_channel_ids?: any; mode?: number };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Remove invite target user, DELETE /invites/{code}/target-users/{user_id}. Costs 1 credit. */
  removeInviteTargetUser(arg: {
    code: string;
    user_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Search guild members, GET /guilds/{guild_id}/members/search. Costs nothing. */
  searchGuildMembers(arg: {
    guild_id: string;
    limit?: number;
    query: string;
  }): Promise<any>;
  /** Send soundboard sound, POST /channels/{channel_id}/send-soundboard-sound. Costs 1 credit. */
  sendSoundboardSound(arg: {
    channel_id: string;
    /** The JSON body of the request. */ body: { sound_id: string; source_guild_id?: string };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Set channel permission overwrite, PUT /channels/{channel_id}/permissions/{overwrite_id}. Costs 1 credit. */
  setChannelPermissionOverwrite(arg: {
    channel_id: string;
    overwrite_id: string;
    /** The JSON body of the request. */ body?: { type?: number; allow?: number; deny?: number };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Sync guild template, PUT /guilds/{guild_id}/templates/{code}. Costs 1 credit. */
  syncGuildTemplate(arg: {
    guild_id: string;
    code: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
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
  }): Promise<any>;
  /** Trigger typing indicator, POST /channels/{channel_id}/typing. Costs 1 credit. */
  triggerTypingIndicator(arg: {
    channel_id: string;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Unban user from guild, DELETE /guilds/{guild_id}/bans/{user_id}. Costs 1 credit. */
  unbanUserFromGuild(arg: {
    guild_id: string;
    user_id: string;
    /** The JSON body of the request. */ body?: any;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Update auto moderation rule, PATCH /guilds/{guild_id}/auto-moderation/rules/{rule_id}. Costs 1 credit. */
  updateAutoModerationRule(arg: {
    guild_id: string;
    rule_id: string;
    /** The JSON body of the request. */ body?: { name?: any; event_type?: number; actions?: any; enabled?: boolean; exempt_roles?: any; exempt_channels?: any; trigger_type?: number; trigger_metadata?: any };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Update channel, PATCH /channels/{channel_id}. Costs 1 credit. */
  updateChannel(arg: {
    channel_id: string;
    /** The JSON body of the request. */ body?: { name?: any; icon?: any; type?: number; position?: number; topic?: any; bitrate?: number; user_limit?: number; nsfw?: boolean; rate_limit_per_user?: number; parent_id?: string; permission_overwrites?: any; rtc_region?: any; video_quality_mode?: number; default_auto_archive_duration?: number; default_reaction_emoji?: any; default_thread_rate_limit_per_user?: number; default_sort_order?: number; default_forum_layout?: number; default_tag_setting?: any; flags?: number; available_tags?: any; archived?: boolean; locked?: boolean; invitable?: boolean; auto_archive_duration?: number; applied_tags?: any };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Update guild, PATCH /guilds/{guild_id}. Costs 1 credit. */
  updateGuild(arg: {
    guild_id: string;
    /** The JSON body of the request. */ body?: { name?: any; description?: any; region?: any; icon?: any; verification_level?: number; default_message_notifications?: number; explicit_content_filter?: number; preferred_locale?: any; afk_timeout?: number; afk_channel_id?: string; system_channel_id?: string; splash?: any; banner?: any; system_channel_flags?: number; features?: any; discovery_splash?: any; home_header?: any; rules_channel_id?: string; safety_alerts_channel_id?: string; public_updates_channel_id?: string; premium_progress_bar_enabled?: boolean };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Update guild emoji, PATCH /guilds/{guild_id}/emojis/{emoji_id}. Costs 1 credit. */
  updateGuildEmoji(arg: {
    guild_id: string;
    emoji_id: string;
    /** The JSON body of the request. */ body?: { name?: any; roles?: any };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Update guild incident actions, PUT /guilds/{guild_id}/incident-actions. Costs 1 credit. */
  updateGuildIncidentActions(arg: {
    guild_id: string;
    /** The JSON body of the request. */ body?: { invites_disabled_until?: any; dms_disabled_until?: any };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Update guild member, PATCH /guilds/{guild_id}/members/{user_id}. Costs 1 credit. */
  updateGuildMember(arg: {
    guild_id: string;
    user_id: string;
    /** The JSON body of the request. */ body?: { nick?: any; roles?: any; mute?: boolean; deaf?: boolean; channel_id?: string; communication_disabled_until?: any; flags?: number };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Update guild role, PATCH /guilds/{guild_id}/roles/{role_id}. Costs 1 credit. */
  updateGuildRole(arg: {
    guild_id: string;
    role_id: string;
    /** The JSON body of the request. */ body?: { name?: any; permissions?: number; color?: number; colors?: any; hoist?: boolean; mentionable?: boolean; icon?: any; unicode_emoji?: any };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Update guild scheduled event, PATCH /guilds/{guild_id}/scheduled-events/{guild_scheduled_event_id}. Costs 1 credit. */
  updateGuildScheduledEvent(arg: {
    guild_id: string;
    guild_scheduled_event_id: string;
    /** The JSON body of the request. */ body?: { status?: number; name?: any; description?: any; image?: any; scheduled_start_time?: any; scheduled_end_time?: any; entity_type?: number; privacy_level?: number; channel_id?: string; recurrence_rule?: any; entity_metadata?: any };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Update guild scheduled event exception, PATCH /guilds/{guild_id}/scheduled-events/{guild_scheduled_event_id}/exceptions/{exception_id}. Costs 1 credit. */
  updateGuildScheduledEventException(arg: {
    guild_id: string;
    guild_scheduled_event_id: string;
    exception_id: string;
    /** The JSON body of the request. */ body?: { scheduled_start_time?: any; scheduled_end_time?: any; is_canceled?: boolean };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Update guild soundboard sound, PATCH /guilds/{guild_id}/soundboard-sounds/{sound_id}. Costs 1 credit. */
  updateGuildSoundboardSound(arg: {
    guild_id: string;
    sound_id: string;
    /** The JSON body of the request. */ body?: { name?: any; volume?: number; emoji_id?: string; emoji_name?: any };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Update guild sticker, PATCH /guilds/{guild_id}/stickers/{sticker_id}. Costs 1 credit. */
  updateGuildSticker(arg: {
    guild_id: string;
    sticker_id: string;
    /** The JSON body of the request. */ body?: { name?: any; tags?: any; description?: any };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Update guild template, PATCH /guilds/{guild_id}/templates/{code}. Costs 1 credit. */
  updateGuildTemplate(arg: {
    guild_id: string;
    code: string;
    /** The JSON body of the request. */ body?: { name?: any; description?: any };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Update guild welcome screen, PATCH /guilds/{guild_id}/welcome-screen. Costs 1 credit. */
  updateGuildWelcomeScreen(arg: {
    guild_id: string;
    /** The JSON body of the request. */ body?: { description?: any; welcome_channels?: any; enabled?: boolean };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Update guild widget settings, PATCH /guilds/{guild_id}/widget. Costs 1 credit. */
  updateGuildWidgetSettings(arg: {
    guild_id: string;
    /** The JSON body of the request. */ body?: { channel_id?: string; enabled?: boolean };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Update lobby message external moderation metadata, PUT /lobbies/{lobby_id}/messages/{message_id}/moderation-metadata. Costs 1 credit. */
  updateLobbyMessageExternalModerationMetadata(arg: {
    lobby_id: string;
    message_id: string;
    /** The JSON body of the request. */ body?: any;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Update message, PATCH /channels/{channel_id}/messages/{message_id}. Costs 1 credit. */
  updateMessage(arg: {
    channel_id: string;
    message_id: string;
    /** The JSON body of the request. */ body?: { content?: any; embeds?: any; flags?: number; allowed_mentions?: any; sticker_ids?: any; components?: any; attachments?: any };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Update self voice state, PATCH /guilds/{guild_id}/voice-states/@me. Costs 1 credit. */
  updateSelfVoiceState(arg: {
    guild_id: string;
    /** The JSON body of the request. */ body?: { request_to_speak_timestamp?: any; suppress?: boolean; channel_id?: string };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Update stage instance, PATCH /stage-instances/{channel_id}. Costs 1 credit. */
  updateStageInstance(arg: {
    channel_id: string;
    /** The JSON body of the request. */ body?: { topic?: any; privacy_level?: number };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Update user message external moderation metadata, PUT /partner-sdk/dms/{user_id_1}/{user_id_2}/messages/{message_id}/moderation-metadata. Costs 1 credit. */
  updateUserMessageExternalModerationMetadata(arg: {
    user_id_1: string;
    user_id_2: string;
    message_id: string;
    /** The JSON body of the request. */ body?: any;
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Update voice channel status, PUT /channels/{channel_id}/voice-status. Costs 1 credit. */
  updateVoiceChannelStatus(arg: {
    channel_id: string;
    /** The JSON body of the request. */ body?: { status?: any };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Update voice state, PATCH /guilds/{guild_id}/voice-states/{user_id}. Costs 1 credit. */
  updateVoiceState(arg: {
    guild_id: string;
    user_id: string;
    /** The JSON body of the request. */ body?: { suppress?: boolean; channel_id?: string };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
  /** Update webhook, PATCH /webhooks/{webhook_id}. Costs 1 credit. */
  updateWebhook(arg: {
    webhook_id: string;
    /** The JSON body of the request. */ body?: { name?: any; avatar?: any; channel_id?: string };
    /** The reason in the server's audit log. */ reason?: string;
  }): Promise<any>;
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
