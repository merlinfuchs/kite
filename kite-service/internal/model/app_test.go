package model

import (
	"encoding/json"
	"testing"
	"time"
)

func TestAppDiscordStatusUnmarshalLegacy(t *testing.T) {
	var s AppDiscordStatus
	err := json.Unmarshal([]byte(`{"status":"dnd","activity_type":3,"activity_name":"you"}`), &s)
	if err != nil {
		t.Fatal(err)
	}

	if len(s.Statuses) != 1 || s.ActiveID != "default" {
		t.Fatalf("expected one default status, got %+v", s)
	}
	entry := s.Statuses[0]
	if entry.ID != "default" || entry.Status != "dnd" || entry.ActivityType != 3 || entry.ActivityName != "you" {
		t.Fatalf("legacy fields not carried over: %+v", entry)
	}
}

func TestAppDiscordStatusUnmarshalCurrent(t *testing.T) {
	var s AppDiscordStatus
	err := json.Unmarshal([]byte(`{"statuses":[{"id":"a","activity_name":"one"},{"id":"b"}],"active_id":"b","rotate_enabled":true}`), &s)
	if err != nil {
		t.Fatal(err)
	}

	if len(s.Statuses) != 2 || s.ActiveID != "b" || !s.RotateEnabled || s.Statuses[0].ActivityName != "one" {
		t.Fatalf("unexpected result: %+v", s)
	}
	if s.ActiveEntry().ID != "b" {
		t.Fatalf("expected active entry b, got %s", s.ActiveEntry().ID)
	}
}

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
