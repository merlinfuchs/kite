package engine

import (
	"testing"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/ws"
)

type testForumPostEvent struct {
	ws.Event
}

func (testForumPostEvent) EventType() ws.EventType {
	return "FORUM_POST_CREATE"
}

func (testForumPostEvent) IsForumPostEvent() bool {
	return true
}

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
		if got := l.shouldHandleEvent(tt.event, 0); got != tt.want {
			t.Errorf("%s: got %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestShouldHandleForumPostEvents(t *testing.T) {
	l := &EventListener{}
	if !l.shouldHandleEvent(testForumPostEvent{}, 0) {
		t.Error("forum post event was rejected")
	}
}

// A flow that adds or removes a reaction must not trigger its own reaction
// listeners.
func TestShouldHandleReactionsIgnoresOwnApp(t *testing.T) {
	const botID = discord.UserID(1)
	const userID = discord.UserID(2)

	tests := []struct {
		name  string
		event ws.Event
		want  bool
	}{
		{"reaction add by user", &gateway.MessageReactionAddEvent{UserID: userID}, true},
		{"reaction add by app", &gateway.MessageReactionAddEvent{UserID: botID}, false},
		{"reaction remove by user", &gateway.MessageReactionRemoveEvent{UserID: userID}, true},
		{"reaction remove by app", &gateway.MessageReactionRemoveEvent{UserID: botID}, false},
	}

	l := &EventListener{}
	for _, tt := range tests {
		if got := l.shouldHandleEvent(tt.event, botID); got != tt.want {
			t.Errorf("%s: got %v, want %v", tt.name, got, tt.want)
		}
	}
}
