package eval

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/state/store"
	"github.com/diamondburned/arikawa/v3/state/store/defaultstore"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
)

const (
	testGuildID = discord.GuildID(100)
	// A real ID, created 2016-04-30.
	testUserID = discord.UserID(175928847299117063)
)

func testSession(t *testing.T) *state.State {
	t.Helper()

	cabinet := defaultstore.New()
	guild := discord.Guild{
		ID:                 testGuildID,
		Name:               "Kite HQ",
		Icon:               "abc",
		OwnerID:            7,
		ApproximateMembers: 1234,
		NitroBoosters:      3,
	}
	if err := cabinet.GuildSet(&guild, true); err != nil {
		t.Fatal(err)
	}
	roles := []discord.Role{
		{ID: discord.RoleID(testGuildID), Name: "@everyone", Position: 0},
		{ID: 11, Name: "Member", Position: 1},
		{ID: 12, Name: "Admin", Position: 5},
		{ID: 13, Name: "Mod", Position: 3},
	}
	for i := range roles {
		if err := cabinet.RoleSet(testGuildID, &roles[i], true); err != nil {
			t.Fatal(err)
		}
	}

	return &state.State{Cabinet: cabinet}
}

func evalString(t *testing.T, c Context, template string) string {
	t.Helper()

	got, err := EvalTemplateToString(context.Background(), template, c)
	if err != nil {
		t.Fatalf("eval %q: %v", template, err)
	}
	return got
}

func TestGuildPlaceholders(t *testing.T) {
	env := newEventEnv(&gateway.GuildMemberAddEvent{
		GuildID: testGuildID,
		Member:  discord.Member{User: discord.User{ID: testUserID}},
	}, testSession(t))
	c := Context{Env: Env{"guild": env.Guild}}

	got := evalString(t, c, "{{guild.id}}|{{guild.name}}|{{guild.member_count}}|{{guild.boost_count}}|{{guild.owner_id}}")
	if want := "100|Kite HQ|1234|3|7"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if got := evalString(t, c, "{{guild.icon_url}}"); got != "https://cdn.discordapp.com/icons/100/abc.png" {
		t.Errorf("icon_url: got %q", got)
	}
	// Still usable as the ID, like before the other fields existed.
	if got := evalString(t, c, "{{guild}}"); got != "100" {
		t.Errorf("guild: got %q", got)
	}
}

func TestGuildPlaceholdersWithoutCache(t *testing.T) {
	// Neither a missing session nor an uncached server may fail the flow.
	for name, session := range map[string]*state.State{
		"no session": nil,
		"not cached": {Cabinet: defaultstore.New()},
	} {
		t.Run(name, func(t *testing.T) {
			i := &discord.InteractionEvent{
				GuildID: testGuildID,
				Member:  &discord.Member{User: discord.User{ID: testUserID}, RoleIDs: []discord.RoleID{11}},
				Data:    &discord.StringSelectInteraction{},
			}
			env := newInteractionEnv(i, session)
			c := Context{Env: Env{"guild": env.Guild, "user": env.User}}

			got := evalString(t, c, "[{{guild.id}}|{{guild.name}}|{{guild.member_count}}|{{guild.owner_id}}|{{guild.role_count}}|{{user.role_mentions}}]")
			if want := "[100||0||0|<@&11>]"; got != want {
				t.Errorf("got %q, want %q", got, want)
			}

			// Without the roles of the server, what needs them isn't known.
			// Failing is better than claiming an admin isn't one.
			for _, template := range []string{"{{user.is_admin}}", "{{user.top_role.name}}", "{{user.permissions}}"} {
				_, err := EvalTemplateToString(context.Background(), template, c)
				if err == nil || !strings.Contains(err.Error(), "isn't cached") {
					t.Errorf("%s: got error %v", template, err)
				}
			}
		})
	}
}

