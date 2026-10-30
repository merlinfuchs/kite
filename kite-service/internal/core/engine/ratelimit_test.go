package engine

import (
	"testing"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
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

func TestServerProfileRateLimit(t *testing.T) {
	l := NewBlockRateLimiter()

	for i := 0; i < 5; i++ {
		if !l.Allow("a", serverProfileRateLimit(1)) {
			t.Fatalf("expected change %d to be allowed", i+1)
		}
	}
	if l.Allow("a", serverProfileRateLimit(1)) {
		t.Fatal("expected sixth change in a minute to be limited")
	}
	if !l.Allow("a", serverProfileRateLimit(discord.GuildID(2))) {
		t.Fatal("expected other servers to have their own limit")
	}
}
