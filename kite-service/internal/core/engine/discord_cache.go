package engine

import (
	"encoding/json"
	"net/http"
	"regexp"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
)

// Reads of Discord's API that the gateway's cache can answer, for code and the
// Discord API Request block. The cache is kept up to date by the gateway
// events of the default intents, except for members, which need the Server
// Members intent. Each answer has the fields Discord's API returns, which
// TestCachedAPIResponsesMatchTheSpec checks against Discord's spec.

// cachedRoutes are the reads the cache answers, by path. answer gets the IDs
// in the path and returns false if the cache can't answer.
var cachedRoutes = []struct {
	path   *regexp.Regexp
	answer func(p *DiscordProvider, ids []discord.Snowflake) (any, bool)
}{
	{regexp.MustCompile(`^/guilds/(\d+)/members/(\d+)$`), (*DiscordProvider).cachedMember},
	{regexp.MustCompile(`^/guilds/(\d+)/roles$`), (*DiscordProvider).cachedRoles},
	{regexp.MustCompile(`^/guilds/(\d+)/roles/(\d+)$`), (*DiscordProvider).cachedRole},
	{regexp.MustCompile(`^/channels/(\d+)$`), (*DiscordProvider).cachedChannel},
	{regexp.MustCompile(`^/guilds/(\d+)/channels$`), (*DiscordProvider).cachedChannels},
}

// cachedAPIResponse answers req from the cache, or returns false when the
// cache can't, like for a member it doesn't have. Requests with a query
// always go to Discord.
func (p *DiscordProvider) cachedAPIResponse(req provider.DiscordAPIRequest) ([]byte, bool) {
	if req.Method != http.MethodGet {
		return nil, false
	}
	for _, route := range cachedRoutes {
		m := route.path.FindStringSubmatch(req.Path)
		if m == nil {
			continue
		}
		ids := make([]discord.Snowflake, len(m)-1)
		for i, s := range m[1:] {
			ids[i], _ = discord.ParseSnowflake(s)
		}
		res, ok := route.answer(p, ids)
		if !ok {
			return nil, false
		}
		b, err := json.Marshal(res)
		return b, err == nil
	}
	return nil, false
}

func (p *DiscordProvider) cachedMember(ids []discord.Snowflake) (any, bool) {
	// Without the intent, members in the cache come from messages and
	// interactions, and their roles and nickname can be outdated.
	if !p.session.HasIntents(gateway.IntentGuildMembers) {
		return nil, false
	}
	member, err := p.session.Cabinet.Member(discord.GuildID(ids[0]), discord.UserID(ids[1]))
	if err != nil {
		return nil, false
	}
	return memberJSON(*member), true
}

func (p *DiscordProvider) cachedRoles(ids []discord.Snowflake) (any, bool) {
	roles, err := p.session.Cabinet.Roles(discord.GuildID(ids[0]))
	if err != nil {
		return nil, false
	}
	list := make([]map[string]any, len(roles))
	for i, r := range roles {
		list[i] = roleJSON(r)
	}
	return list, true
}

func (p *DiscordProvider) cachedRole(ids []discord.Snowflake) (any, bool) {
	role, err := p.session.Cabinet.Role(discord.GuildID(ids[0]), discord.RoleID(ids[1]))
	if err != nil {
		return nil, false
	}
	return roleJSON(*role), true
}

func (p *DiscordProvider) cachedChannel(ids []discord.Snowflake) (any, bool) {
	channel, err := p.session.Cabinet.Channel(discord.ChannelID(ids[0]))
	if err != nil || !cachedChannelType(channel) {
		return nil, false
	}
	// The cache has every channel of the server, also those the bot can't
	// see, for which Discord answers 403.
	if !p.cachedCanView(channel) {
		return nil, false
	}
	return channelJSON(*channel), true
}

// cachedCanView reports whether the bot can see channel, from the cache
// only, unlike state.Permissions, which asks Discord for what's missing.
func (p *DiscordProvider) cachedCanView(channel *discord.Channel) bool {
	cabinet := p.session.Cabinet
	me, err := cabinet.Me()
	if err != nil {
		return false
	}
	guild, err := cabinet.Guild(channel.GuildID)
	if err != nil {
		return false
	}
	member, err := cabinet.Member(channel.GuildID, me.ID)
	if err != nil {
		return false
	}
	roles, err := cabinet.Roles(channel.GuildID)
	if err != nil {
		return false
	}
	return discord.CalcOverrides(*guild, *channel, *member, roles).Has(discord.PermissionViewChannel)
}