func TestNoGuildInDirectMessages(t *testing.T) {
	i := &discord.InteractionEvent{
		User: &discord.User{ID: testUserID},
		Data: &discord.StringSelectInteraction{},
	}
	env := newInteractionEnv(i, testSession(t))
	if env.Guild != nil {
		t.Errorf("guild: got %+v, want nil", env.Guild)
	}

	// Member placeholders are empty for a user outside of a server.
	c := Context{Env: Env{"user": env.User}}
	got := evalString(t, c, "[{{user.joined_at}}|{{user.top_role}}|{{user.top_role.mention}}|{{user.role_mentions}}]")
	if want := "[0|||]"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	// So are the server placeholders, while the server itself stays nil.
	dms := map[string]Context{
		"interaction": {Env: Env{"guild": env.Guild}},
		"event": NewContext(Env{"guild": newEventEnv(&gateway.MessageCreateEvent{
			Message: discord.Message{ID: 1, ChannelID: 50, Author: discord.User{ID: testUserID}},
		}, nil).Guild}),
	}
	for name, c := range dms {
		c.Env["origin"] = map[string]any{"guild": c.Env["guild"]}
		got := evalString(t, c, "[{{guild}}|{{guild.id}}|{{guild.name}}|{{guild.member_count}}|{{guild.role_count}}|{{guild.rules_channel.mention}}|{{origin.guild.name}}|{{guild == nil}}|{{guild?.name == nil}}]")
		if want := "[|||0|0|||true|true]"; got != want {
			t.Errorf("%s: got %q, want %q", name, got, want)
		}
	}
}

func TestUserPlaceholders(t *testing.T) {
	joined := time.Date(2024, 5, 1, 12, 0, 0, 0, time.UTC)
	i := &discord.InteractionEvent{
		GuildID: testGuildID,
		Member: &discord.Member{
			User:    discord.User{ID: testUserID, Bot: true},
			Joined:  discord.NewTimestamp(joined),
			RoleIDs: []discord.RoleID{11, 13, 12},
		},
		Data: &discord.StringSelectInteraction{},
	}
	env := newInteractionEnv(i, testSession(t))
	c := Context{Env: Env{"user": env.User}}

	tests := map[string]string{
		"{{user.is_bot}}":           "true",
		"{{user.created_at}}":       "1462015105",
		"{{user.joined_at}}":        "1714564800",
		"{{user.top_role}}":         "<@&12>",
		"{{user.top_role.id}}":      "12",
		"{{user.top_role.name}}":    "Admin",
		"{{user.top_role.mention}}": "<@&12>",
		// Highest role first.
		"{{user.role_mentions}}": "<@&12> <@&13> <@&11>",
		// What account age checks look like.
		"{{user.created_at < now().Unix() - 86400}}": "true",
		"<t:{{user.joined_at}}:R>":                   "<t:1714564800:R>",
	}
	for template, want := range tests {
		if got := evalString(t, c, template); got != want {
			t.Errorf("%s: got %q, want %q", template, got, want)
		}
	}
}

func TestTopRoleWithoutRolesIsEveryone(t *testing.T) {
	env := newEventEnv(&gateway.GuildMemberAddEvent{
		GuildID: testGuildID,
		Member:  discord.Member{User: discord.User{ID: testUserID}},
	}, testSession(t))
	c := Context{Env: Env{"user": env.User}}

	if got := evalString(t, c, "{{user.top_role.name}}|{{user.role_mentions}}|"); got != "@everyone||" {
		t.Errorf("got %q", got)
	}
}

func TestCommandArgMemberUsesGuildRoles(t *testing.T) {
	data := &discord.CommandInteraction{
		Options: []discord.CommandInteractionOption{
			{Name: "target", Type: discord.UserOptionType, Value: []byte(`"5"`)},
		},
	}
	data.Resolved.Users = map[discord.UserID]discord.User{5: {ID: 5}}
	data.Resolved.Members = map[discord.UserID]discord.Member{5: {RoleIDs: []discord.RoleID{11, 13}}}

	i := &discord.InteractionEvent{
		GuildID: testGuildID,
		Member:  &discord.Member{User: discord.User{ID: testUserID}},
		Data:    data,
	}
	c := NewContext(Env{})
	env := newInteractionEnv(i, testSession(t))
	c.Env["arg"] = func(name string) any { return env.Command.Args[name] }

	if got := evalString(t, c, "{{arg('target').top_role.name}}"); got != "Mod" {
		t.Errorf("got %q", got)
	}
}

