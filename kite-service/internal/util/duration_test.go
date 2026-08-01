package util

import (
	"testing"
	"time"
)

func TestIntervalOrDefault(t *testing.T) {
	tests := []struct {
		name       string
		configured time.Duration
		fallback   time.Duration
		want       time.Duration
	}{
		{"configured value is used", 30 * time.Second, 5 * time.Second, 30 * time.Second},
		{"zero falls back", 0, 5 * time.Second, 5 * time.Second},
		{"negative falls back", -1 * time.Second, 5 * time.Second, 5 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IntervalOrDefault(tt.configured, tt.fallback); got != tt.want {
				t.Errorf("IntervalOrDefault(%v, %v) = %v, want %v",
					tt.configured, tt.fallback, got, tt.want)
			}
		})
	}
}
