package engine

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// BlockRateLimit caps how often a group of blocks can run per app.
type BlockRateLimit struct {
	Key   string
	Every time.Duration
	Burst int
}

// gatewayCommandRateLimit is shared by all blocks that send gateway commands.
// arikawa already keeps a connection under Discord's 120 commands per minute
// by queueing, but heartbeats wait in the same queue, so a flow spamming these
// blocks could delay them long enough to cause reconnects.
var gatewayCommandRateLimit = BlockRateLimit{
	Key:   "gateway_command",
	Every: 10 * time.Second,
	Burst: 5,
}

// BlockRateLimiter tracks block rate limits across flow executions.
type BlockRateLimiter struct {
	mu       sync.Mutex
	limiters map[string]*rate.Limiter
}

func NewBlockRateLimiter() *BlockRateLimiter {
	return &BlockRateLimiter{
		limiters: make(map[string]*rate.Limiter),
	}
}

// Allow reports whether the app can run a block with the given limit now, and
// counts the run if so. A nil limiter allows everything.
func (l *BlockRateLimiter) Allow(appID string, limit BlockRateLimit) bool {
	if l == nil {
		return true
	}

	key := appID + ":" + limit.Key

	l.mu.Lock()
	defer l.mu.Unlock()

	limiter, ok := l.limiters[key]
	if !ok {
		limiter = rate.NewLimiter(rate.Every(limit.Every), limit.Burst)
		l.limiters[key] = limiter
	}

	return limiter.Allow()
}

// Sweep drops limiters that have fully refilled, as they behave the same as a
// new one.
func (l *BlockRateLimiter) Sweep() {
	l.mu.Lock()
	defer l.mu.Unlock()

	for key, limiter := range l.limiters {
		if limiter.Tokens() >= float64(limiter.Burst()) {
			delete(l.limiters, key)
		}
	}
}