func testChannelSession(t *testing.T) *state.State {
	t.Helper()

	session := testSession(t)
	channels := []discord.Channel{
		{ID: 20, GuildID: testGuildID, Name: "Tickets", Type: discord.GuildCategory},
		{ID: 21, GuildID: testGuildID, Name: "support", Type: discord.GuildText, ParentID: 20},
		{ID: 22, GuildID: testGuildID, Name: "ticket-1", Type: discord.GuildPublicThread, ParentID: 21},
		{ID: 23, GuildID: testGuildID, Name: "lobby", Type: discord.GuildVoice},
	}
	for i := range channels {
		if err := session.Cabinet.ChannelSet(&channels[i], true); err != nil {
			t.Fatal(err)
		}
	}
	return session
}

func TestChannelPlaceholders(t *testing.T) {
	session := testChannelSession(t)
	const template = "{{channel}}|{{channel.id}}|{{channel.name}}|{{channel.mention}}|{{channel.type}}|{{channel.category_id}}|{{channel.category_name}}"

	tests := map[discord.ChannelID]string{
		21: "21|21|support|<#21>|text|20|Tickets",
		// A thread is in the category of its channel.
		22: "22|22|ticket-1|<#22>|thread|20|Tickets",
		23: "23|23|lobby|<#23>|voice||",
		// Not cached, so only the ID is known.
		99: "99|99||<#99>|||",
	}
	for id, want := range tests {
		env := newEventEnv(&gateway.MessageDeleteEvent{ID: 1, ChannelID: id, GuildID: testGuildID}, session)
		c := Context{Env: Env{"channel": env.Channel}}
		if got := evalString(t, c, template); got != want {
			t.Errorf("channel %d: got %q, want %q", id, got, want)
		}
	}
}

func TestChannelPlaceholdersFromInteraction(t *testing.T) {
	// Discord sends the channel along with an interaction, which is used when
	// the channel isn't cached, like in direct messages.
	i := &discord.InteractionEvent{
		ChannelID: 50,
		Channel:   &discord.Channel{ID: 50, Type: discord.DirectMessage},
		User:      &discord.User{ID: testUserID},
		Data:      &discord.StringSelectInteraction{},
	}
	for name, session := range map[string]*state.State{"no session": nil, "not cached": testChannelSession(t)} {
		env := newInteractionEnv(i, session)
		c := Context{Env: Env{"channel": env.Channel}}
		if got := evalString(t, c, "{{channel.id}}|{{channel.type}}|{{channel.category_id}}|"); got != "50|dm||" {
			t.Errorf("%s: got %q", name, got)
		}
	}
}

func TestCommandArgChannelUsesCache(t *testing.T) {
	// Channels of command arguments are partial, the cache knows the category.
	data := &discord.CommandInteraction{
		Options: []discord.CommandInteractionOption{
			{Name: "where", Type: discord.ChannelOptionType, Value: []byte(`"21"`)},
		},
	}
	data.Resolved.Channels = map[discord.ChannelID]discord.Channel{21: {ID: 21, Name: "support"}}

	i := &discord.InteractionEvent{
		GuildID: testGuildID,
		Member:  &discord.Member{User: discord.User{ID: testUserID}},
		Data:    data,
	}
	env := newInteractionEnv(i, testChannelSession(t))
	c := NewContext(Env{})
	c.Env["arg"] = func(name string) any { return env.Command.Args[name] }

	if got := evalString(t, c, "{{arg('where').name}}|{{arg('where').category_name}}"); got != "support|Tickets" {
		t.Errorf("got %q", got)
	}
}

