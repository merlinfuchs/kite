package engine

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The TypeScript declarations of the objects of Discord's API, generated from
// its spec by kite-web/scripts/discord-api.mjs.
const discordTypesFile = "../../../../kite-web/src/lib/integrations/discord/types.json"

// cacheTestState has a server with a member, two roles and two channels in
// its cache, once with every field set and once with as few as Discord sends.
func cacheTestState(intents gateway.Intents) *state.State {
	s := state.New("Bot token")
	s.AddIntents(intents)
	s.Cabinet.MyselfSet(discord.User{ID: 99, Username: "bot", Bot: true}, false)

	now := discord.NewTimestamp(time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC))
	enabled := true
	members := []discord.Member{
		{User: discord.User{ID: 10, Username: "kite"}, Joined: now},
		{
			User: discord.User{
				ID: 11, Username: "flyer", Discriminator: "0", Avatar: "a1", DisplayName: "Flyer",
				Bot: true, PublicFlags: 64, Flags: 64, Banner: "b1", Accent: 255,
				PrimaryGuild: &discord.PrimaryGuild{IdentityGuildID: 1, IdentityEnabled: option.Bool(&enabled), Tag: "KITE", Badge: "c1"},
			},
			Nick: "Fly", RoleIDs: []discord.RoleID{3}, Avatar: "a2", Banner: "b2", Joined: now,
			BoostedSince: now, CommunicationDisabledUntil: now, Deaf: true, Mute: true, Flags: 1, IsPending: true,
		},
	}
	// The bot, which the cache always has.
	members = append(members, discord.Member{User: discord.User{ID: 99, Username: "bot", Bot: true}, Joined: now})
	for _, m := range members {
		s.Cabinet.MemberSet(1, &m, false)
	}

	roles := []discord.Role{
		// @everyone has the ID of the server.
		{ID: 1, Name: "@everyone", Permissions: discord.PermissionViewChannel},
		{
			ID: 3, Name: "Mod", Permissions: discord.PermissionBanMembers, Position: 1, Color: 255,
			Hoist: true, Mentionable: true, Icon: "i1", UnicodeEmoji: "🪁",
			Tags:   discord.RoleTags{BotID: 4, PremiumSubscriber: true},
			Colors: discord.RoleColors{Primary: 255, Secondary: 65280, Tertiary: discord.NullColor},
			Flags:  discord.RoleInPrompt,
		},
	}
	for _, r := range roles {
		s.Cabinet.RoleSet(1, &r, false)
	}

	channels := []discord.Channel{
		{ID: 5, GuildID: 1, Type: discord.GuildText, Name: "general"},
		{
			ID: 6, GuildID: 1, Type: discord.GuildText, Name: "rules", Position: 2, Topic: "Be nice",
			NSFW: true, ParentID: 7, LastMessageID: 8, UserRateLimit: 5, Flags: 2,
			Overwrites: []discord.Overwrite{{ID: 1, Type: discord.OverwriteRole, Deny: discord.PermissionSendMessages}},
		},
		{ID: 7, GuildID: 1, Type: discord.GuildCategory, Name: "Info"},
		{ID: 9, GuildID: 1, Type: discord.GuildVoice, Name: "Voice", VoiceBitrate: 64000},
		{ID: 12, GuildID: 1, Type: discord.GuildPublicThread, Name: "thread", ParentID: 5},
		// Hidden from the bot.
		{
			ID: 13, GuildID: 1, Type: discord.GuildText, Name: "staff",
			Overwrites: []discord.Overwrite{{ID: 1, Type: discord.OverwriteRole, Deny: discord.PermissionViewChannel}},
		},
	}
	s.Cabinet.GuildSet(&discord.Guild{ID: 1, Name: "Kite"}, false)
	for _, c := range channels {
		s.Cabinet.ChannelSet(&c, false)
	}
	return s
}

func cachedResponse(t *testing.T, p *DiscordProvider, path string) any {
	t.Helper()
	b, ok := p.cachedAPIResponse(provider.DiscordAPIRequest{Method: http.MethodGet, Path: path})
	require.True(t, ok, path)
	var v any
	require.NoError(t, json.Unmarshal(b, &v))
	return v
}

