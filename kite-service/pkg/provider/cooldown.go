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
	// new cooldown of duration and returns zero remaining, along with when
	// the new cooldown expires.
	CheckAndStart(ctx context.Context, key string, duration time.Duration) (remaining time.Duration, expiresAt time.Time, err error)
	// Reset ends the cooldown for key early, but only if it's still the one
	// that expires at expiresAt, so it never ends a cooldown that someone else
	// started after this one expired.
	Reset(ctx context.Context, key string, expiresAt time.Time) error
}

// cooldownSweepInterval is how often MemoryCooldownProvider removes expired
// cooldowns.
const cooldownSweepInterval = time.Minute

// MemoryCooldownProvider tracks cooldowns in-process. Cooldowns are lost on
// restart and aren't shared across clustered instances -- fine for a
// single-instance deployment and short cooldowns, but a multi-cluster setup
// would need a shared store (e.g. Redis or Postgres) instead.
type MemoryCooldownProvider struct {
	mu      sync.Mutex
	expires map[string]time.Time
}

func NewMemoryCooldownProvider() *MemoryCooldownProvider {
	return &MemoryCooldownProvider{
		expires: make(map[string]time.Time),
	}
}

// Run periodically removes expired cooldowns so the map doesn't grow forever.
// It returns once ctx is done.
func (p *MemoryCooldownProvider) Run(ctx context.Context) {
	ticker := time.NewTicker(cooldownSweepInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.sweep(time.Now())
		}
	}
}

func (p *MemoryCooldownProvider) sweep(now time.Time) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for k, v := range p.expires {
		if !now.Before(v) {
			delete(p.expires, k)
		}
	}
}

func (p *MemoryCooldownProvider) CheckAndStart(ctx context.Context, key string, duration time.Duration) (time.Duration, time.Time, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()

	if expiresAt, ok := p.expires[key]; ok && now.Before(expiresAt) {
		return expiresAt.Sub(now), expiresAt, nil
	}

	expiresAt := now.Add(duration)
	p.expires[key] = expiresAt
	return 0, expiresAt, nil
}

func (p *MemoryCooldownProvider) Reset(ctx context.Context, key string, expiresAt time.Time) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if current, ok := p.expires[key]; ok && current.Equal(expiresAt) {
		delete(p.expires, key)
	}
	return nil
}

type MockCooldownProvider struct{}

func (p *MockCooldownProvider) CheckAndStart(ctx context.Context, key string, duration time.Duration) (time.Duration, time.Time, error) {
	return 0, time.Time{}, nil
}

func (p *MockCooldownProvider) Reset(ctx context.Context, key string, expiresAt time.Time) error {
	return nil
}
