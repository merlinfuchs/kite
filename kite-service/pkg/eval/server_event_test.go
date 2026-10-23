package eval

import (
	"testing"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/ws"
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
}