// Code gets the objects of Discord's API, so a cached one has every field
// the spec requires, with values of its type.
func TestCachedAPIResponsesMatchTheSpec(t *testing.T) {
	b, err := os.ReadFile(discordTypesFile)
	require.NoError(t, err)
	var decls map[string]string
	require.NoError(t, json.Unmarshal(b, &decls))
	spec := specTypes{decls: decls}

	// arikawa's JSON of a member, without the fixes, doesn't.
	raw := toJSONMap(discord.Member{User: discord.User{ID: 10}})
	assert.ElementsMatch(t, []string{
		".avatar: missing",
		".banner: missing",
		".nick: missing",
		".premium_since: missing",
		".joined_at: null isn't string",
		".roles: null isn't string[]",
		".user.primary_guild: missing",
		".user.public_flags: missing",
		".user.flags: missing",
	}, spec.check("discord.GuildMemberResponse", raw, ""))

	p := &DiscordProvider{session: cacheTestState(gateway.IntentGuilds | gateway.IntentGuildMembers)}
	tests := []struct {
		path string
		typ  string
	}{
		{"/guilds/1/members/10", "discord.GuildMemberResponse"},
		{"/guilds/1/members/11", "discord.GuildMemberResponse"},
		{"/guilds/1/roles", "discord.GuildRoleResponse[]"},
		{"/guilds/1/roles/1", "discord.GuildRoleResponse"},
		{"/guilds/1/roles/3", "discord.GuildRoleResponse"},
		{"/channels/5", "discord.GuildChannelResponse"},
		{"/channels/6", "discord.GuildChannelResponse"},
		{"/channels/7", "discord.GuildChannelResponse"},
		{"/channels/9", "discord.GuildChannelResponse"},
		{"/guilds/1/channels", "discord.GuildChannelResponse[]"},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			for _, problem := range spec.check(tt.typ, cachedResponse(t, p, tt.path), "") {
				t.Error(problem)
			}
		})
	}
}

func TestCachedAPIResponses(t *testing.T) {
	p := &DiscordProvider{session: cacheTestState(gateway.IntentGuilds | gateway.IntentGuildMembers)}

	member := cachedResponse(t, p, "/guilds/1/members/11").(map[string]any)
	assert.Equal(t, "Fly", member["nick"])
	assert.Equal(t, []any{"3"}, member["roles"])
	assert.Equal(t, "KITE", member["user"].(map[string]any)["primary_guild"].(map[string]any)["tag"])

	role := cachedResponse(t, p, "/guilds/1/roles/3").(map[string]any)
	assert.Equal(t, map[string]any{"primary_color": 255.0, "secondary_color": 65280.0, "tertiary_color": nil}, role["colors"])
	assert.Equal(t, "4", role["permissions"])

	// Threads aren't listed, like by Discord.
	channels := cachedResponse(t, p, "/guilds/1/channels").([]any)
	assert.Len(t, channels, 5)

	// The cache doesn't keep the last message up to date.
	channel := cachedResponse(t, p, "/channels/6").(map[string]any)
	assert.NotContains(t, channel, "last_message_id")
}

func TestCachedAPIResponsesFallBack(t *testing.T) {
	var paths []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.RequestURI())
		w.Write([]byte(`{}`))
	}))
	defer server.Close()
	endpoint := api.Endpoint
	api.Endpoint = server.URL + api.Path + "/"
	defer func() { api.Endpoint = endpoint }()

	// Without the Server Members intent, members aren't kept up to date.
	p := &DiscordProvider{session: cacheTestState(gateway.IntentGuilds)}
	for _, path := range []string{
		"/guilds/1/members/10",
		// Not in the cache.
		"/guilds/1/roles/99",
		"/channels/99",
		// Threads, channels the bot can't see, queries and other endpoints.
		"/channels/12",
		"/channels/13",
		"/guilds/1/channels?foo=1",
		"/guilds/1",
	} {
		_, err := p.APIRequest(context.Background(), provider.DiscordAPIRequest{Method: http.MethodGet, Path: path})
		require.NoError(t, err)
	}
	_, err := p.APIRequest(context.Background(), provider.DiscordAPIRequest{Method: http.MethodPost, Path: "/channels/5", Body: []byte(`{}`)})
	require.NoError(t, err)
	_, err = p.APIRequest(context.Background(), provider.DiscordAPIRequest{Method: http.MethodGet, Path: "/channels/5"})
	require.NoError(t, err)

	assert.Equal(t, []string{
		"/api/v9/guilds/1/members/10",
		"/api/v9/guilds/1/roles/99",
		"/api/v9/channels/99",
		"/api/v9/channels/12",
		"/api/v9/channels/13",
		"/api/v9/guilds/1/channels?foo=1",
		"/api/v9/guilds/1",
		"/api/v9/channels/5",
	}, paths)
}

