package provider

import (
	"context"
	"sync"
	"time"
)

// CooldownProvider tracks cooldowns for cooldown blocks in flows.
type CooldownProvider interface {
	// CheckAndStart reports how much longer key is on cooldown for. If key
	// isn't on cooldown (or its previous cooldown has expired), it starts a
	// new cooldown of duration and returns zero.
	CheckAndStart(ctx context.Context, key string, duration time.Duration) (time.Duration, error)
}

// MemoryCooldownProvider tracks cooldowns in-process. Cooldowns are lost on
// restart and aren't shared across clustered instances -- fine for a
// single-instance deployment, but a multi-cluster setup would need a shared
// store (e.g. Redis or Postgres) instead.
type MemoryCooldownProvider struct {
	mu      sync.Mutex
	expires map[string]time.Time
}

func NewMemoryCooldownProvider() *MemoryCooldownProvider {
	return &MemoryCooldownProvider{
		expires: make(map[string]time.Time),
	}
}

func (p *MemoryCooldownProvider) CheckAndStart(ctx context.Context, key string, duration time.Duration) (time.Duration, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()

	if expiresAt, ok := p.expires[key]; ok && now.Before(expiresAt) {
		return expiresAt.Sub(now), nil
	}

	p.expires[key] = now.Add(duration)

	// Occasionally sweep expired entries so the map doesn't grow forever.
	if len(p.expires)%512 == 0 {
		for k, v := range p.expires {
			if now.After(v) {
				delete(p.expires, k)
			}
		}
	}

	return 0, nil
}

type MockCooldownProvider struct{}

func (p *MockCooldownProvider) CheckAndStart(ctx context.Context, key string, duration time.Duration) (time.Duration, error) {
	return 0, nil
}
