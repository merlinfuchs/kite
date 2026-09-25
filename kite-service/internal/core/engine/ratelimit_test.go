package engine

import (
	"testing"
	"time"
)

func TestBlockRateLimiter(t *testing.T) {
	l := NewBlockRateLimiter()
	limit := BlockRateLimit{Key: "test", Every: time.Hour, Burst: 2}

	if !l.Allow("a", limit) || !l.Allow("a", limit) {
		t.Fatal("expected burst to be allowed")
	}
	if l.Allow("a", limit) {
		t.Fatal("expected run after burst to be limited")
	}
	if !l.Allow("b", limit) {
		t.Fatal("expected other apps to have their own limit")
	}

	l.Sweep()
	if len(l.limiters) != 2 {
		t.Fatalf("expected used limiters to be kept, got %d", len(l.limiters))
	}
}