func TestMemberPlaceholders(t *testing.T) {
	session := testSession(t)
	roles := []discord.Role{
		{ID: discord.RoleID(testGuildID), Name: "@everyone", Permissions: discord.PermissionSendMessages},
		{ID: 11, Name: "Member", Position: 1},
		{ID: 12, Name: "Admin", Position: 5, Permissions: discord.PermissionAdministrator},
		{ID: 13, Name: "Mod", Position: 3, Color: 0x5865f2, Permissions: discord.PermissionBanMembers | discord.PermissionKickMembers},
	}
	for i := range roles {
		if err := session.Cabinet.RoleSet(testGuildID, &roles[i], true); err != nil {
			t.Fatal(err)
		}
	}

	member := func(m discord.Member) Context {
		i := &discord.InteractionEvent{GuildID: testGuildID, Member: &m, Data: &discord.StringSelectInteraction{}}
		return Context{Env: Env{"user": newInteractionEnv(i, session).User}}
	}
	boosted := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	until := time.Now().Add(time.Hour)

	mod := member(discord.Member{
		User:                       discord.User{ID: testUserID},
		RoleIDs:                    []discord.RoleID{11, 13},
		BoostedSince:               discord.NewTimestamp(boosted),
		CommunicationDisabledUntil: discord.NewTimestamp(until),
	})
	tests := map[string]string{
		"{{user.role_count}}":     "2",
		"{{user.role_names}}":     "Mod, Member",
		"{{user.color}}":          "#5865f2",
		"{{user.is_booster}}":     "true",
		"{{user.boosting_since}}": "1735689600",
		"{{user.is_timed_out}}":   "true",
		"{{user.is_owner}}":       "false",
		"{{user.is_admin}}":       "false",
		// From the Mod role and from @everyone.
		`{{"ban_members" in user.permissions}}`:   "true",
		`{{"send_messages" in user.permissions}}`: "true",
		`{{"manage_guild" in user.permissions}}`:  "false",
		"{{len(user.permissions)}}":               "3",
		"{{user.timeout_until > now().Unix()}}":   "true",
	}
	for template, want := range tests {
		if got := evalString(t, mod, template); got != want {
			t.Errorf("%s: got %q, want %q", template, got, want)
		}
	}

	// A timeout that has ended, and no boost.
	plain := member(discord.Member{
		User:                       discord.User{ID: testUserID},
		CommunicationDisabledUntil: discord.NewTimestamp(time.Now().Add(-time.Hour)),
	})
	if got := evalString(t, plain, "{{user.is_timed_out}}|{{user.timeout_until}}|{{user.is_booster}}|{{user.boosting_since}}|{{user.color}}|{{user.role_names}}|{{user.role_count}}"); got != "false|0|false|0|||0" {
		t.Errorf("plain member: got %q", got)
	}

	// Administrators and the owner have every permission.
	admin := member(discord.Member{User: discord.User{ID: testUserID}, RoleIDs: []discord.RoleID{12}})
	owner := member(discord.Member{User: discord.User{ID: 7}})
	for name, c := range map[string]Context{"admin": admin, "owner": owner} {
		want := "true|true|" + map[string]string{"admin": "false", "owner": "true"}[name]
		if got := evalString(t, c, `{{user.is_admin}}|{{"manage_guild" in user.permissions}}|{{user.is_owner}}`); got != want {
			t.Errorf("%s: got %q, want %q", name, got, want)
		}
	}

	// Outside of a server nothing is known, and nothing fails.
	i := &discord.InteractionEvent{User: &discord.User{ID: testUserID}, Data: &discord.StringSelectInteraction{}}
	dm := Context{Env: Env{"user": newInteractionEnv(i, session).User}}
	if got := evalString(t, dm, `{{user.is_admin}}|{{user.is_owner}}|{{user.is_booster}}|{{user.is_timed_out}}|{{user.role_count}}|{{"ban_members" in user.permissions}}|{{user.color}}`); got != "false|false|false|false|0|false|" {
		t.Errorf("dm: got %q", got)
	}
}

