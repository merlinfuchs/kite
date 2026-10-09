package provider

import (
	"context"
	"testing"
	"time"
)

func TestMemoryCooldownProvider(t *testing.T) {
	ctx := context.Background()
	p := NewMemoryCooldownProvider()

	remaining, expiresAt, err := p.CheckAndStart(ctx, "a", time.Minute)
	if err != nil || remaining != 0 {
		t.Fatalf("first use: got remaining %v, err %v", remaining, err)
	}

	remaining, _, err = p.CheckAndStart(ctx, "a", time.Minute)
	if err != nil || remaining <= 0 {
		t.Fatalf("second use: expected to be on cooldown, got remaining %v, err %v", remaining, err)
	}

	// Other keys aren't affected.
	remaining, _, _ = p.CheckAndStart(ctx, "b", time.Minute)
	if remaining != 0 {
		t.Fatalf("other key: expected no cooldown, got %v", remaining)
	}

	// Resetting with a stale expiry leaves the cooldown alone.
	_ = p.Reset(ctx, "a", expiresAt.Add(time.Second))
	if remaining, _, _ = p.CheckAndStart(ctx, "a", time.Minute); remaining <= 0 {
		t.Fatalf("stale reset: expected cooldown to remain")
	}

	_ = p.Reset(ctx, "a", expiresAt)
	if remaining, _, _ = p.CheckAndStart(ctx, "a", time.Minute); remaining != 0 {
		t.Fatalf("after reset: expected no cooldown, got %v", remaining)
	}
}

func TestMemoryCooldownProviderSweep(t *testing.T) {
	ctx := context.Background()
	p := NewMemoryCooldownProvider()

	_, _, _ = p.CheckAndStart(ctx, "short", time.Second)
	_, _, _ = p.CheckAndStart(ctx, "long", time.Hour)

	p.sweep(time.Now().Add(time.Minute))

	p.mu.Lock()
	defer p.mu.Unlock()
	if _, ok := p.expires["short"]; ok {
		t.Error("expired cooldown wasn't swept")
	}
	if _, ok := p.expires["long"]; !ok {
		t.Error("active cooldown was swept")
	}
}
