package guildevent

import "testing"

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
