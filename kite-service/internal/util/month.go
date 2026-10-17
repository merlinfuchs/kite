package util

import "time"

// StartAndEndOfMonth returns the first and last moment of the month t is in.
func StartAndEndOfMonth(t time.Time) (time.Time, time.Time) {
	year, month, _ := t.Date()
	start := time.Date(year, month, 1, 0, 0, 0, 0, t.Location())
	end := start.AddDate(0, 1, 0).Add(-time.Nanosecond)
	return start, end
}
