package schedule

import (
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	for _, expr := range []string{"*/5 * * * *", "30 */5 * * * *", "@hourly", "@every 10s"} {
		if _, err := Parse(expr); err != nil {
			t.Errorf("Parse(%q): %v", expr, err)
		}
	}

	for _, expr := range []string{"", "* * *", "61 * * * *", "not a cron"} {
		if _, err := Parse(expr); err == nil {
			t.Errorf("Parse(%q): expected error", expr)
		}
	}
}

func TestMinGap(t *testing.T) {
	from := time.Date(2026, 9, 25, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		expr string
		want time.Duration
	}{
		{"*/5 * * * *", 5 * time.Minute},
		{"* * * * * *", time.Second},
		{"0 * * * *", time.Hour},
		// Runs twice an hour but the two runs are a minute apart.
		{"0,1 * * * *", time.Minute},
	}

	for _, tt := range tests {
		s, err := Parse(tt.expr)
		if err != nil {
			t.Fatalf("Parse(%q): %v", tt.expr, err)
		}
		if got := s.MinGap(from); got != tt.want {
			t.Errorf("MinGap(%q) = %s, want %s", tt.expr, got, tt.want)
		}
	}
}