// specTypes checks JSON values against the TypeScript declarations of
// types.json, which have a field per line like `  nick: string | null;`.
type specTypes struct {
	decls map[string]string
}

var specField = regexp.MustCompile(`^  ([A-Za-z_$][\w$]*|"[^"]+")(\??): (.+);$`)

func (s specTypes) check(typ string, v any, at string) []string {
	var problems []string
	for _, option := range splitTopLevel(typ, " | ") {
		p := s.checkOne(option, v, at)
		if len(p) == 0 {
			return nil
		}
		problems = append(problems, p...)
	}
	if len(problems) == 0 {
		return []string{at + ": doesn't match " + typ}
	}
	return problems
}

func (s specTypes) checkOne(typ string, v any, at string) []string {
	typ = strings.TrimSpace(typ)
	if strings.HasPrefix(typ, "(") && strings.HasSuffix(typ, ")") {
		return s.check(typ[1:len(typ)-1], v, at)
	}
	mismatch := []string{at + ": " + jsonKind(v) + " isn't " + typ}
	switch {
	case typ == "any":
		return nil
	case typ == "null":
		if v == nil {
			return nil
		}
	case typ == "string" || strings.HasPrefix(typ, `"`):
		if _, ok := v.(string); ok {
			return nil
		}
	case typ == "number" || regexp.MustCompile(`^-?\d`).MatchString(typ):
		if _, ok := v.(float64); ok {
			return nil
		}
	case typ == "boolean" || typ == "true" || typ == "false":
		if _, ok := v.(bool); ok {
			return nil
		}
	case strings.HasSuffix(typ, "[]"):
		list, ok := v.([]any)
		if !ok {
			return mismatch
		}
		var problems []string
		for i, item := range list {
			problems = append(problems, s.check(strings.TrimSuffix(typ, "[]"), item, at+"["+strconv.Itoa(i)+"]")...)
		}
		return problems
	case strings.HasPrefix(typ, "Record<"):
		if _, ok := v.(map[string]any); ok {
			return nil
		}
	case strings.HasPrefix(typ, "discord."):
		decl := s.decls[strings.TrimPrefix(typ, "discord.")]
		if alias, ok := strings.CutPrefix(decl, "type "); ok {
			_, def, _ := strings.Cut(alias, " = ")
			return s.check(strings.TrimSuffix(def, ";"), v, at)
		}
		obj, ok := v.(map[string]any)
		if !ok {
			return mismatch
		}
		var problems []string
		for _, line := range strings.Split(decl, "\n") {
			m := specField.FindStringSubmatch(line)
			if m == nil {
				continue
			}
			name := strings.Trim(m[1], `"`)
			value, present := obj[name]
			if !present {
				if m[2] == "" {
					problems = append(problems, at+"."+name+": missing")
				}
				continue
			}
			problems = append(problems, s.check(m[3], value, at+"."+name)...)
		}
		return problems
	default:
		return []string{at + ": can't check " + typ}
	}
	return mismatch
}

func jsonKind(v any) string {
	switch v.(type) {
	case nil:
		return "null"
	case string:
		return "string"
	case float64:
		return "number"
	case bool:
		return "boolean"
	case []any:
		return "list"
	}
	return "object"
}

// splitTopLevel splits a type at sep outside of brackets.
func splitTopLevel(typ, sep string) []string {
	var parts []string
	depth, start := 0, 0
	for i := 0; i < len(typ); i++ {
		switch typ[i] {
		case '(', '{', '[', '<':
			depth++
		case ')', '}', ']', '>':
			depth--
		default:
			if depth == 0 && strings.HasPrefix(typ[i:], sep) {
				parts = append(parts, typ[start:i])
				start = i + len(sep)
			}
		}
	}
	return append(parts, typ[start:])
}
