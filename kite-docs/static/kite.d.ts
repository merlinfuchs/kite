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
  /** Applications get activity instance, GET /applications/{application_id}/activity-instances/{instance_id}. Costs nothing. */
  applicationsGetActivityInstance(arg: {
    application_id: string;
    instance_id: string;
  }): Promise<any>;
  /** Count guild scheduled event users, GET /guilds/{guild_id}/scheduled-events/{guild_scheduled_event_id}/users/counts. Costs nothing. */
  countGuildScheduledEventUsers(arg: {
    guild_id: string;
    guild_scheduled_event_id: string;
    guild_scheduled_event_exception_ids?: (string | number)[];
  }): Promise<any>;
  /** Deprecated list pins, GET /channels/{channel_id}/pins. Costs nothing. */
  deprecatedListPins(arg: {
    channel_id: string;
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
  /** Preview prune guild, GET /guilds/{guild_id}/prune. Costs nothing. */
  previewPruneGuild(arg: {
    guild_id: string;
    days?: number;
    include_roles?: string;
  }): Promise<any>;
  /** Search guild members, GET /guilds/{guild_id}/members/search. Costs nothing. */
  searchGuildMembers(arg: {
    guild_id: string;
    limit?: number;
    query: string;
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
};

declare const cookieApi: {
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
  /** Add an item to the end of the list in a stored variable, or text to the end of other values, and get the new value. Set the variable to an empty list first to start a list. Calls at the same time don't overwrite each other. Costs 1 credit. */
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
