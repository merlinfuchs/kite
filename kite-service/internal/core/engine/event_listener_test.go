package engine

import (
	"testing"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/ws"
)

// Every guild the bot is in arrives as a raw GUILD_CREATE on connect, so only
// arikawa's derived join and leave events may trigger the bot server listeners.
func TestShouldHandleGuildJoinAndLeave(t *testing.T) {
	tests := []struct {
		name  string
		event ws.Event
		want  bool
	}{
		{"raw guild create", &gateway.GuildCreateEvent{}, false},
		{"raw guild delete", &gateway.GuildDeleteEvent{}, false},
		{"guild ready on connect", &state.GuildReadyEvent{GuildCreateEvent: &gateway.GuildCreateEvent{}}, false},
		{"guild available after outage", &state.GuildAvailableEvent{GuildCreateEvent: &gateway.GuildCreateEvent{}}, false},
		{"guild unavailable", &state.GuildUnavailableEvent{GuildDeleteEvent: &gateway.GuildDeleteEvent{Unavailable: true}}, false},
		{"guild join", &state.GuildJoinEvent{GuildCreateEvent: &gateway.GuildCreateEvent{}}, true},
		{"guild leave", &state.GuildLeaveEvent{GuildDeleteEvent: &gateway.GuildDeleteEvent{}}, true},
	}

	l := &EventListener{}
	for _, tt := range tests {
		if got := l.shouldHandleEvent(tt.event); got != tt.want {
			t.Errorf("%s: got %v, want %v", tt.name, got, tt.want)
		}
	}
}

// A flow that creates an invite must not trigger its own listener.
func TestShouldHandleInviteCreate(t *testing.T) {
	tests := []struct {
		name  string
		event ws.Event
		want  bool
	}{
		{"invite by a user", &gateway.InviteCreateEvent{Inviter: &discord.User{}}, true},
		{"invite by a bot", &gateway.InviteCreateEvent{Inviter: &discord.User{Bot: true}}, false},
		{"invite without inviter", &gateway.InviteCreateEvent{}, true},
	}

	l := &EventListener{}
	for _, tt := range tests {
		if got := l.shouldHandleEvent(tt.event); got != tt.want {
			t.Errorf("%s: got %v, want %v", tt.name, got, tt.want)
		}
	}
}