func (p *DiscordProvider) cachedChannels(ids []discord.Snowflake) (any, bool) {
	channels, err := p.session.Cabinet.Channels(discord.GuildID(ids[0]))
	if err != nil {
		return nil, false
	}
	// Like Discord, without threads.
	list := make([]map[string]any, 0, len(channels))
	for _, c := range channels {
		if cachedChannelType(&c) {
			list = append(list, channelJSON(c))
		}
	}
	return list, true
}

// cachedChannelType reports whether c is a channel of a server and not a
// thread, whose API objects have fields the cache doesn't keep up to date,
// like the thread member of the bot.
func cachedChannelType(c *discord.Channel) bool {
	switch c.Type {
	case discord.GuildAnnouncementThread, discord.GuildPublicThread, discord.GuildPrivateThread:
		return false
	}
	return c.GuildID.IsValid()
}

// toJSONMap turns v into the map of its JSON.
func toJSONMap(v any) map[string]any {
	b, _ := json.Marshal(v)
	var m map[string]any
	_ = json.Unmarshal(b, &m)
	return m
}

// nullIfEmpty sets the fields that arikawa leaves out or writes as "" to null,
// like Discord does.
func nullIfEmpty(m map[string]any, keys ...string) {
	for _, k := range keys {
		if v, ok := m[k]; !ok || v == "" {
			m[k] = nil
		}
	}
}

// withDefault sets the fields arikawa leaves out when they're zero.
func withDefault(m map[string]any, defaults map[string]any) {
	for k, v := range defaults {
		if _, ok := m[k]; !ok {
			m[k] = v
		}
	}
}

func memberJSON(member discord.Member) map[string]any {
	m := toJSONMap(member)
	nullIfEmpty(m, "nick", "avatar", "banner", "premium_since", "communication_disabled_until")
	if member.RoleIDs == nil {
		m["roles"] = []any{}
	}
	m["user"] = userJSON(member.User)
	return m
}

func userJSON(user discord.User) map[string]any {
	m := toJSONMap(user)
	nullIfEmpty(m, "avatar", "global_name", "banner")
	withDefault(m, map[string]any{"flags": 0, "public_flags": 0})
	if user.PrimaryGuild != nil {
		g := toJSONMap(user.PrimaryGuild)
		nullIfEmpty(g, "tag", "badge")
		m["primary_guild"] = g
	} else {
		m["primary_guild"] = nil
	}
	return m
}

func roleJSON(role discord.Role) map[string]any {
	m := toJSONMap(role)
	nullIfEmpty(m, "icon", "unicode_emoji")
	// Roles cached before Discord sent colors.
	if role.Colors == (discord.RoleColors{}) {
		m["colors"] = toJSONMap(discord.RoleColors{
			Primary:   role.Color,
			Secondary: discord.NullColor,
			Tertiary:  discord.NullColor,
		})
	}
	return m
}

func channelJSON(channel discord.Channel) map[string]any {
	m := toJSONMap(channel)
	withDefault(m, map[string]any{
		"type":                  0,
		"position":              0,
		"flags":                 0,
		"nsfw":                  false,
		"permission_overwrites": []any{},
	})
	nullIfEmpty(m, "parent_id")
	// Only Discord's answers to interactions have the bot's permissions. The
	// cache doesn't update the last message and pin, so they're left out.
	delete(m, "permissions")
	delete(m, "last_message_id")
	delete(m, "last_pin_timestamp")
	switch channel.Type {
	case discord.GuildText, discord.GuildAnnouncement, discord.GuildForum:
		nullIfEmpty(m, "topic")
		withDefault(m, map[string]any{"rate_limit_per_user": 0})
	case discord.GuildVoice, discord.GuildStageVoice:
		nullIfEmpty(m, "rtc_region")
		withDefault(m, map[string]any{"rate_limit_per_user": 0, "bitrate": 0, "user_limit": 0})
	}
	return m
}
