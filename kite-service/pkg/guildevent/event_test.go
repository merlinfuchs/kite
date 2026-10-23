package guildevent

import (
	"testing"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
)

func TestFromMessageDelete(t *testing.T) {
	msg := &discord.Message{ID: 7, ChannelID: 5, GuildID: 2, Content: "hello"}

	got := FromMessageDelete(msg, &gateway.MessageDeleteEvent{ID: 7, ChannelID: 5, GuildID: 2})
	if got == nil || got.Message != msg || got.ID != 7 {
		t.Errorf("got %+v, want the event with the message", got)
	}

	// Message Delete listeners only run for servers.
	if got := FromMessageDelete(msg, &gateway.MessageDeleteEvent{ID: 7, ChannelID: 5}); got != nil {
		t.Errorf("got %+v for a DM, want nil", got)
	}
}

func TestNewKnowsEveryEvent(t *testing.T) {
	for _, e := range []Event{&MessageDelete{}} {
		got := New(e.EventType())
		if got == nil || got.EventType() != e.EventType() {
			t.Errorf("New(%q) = %T", e.EventType(), got)
		}
	}
	if New("GUILD_BAN_ADD") != nil {
		t.Error("New returned an event for a Discord event type")
	}
}
