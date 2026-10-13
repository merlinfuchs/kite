package model

import (
	"gopkg.in/guregu/null.v4"
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

func TestHasInvalidToken(t *testing.T) {
	tests := []struct {
		enabled bool
		reason  string
		want    bool
	}{
		{false, AppDisabledReasonInvalidToken, true},
		{false, "Failed to connect to gateway: websocket closed, reason: 4004 Authentication failed.", true},
		{false, "No credits remaining", false},
		{true, AppDisabledReasonInvalidToken, false},
	}

	for _, test := range tests {
		app := &App{Enabled: test.enabled, DisabledReason: null.StringFrom(test.reason)}
		if got := HasInvalidToken(app); got != test.want {
			t.Errorf("%v %q: got %v, want %v", test.enabled, test.reason, got, test.want)
		}
	}
}