func TestServerInfoPlaceholders(t *testing.T) {
	session := testChannelSession(t)
	guild, _ := session.Cabinet.Guild(testGuildID)
	guild.NitroBoost = discord.NitroLevel2
	guild.Banner = "bnr"
	guild.Description = "A place for kites"
	guild.VanityURLCode = "kite"
	guild.RulesChannelID = 21
	if err := session.Cabinet.GuildSet(guild, true); err != nil {
		t.Fatal(err)
	}
	if err := session.Cabinet.EmojiSet(testGuildID, []discord.Emoji{{ID: 1}, {ID: 2}}, true); err != nil {
		t.Fatal(err)
	}

	env := newEventEnv(&gateway.GuildMemberAddEvent{
		GuildID: testGuildID,
		Member:  discord.Member{User: discord.User{ID: testUserID}},
	}, session)
	c := Context{Env: Env{"guild": env.Guild}}

	tests := map[string]string{
		"{{guild.boost_level}}": "2",
		// Server IDs are this small only in tests.
		"{{guild.created_at}}":  "1420070400",
		"{{guild.banner_url}}":  "https://cdn.discordapp.com/banners/100/bnr.png",
		"{{guild.description}}": "A place for kites",
		"{{guild.vanity_url}}":  "https://discord.gg/kite",
		// Without @everyone.
		"{{guild.role_count}}": "3",
		// Without the category and the thread.
		"{{guild.channel_count}}":       "2",
		"{{guild.emoji_count}}":         "2",
		"{{guild.rules_channel}}":       "<#21>",
		"{{guild.rules_channel.id}}":    "21",
		"{{guild.rules_channel.name}}":  "support",
		"[{{guild.system_channel}}]":    "[]",
		"[{{guild.system_channel.id}}]": "[]",
	}
	for template, want := range tests {
		if got := evalString(t, c, template); got != want {
			t.Errorf("%s: got %q, want %q", template, got, want)
		}
	}

	// A server that isn't cached has none of it, and nothing fails.
	uncached := newEventEnv(&gateway.GuildMemberAddEvent{
		GuildID: 555,
		Member:  discord.Member{User: discord.User{ID: testUserID}},
	}, session)
	c = Context{Env: Env{"guild": uncached.Guild}}
	if got := evalString(t, c, "[{{guild.boost_level}}|{{guild.vanity_url}}|{{guild.role_count}}|{{guild.rules_channel}}|{{guild.rules_channel.mention}}]"); got != "[0||0||]" {
		t.Errorf("uncached: got %q", got)
	}
}

func TestMemberResultNeedsServer(t *testing.T) {
	// A member from a block result or variable doesn't know its server, so
	// what needs its roles fails instead of claiming an admin isn't one.
	member := discord.Member{
		User:    discord.User{ID: testUserID},
		Joined:  discord.NewTimestamp(time.Date(2024, 5, 1, 12, 0, 0, 0, time.UTC)),
		RoleIDs: []discord.RoleID{11, 12},
	}
	c := NewContext(Env{"result": func(string) any {
		return NewThingEnv(thing.NewDiscordMember(member))
	}})

	if got := evalString(t, c, "{{result('a').joined_at}}|{{result('a').role_count}}|{{result('a').role_mentions}}"); got != "1714564800|2|<@&11> <@&12>" {
		t.Errorf("got %q", got)
	}
	for _, field := range []string{"is_admin", "is_owner", "permissions", "top_role", "role_names", "color"} {
		_, err := EvalTemplateToString(context.Background(), "{{result('a')."+field+"}}", c)
		if err == nil || !strings.Contains(err.Error(), field+" isn't known for a member from a block result or variable") {
			t.Errorf("%s: got error %v", field, err)
		}
	}
}

func TestResumedMemberKeepsServer(t *testing.T) {
	session := testSession(t)
	admin := discord.Role{ID: 12, Name: "Admin", Position: 5, Permissions: discord.PermissionAdministrator}
	if err := session.Cabinet.RoleSet(testGuildID, &admin, true); err != nil {
		t.Fatal(err)
	}
	i := &discord.InteractionEvent{
		GuildID: testGuildID,
		Member:  &discord.Member{User: discord.User{ID: testUserID}, RoleIDs: []discord.RoleID{12}},
		Data:    &discord.StringSelectInteraction{},
	}
	origin := Context{Env: Env{"user": newInteractionEnv(i, session).User}}
	c := NewContext(Env{})
	c.SetResumeContext([]Context{origin})

	if got := evalString(t, c, "{{origin.user.is_admin}}|{{origin.user.top_role.name}}"); got != "true|Admin" {
		t.Errorf("got %q", got)
	}
}

// countingRoles and countingChannels count how often all roles or channels of
// a server are read.
type countingRoles struct {
	store.RoleStore
	calls int
}

func (s *countingRoles) Roles(id discord.GuildID) ([]discord.Role, error) {
	s.calls++
	return s.RoleStore.Roles(id)
}

