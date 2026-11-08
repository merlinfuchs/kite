package eval

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestParseDuration(t *testing.T) {
	tests := map[string]any{
		// Go durations, which the builtin accepted.
		"1h30m":     90 * time.Minute,
		"500ms":     500 * time.Millisecond,
		"1µs":       time.Microsecond,
		"-2h":       -2 * time.Hour,
		"1.5h":      90 * time.Minute,
		"0":         time.Duration(0),
		" 2h45m30s": 2*time.Hour + 45*time.Minute + 30*time.Second,
		// What members type.
		"1d":                    24 * time.Hour,
		"1d12h":                 36 * time.Hour,
		"2h 30m":                150 * time.Minute,
		"2H":                    2 * time.Hour,
		"1M":                    time.Minute,
		"2 hours":               2 * time.Hour,
		"10 min":                10 * time.Minute,
		"1 day 2 hours":         26 * time.Hour,
		"1 hour and 30 minutes": 90 * time.Minute,
		"1 day, 2 hours":        26 * time.Hour,
		"1w":                    7 * 24 * time.Hour,
		"1.5 days":              36 * time.Hour,
		// Anything else.
		"":                       nil,
		"abc":                    nil,
		"30":                     nil,
		"h":                      nil,
		"1x":                     nil,
		"2 hours ago":            nil,
		"1 month":                nil,
		"-1 day":                 nil,
		"and":                    nil,
		"99999999999999999999d":  nil,
		strings.Repeat("1d", 51): nil,
	}
	for text, want := range tests {
		d, ok := parseDuration(text)
		var got any
		if ok {
			got = d
		}
		if got != want {
			t.Errorf("parseDuration(%q) = %v, want %v", text, got, want)
		}
	}
}

// duration keeps working like the builtin for Go durations, and is nil for
// text that isn't a duration.
func TestDurationInExpression(t *testing.T) {
	c := NewContext(Env{"time": "1d 12h", "chat": "hello"})
	for _, expression := range []string{
		`duration("2h").Seconds() == 7200`,
		`duration(time).Seconds() == 129600`,
		`now().Add(duration("1h")).After(now())`,
		`duration("1h") > duration("30m")`,
		`duration("1h") + duration("30m") == duration("90m")`,
		`duration(chat) == nil && duration(missing) == nil`,
	} {
		got, err := Eval(context.Background(), expression, c)
		if err != nil {
			t.Errorf("%s: %v", expression, err)
			continue
		}
		if got.Value != true {
			t.Errorf("%s = %v", expression, got.Value)
		}
	}
}
