package util

import "time"

// IntervalOrDefault guards against a zero or negative configured interval,
// which would panic time.NewTicker.
func IntervalOrDefault(configured, fallback time.Duration) time.Duration {
	if configured <= 0 {
		return fallback
	}
	return configured
}
