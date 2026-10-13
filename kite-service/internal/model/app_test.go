package model

import (
	"gopkg.in/guregu/null.v4"
	"os"
	"strings"
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

// The migration that rewrote the old reasons has to use the same text, or the
// web app won't recognize those apps.
func TestDisabledReasonsMatchMigration(t *testing.T) {
	raw, err := os.ReadFile("../db/postgres/migrations/057_rewrite_gateway_disabled_reasons.up.sql")
	if err != nil {
		t.Fatal(err)
	}

	for _, reason := range []string{AppDisabledReasonInvalidToken, AppDisabledReasonDisallowedIntents} {
		if !strings.Contains(string(raw), "'"+reason+"'") {
			t.Errorf("migration doesn't set %q", reason)
		}
	}
}
