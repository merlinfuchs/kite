package model

import (
	"testing"
	"time"
)

func TestAppDiscordStatusRotationEntry(t *testing.T) {
	s := AppDiscordStatus{
		RotateEnabled: true,
		Statuses:      []AppDiscordStatusEntry{{ID: "a"}, {ID: "b"}, {ID: "c"}},
	}

	start := time.Unix(60*300, 0)
	for i, want := range []string{"a", "b", "c", "a"} {
		got := s.RotationEntry(start.Add(time.Duration(i) * time.Minute)).ID
		if got != want {
			t.Fatalf("minute %d: expected %s, got %s", i, want, got)
		}
	}
	if got := s.RotationEntry(start.Add(59 * time.Second)).ID; got != "a" {
		t.Fatalf("expected entry to stay for the whole minute, got %s", got)
	}
}
