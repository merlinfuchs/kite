package eval

import (
	"strconv"
	"testing"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/ws"
	"github.com/kitecloud/kite/kite-service/pkg/guildevent"
)

// eventContext is like NewContextFromEvent, without the app, which needs a
// connected session.
func eventContext(event ws.Event, session *state.State) Context {
	env := NewEventEnv(event, session)
	return Context{Env: Env{
		"event":   env,
		"user":    env.User,
		"member":  env.Member,
		"channel": env.Channel,
		"guild":   env.Guild,
		"role":    env.Role,
	}}
}

func TestBanPlaceholders(t *testing.T) {
	user := discord.User{ID: testUserID, Username: "alice"}
	for _, e := range []ws.Event{
		&gateway.GuildBanAddEvent{GuildID: testGuildID, User: user},
		&gateway.GuildBanRemoveEvent{GuildID: testGuildID, User: user},
	} {
		got := evalString(t, eventContext(e, testSession(t)), "{{user.id}}|{{user.username}}|{{guild.name}}")
		if want := "175928847299117063|alice|Kite HQ"; got != want {
			t.Errorf("%T: got %q, want %q", e, got, want)
		}
	}
}

func TestMemberRolePlaceholders(t *testing.T) {
	member := discord.Member{User: discord.User{ID: testUserID, Username: "alice"}, RoleIDs: []discord.RoleID{11, 13}}

	added := eventContext(&guildevent.MemberRoleAdd{MemberRole: guildevent.MemberRole{GuildID: testGuildID, Member: member, RoleID: 13}}, testSession(t))
	got := evalString(t, added, "{{role}}|{{role.id}}|{{role.name}}|{{role.mention}}|{{user.username}}|{{user.role_count}}")
	if want := "<@&13>|13|Mod|<@&13>|alice|2"; got != want {
		t.Errorf("added: got %q, want %q", got, want)
	}

	// A role that isn't cached, like one that was just deleted.
	removed := eventContext(&guildevent.MemberRoleRemove{MemberRole: guildevent.MemberRole{GuildID: testGuildID, Member: member, RoleID: 99}}, testSession(t))
	got = evalString(t, removed, "{{role.id}}|{{role.name}}|{{role.mention}}")
	if want := "99||<@&99>"; got != want {
		t.Errorf("removed: got %q, want %q", got, want)
	}
}

func TestMemberTimeoutPlaceholders(t *testing.T) {
	until := time.Now().Add(time.Hour)
	added := eventContext(&guildevent.MemberTimeoutAdd{MemberTimeout: guildevent.MemberTimeout{
		GuildID: testGuildID,
		Member:  discord.Member{User: discord.User{ID: testUserID}, CommunicationDisabledUntil: discord.NewTimestamp(until)},
	}}, testSession(t))
	if got := evalString(t, added, "{{user.is_timed_out}}|{{user.timeout_until}}|{{guild.name}}"); got != "true|"+strconv.FormatInt(until.Unix(), 10)+"|Kite HQ" {
		t.Errorf("added: got %q", got)
	}

	removed := eventContext(&guildevent.MemberTimeoutRemove{MemberTimeout: guildevent.MemberTimeout{
		GuildID: testGuildID,
		Member:  discord.Member{User: discord.User{ID: testUserID}},
	}}, testSession(t))
	if got := evalString(t, removed, "{{user.is_timed_out}}|{{user.timeout_until}}"); got != "false|0" {
		t.Errorf("removed: got %q", got)
	}
}

func TestChannelEventPlaceholders(t *testing.T) {
	session := testChannelSession(t)
	const template = "{{channel}}|{{channel.name}}|{{channel.mention}}|{{channel.type}}|{{channel.category_name}}|{{guild.name}}"

	created := eventContext(&gateway.ChannelCreateEvent{Channel: discord.Channel{ID: 21, GuildID: testGuildID, Name: "support", ParentID: 20}}, session)
	if got, want := evalString(t, created, template), "21|support|<#21>|text|Tickets|Kite HQ"; got != want {
		t.Errorf("created: got %q, want %q", got, want)
	}

	// A deleted channel isn't cached anymore, so it comes from the event.
	deleted := eventContext(&gateway.ChannelDeleteEvent{Channel: discord.Channel{ID: 98, GuildID: testGuildID, Name: "old-chat", Type: discord.GuildText, ParentID: 20}}, session)
	if got, want := evalString(t, deleted, template), "98|old-chat|<#98>|text|Tickets|Kite HQ"; got != want {
		t.Errorf("deleted: got %q, want %q", got, want)
	}

	updated := eventContext(&guildevent.ChannelUpdate{
		Channel: discord.Channel{ID: 21, GuildID: testGuildID, Name: "support", ParentID: 20},
		Before:  &discord.Channel{ID: 21, GuildID: testGuildID, Name: "help", ParentID: 20, UserRateLimit: 5},
		Changes: []string{"name", "slowmode"},
	}, session)
	if got, want := evalString(t, updated, "{{event.before.name}}|{{event.before.type}}|{{event.before.category_name}}|{{channel.name}}|{{event.changes}}"), "help|text|Tickets|support|name, slowmode"; got != want {
		t.Errorf("updated: got %q, want %q", got, want)
	}

	// Without the channel from before, only its ID is known.
	unknown := eventContext(&guildevent.ChannelUpdate{Channel: discord.Channel{ID: 21, GuildID: testGuildID, Name: "support"}}, session)
	if got, want := evalString(t, unknown, "{{event.before.id}}|{{event.before.name}}|{{event.changes}}"), "21||"; got != want {
		t.Errorf("unknown: got %q, want %q", got, want)
	}
}
