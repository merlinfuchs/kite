package eval

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/expr-lang/expr"
)

// durationFunc is duration(text) in expressions. It replaces expr's builtin,
// which only knows Go durations like "1h30m", with one that also reads what
// members type, like "1d 12h" or "2 hours". The text often comes from users,
// so anything else is nil instead of failing the flow.
var durationFunc = expr.Function("duration", func(params ...any) (any, error) {
	// Undefined variables arrive as nil.
	if params[0] == nil {
		return nil, nil
	}
	if d, ok := parseDuration(fmt.Sprint(params[0])); ok {
		return d, nil
	}
	return nil, nil
}, new(func(any) any))

// maxDurationLength bounds the text duration parses.
const maxDurationLength = 100

var durationPartRe = regexp.MustCompile(`(\d+(?:\.\d+)?)\s*([a-z]+)`)

// durationUnits are the units of what people type. Shorter ones, like ms, are
// only part of Go durations.
var durationUnits = map[string]time.Duration{
	"s": time.Second, "sec": time.Second, "secs": time.Second, "second": time.Second, "seconds": time.Second,
	"m": time.Minute, "min": time.Minute, "mins": time.Minute, "minute": time.Minute, "minutes": time.Minute,
	"h": time.Hour, "hr": time.Hour, "hrs": time.Hour, "hour": time.Hour, "hours": time.Hour,
	"d": 24 * time.Hour, "day": 24 * time.Hour, "days": 24 * time.Hour,
	"w": 7 * 24 * time.Hour, "wk": 7 * 24 * time.Hour, "wks": 7 * 24 * time.Hour, "week": 7 * 24 * time.Hour, "weeks": 7 * 24 * time.Hour,
}

// parseDuration reads a Go duration, which keeps what the builtin accepted
// working the same, or else parts like "1d", "12 h" or "2 hours" separated by
// spaces, commas or "and". Months and years have no fixed length, so they
// aren't units, and "m" is minutes in any case.
func parseDuration(text string) (time.Duration, bool) {
	text = strings.TrimSpace(text)
	if len(text) > maxDurationLength {
		return 0, false
	}
	if d, err := time.ParseDuration(text); err == nil {
		return d, true
	}

	text = strings.ToLower(text)
	parts := durationPartRe.FindAllStringSubmatch(text, -1)
	rest := strings.Fields(strings.ReplaceAll(durationPartRe.ReplaceAllString(text, " "), ",", " "))
	for _, word := range rest {
		if word != "and" {
			return 0, false
		}
	}
	if len(parts) == 0 {
		return 0, false
	}

	var total float64
	for _, part := range parts {
		unit, ok := durationUnits[part[2]]
		if !ok {
			return 0, false
		}
		n, err := strconv.ParseFloat(part[1], 64)
		if err != nil {
			return 0, false
		}
		total += n * float64(unit)
	}
	if total >= math.MaxInt64 {
		return 0, false
	}
	return time.Duration(total), true
}
