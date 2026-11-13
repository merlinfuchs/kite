package engine

import (
	"encoding/json"
	"net/http"
	"regexp"
	"slices"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
)

// Reads of Discord's API that the gateway's cache can answer, for code and the
// Discord API Request block. The cache is kept up to date by the gateway
// events of the default intents, except for members, which need the Server
// Members intent. Each answer has the fields Discord's API returns, which
// TestCachedAPIResponsesMatchTheSpec checks against Discord's spec.
var (
	cachedMemberPath   = regexp.MustCompile(`^/guilds/(\d+)/members/(\d+)$`)
	cachedRolesPath    = regexp.MustCompile(`^/guilds/(\d+)/roles$`)
	cachedRolePath     = regexp.MustCompile(`^/guilds/(\d+)/roles/(\d+)$`)
	cachedChannelPath  = regexp.MustCompile(`^/channels/(\d+)$`)
	cachedChannelsPath = regexp.MustCompile(`^/guilds/(\d+)/channels$`)
)

// cachedAPIResponse answers req from the cache, or returns false when the
// cache can't, like for a member it doesn't have. Requests with a query
// always go to Discord.
func (p *DiscordProvider) cachedAPIResponse(req provider.DiscordAPIRequest) ([]byte, bool) {
	if req.Method != http.MethodGet {
		return nil, false
	}
	cabinet := p.session.Cabinet

	var res any
	switch {
	case cachedMemberPath.MatchString(req.Path):
		// Without the intent, members in the cache come from messages and
		// interactions, and their roles and nickname can be outdated.
		if !p.session.HasIntents(gateway.IntentGuildMembers) {
			return nil, false
		}
		m := cachedMemberPath.FindStringSubmatch(req.Path)
		member, err := cabinet.Member(discord.GuildID(snowflake(m[1])), discord.UserID(snowflake(m[2])))
		if err != nil {
			return nil, false
		}
		res = memberJSON(*member)
	case cachedRolesPath.MatchString(req.Path):
		m := cachedRolesPath.FindStringSubmatch(req.Path)
		roles, err := cabinet.Roles(discord.GuildID(snowflake(m[1])))
		if err != nil {
			return nil, false
		}
		list := make([]map[string]any, len(roles))
		for i, r := range roles {
			list[i] = roleJSON(r)
		}
		res = list
	case cachedRolePath.MatchString(req.Path):
		m := cachedRolePath.FindStringSubmatch(req.Path)
		role, err := cabinet.Role(discord.GuildID(snowflake(m[1])), discord.RoleID(snowflake(m[2])))
		if err != nil {
			return nil, false
		}
		res = roleJSON(*role)
	case cachedChannelPath.MatchString(req.Path):
		m := cachedChannelPath.FindStringSubmatch(req.Path)
		channel, err := cabinet.Channel(discord.ChannelID(snowflake(m[1])))
		if err != nil || !cachedChannelType(channel) {
			return nil, false
		}
		res = channelJSON(*channel)
	case cachedChannelsPath.MatchString(req.Path):
		m := cachedChannelsPath.FindStringSubmatch(req.Path)
		channels, err := cabinet.Channels(discord.GuildID(snowflake(m[1])))
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
		res = list
	default:
		return nil, false
	}

	b, err := json.Marshal(res)
	if err != nil {
		return nil, false
	}
	return b, true
}

func snowflake(s string) discord.Snowflake {
	id, _ := discord.ParseSnowflake(s)
	return id
}

// cachedChannelType reports whether c is a channel of a server and not a
// thread, whose API objects have fields the cache doesn't keep up to date,
// like the thread member of the bot.
func cachedChannelType(c *discord.Channel) bool {
	return c.GuildID.IsValid() && !slices.Contains([]discord.ChannelType{
		discord.GuildAnnouncementThread,
		discord.GuildPublicThread,
		discord.GuildPrivateThread,
	}, c.Type)
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
	// Only Discord's answers to interactions have the bot's permissions.
	delete(m, "permissions")
	switch channel.Type {
	case discord.GuildText, discord.GuildAnnouncement, discord.GuildForum:
		nullIfEmpty(m, "topic", "last_message_id")
		withDefault(m, map[string]any{"rate_limit_per_user": 0})
	case discord.GuildVoice, discord.GuildStageVoice:
		nullIfEmpty(m, "rtc_region", "last_message_id")
		withDefault(m, map[string]any{"rate_limit_per_user": 0, "bitrate": 0, "user_limit": 0})
	}
	return m
}