type countingChannels struct {
	store.ChannelStore
	calls int
}

func (s *countingChannels) Channels(id discord.GuildID) ([]discord.Channel, error) {
	s.calls++
	return s.ChannelStore.Channels(id)
}

func TestServerRolesAndChannelsAreReadWhenNeeded(t *testing.T) {
	session := testChannelSession(t)
	roles := &countingRoles{RoleStore: session.Cabinet.RoleStore}
	channels := &countingChannels{ChannelStore: session.Cabinet.ChannelStore}
	session.Cabinet.RoleStore = roles
	session.Cabinet.ChannelStore = channels

	env := newEventEnv(&gateway.MessageCreateEvent{
		Message: discord.Message{ID: 1, ChannelID: 21, GuildID: testGuildID, Author: discord.User{ID: testUserID}},
		Member:  &discord.Member{User: discord.User{ID: testUserID}, RoleIDs: []discord.RoleID{11}},
	}, session)
	c := Context{Env: Env{"guild": env.Guild, "user": env.User, "channel": env.Channel}}

	evalString(t, c, "{{guild.name}} {{user.mention}} {{user.role_count}} {{channel.name}}")
	if roles.calls != 0 || channels.calls != 0 {
		t.Fatalf("read roles %d and channels %d times without needing them", roles.calls, channels.calls)
	}

	// Once for the server and once for the member, however often they're used.
	for i := 0; i < 2; i++ {
		if got := evalString(t, c, "{{guild.channel_count}}|{{guild.role_count}}|{{user.top_role.name}}|{{user.role_names}}"); got != "2|3|Member|Member" {
			t.Errorf("got %q", got)
		}
	}
	if roles.calls != 2 || channels.calls != 1 {
		t.Errorf("read roles %d and channels %d times", roles.calls, channels.calls)
	}
}

func TestPermissionNames(t *testing.T) {
	// Every permission arikawa knows needs a name. It has none of the newer
	// ones, which are checked below.
	known := []discord.Permissions{
		discord.PermissionCreateInstantInvite, discord.PermissionKickMembers, discord.PermissionBanMembers,
		discord.PermissionAdministrator, discord.PermissionManageChannels, discord.PermissionManageGuild,
		discord.PermissionAddReactions, discord.PermissionViewAuditLog, discord.PermissionPrioritySpeaker,
		discord.PermissionStream, discord.PermissionViewChannel, discord.PermissionSendMessages,
		discord.PermissionSendTTSMessages, discord.PermissionManageMessages, discord.PermissionEmbedLinks,
		discord.PermissionAttachFiles, discord.PermissionReadMessageHistory, discord.PermissionMentionEveryone,
		discord.PermissionUseExternalEmojis, discord.PermissionViewGuildInsights, discord.PermissionConnect,
		discord.PermissionSpeak, discord.PermissionMuteMembers, discord.PermissionDeafenMembers,
		discord.PermissionMoveMembers, discord.PermissionUseVAD, discord.PermissionChangeNickname,
		discord.PermissionManageNicknames, discord.PermissionManageRoles, discord.PermissionManageWebhooks,
		discord.PermissionManageEmojisAndStickers, discord.PermissionUseSlashCommands, discord.PermissionRequestToSpeak,
		discord.PermissionManageEvents, discord.PermissionManageThreads, discord.PermissionCreatePublicThreads,
		discord.PermissionCreatePrivateThreads, discord.PermissionUseExternalStickers, discord.PermissionSendMessagesInThreads,
		discord.PermissionStartEmbeddedActivities, discord.PermissionModerateMembers, discord.PermissionViewCreatorMonetizationAnalytics,
		discord.PermissionUseSoundboard, discord.PermissionUseExternalSounds, discord.PermissionSendVoiceMessages,
		discord.PermissionAll,
	}
	for _, permissions := range known {
		for bit := 0; bit < 64; bit++ {
			if permissions&(1<<bit) != 0 && permissionNameByBit[bit] == "" {
				t.Errorf("bit %d has no name", bit)
			}
		}
	}

	got := strings.Join(permissionNames(1<<48|1<<51|1<<52), ",")
	if want := "set_voice_channel_status,pin_messages,bypass_slowmode"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
