package engine

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
)

const (
	// Changes to an app's limits apply after at most this long.
	creditLimitCacheTTL = 30 * time.Second
	// Usage is re-read from the database after this long, which picks up the
	// executions other clusters ran for the same server or user. In between,
	// executions of this process are counted in memory.
	creditUsageCacheTTL = time.Minute
	// Usage entries that weren't checked for this long are dropped.
	creditUsageIdleExpiry = 10 * time.Minute
	creditPruneInterval   = 5 * time.Minute
)

// CreditLimiter enforces the per-server and per-user credit limits of apps.
//
// It's checked before a flow runs, so executions that start at the same time
// can all pass and overrun a limit by what they use together. The app's own
// monthly credits are enforced separately by the usage manager.
type CreditLimiter struct {
	limitStore store.CreditLimitStore
	usageStore store.UsageStore
	now        func() time.Time

	mu        sync.Mutex
	limits    map[string]cachedCreditLimits
	usage     map[creditUsageKey]*cachedCreditUsage
	lastPrune time.Time
}

type cachedCreditLimits struct {
	limits    []*model.CreditLimit
	fetchedAt time.Time
}

type creditUsageKey struct {
	appID       string
	scope       model.CreditLimitScope
	targetID    string
	periodStart time.Time
}

type cachedCreditUsage struct {
	credits   int
	fetchedAt time.Time
	checkedAt time.Time
	// reported is set once the app was told this server or user hit a limit,
	// so a blocked user spamming a command doesn't flood the logs.
	reported bool
}

// CreditLimitTarget is a server or user an execution runs for.
type CreditLimitTarget struct {
	Scope    model.CreditLimitScope
	TargetID string
}

// CreditLimitExceeded describes the limit that stopped an execution.
type CreditLimitExceeded struct {
	Limit    *model.CreditLimit
	TargetID string
	Used     int
	// FirstReport is true the first time the limit is hit for the target in
	// the period, as far as this process knows.
	FirstReport bool
}

func NewCreditLimiter(limitStore store.CreditLimitStore, usageStore store.UsageStore) *CreditLimiter {
	return &CreditLimiter{
		limitStore: limitStore,
		usageStore: usageStore,
		now:        time.Now,
		limits:     make(map[string]cachedCreditLimits),
		usage:      make(map[creditUsageKey]*cachedCreditUsage),
	}
}

// Check returns the first limit one of the targets has reached, or nil if the
// execution can run. Targets with an empty ID are skipped.
func (l *CreditLimiter) Check(ctx context.Context, appID string, targets ...CreditLimitTarget) (*CreditLimitExceeded, error) {
	limits, err := l.appLimits(ctx, appID)
	if err != nil {
		return nil, err
	}
	if len(limits) == 0 {
		return nil, nil
	}

	now := l.now().UTC()

	for _, target := range targets {
		if target.TargetID == "" {
			continue
		}

		for _, limit := range model.EffectiveCreditLimits(limits, target.Scope, target.TargetID) {
			if !limit.Credits.Valid {
				continue
			}

			key := creditUsageKey{
				appID:       appID,
				scope:       target.Scope,
				targetID:    target.TargetID,
				periodStart: limit.Period.Start(now),
			}

			used, err := l.used(ctx, key, now)
			if err != nil {
				return nil, err
			}

			if used >= int(limit.Credits.Int64) {
				return &CreditLimitExceeded{
					Limit:       limit,
					TargetID:    target.TargetID,
					Used:        used,
					FirstReport: l.markReported(key),
				}, nil
			}
		}
	}

	return nil, nil
}

// Record counts credits an execution used towards the usage cached for its
// targets. Usage that isn't cached is read from the database when it's needed.
func (l *CreditLimiter) Record(appID string, credits int, targets ...CreditLimitTarget) {
	if credits <= 0 {
		return
	}

	now := l.now().UTC()

	l.mu.Lock()
	defer l.mu.Unlock()

	for _, target := range targets {
		if target.TargetID == "" {
			continue
		}

		for _, period := range []model.CreditLimitPeriod{model.CreditLimitPeriodDay, model.CreditLimitPeriodMonth} {
			key := creditUsageKey{
				appID:       appID,
				scope:       target.Scope,
				targetID:    target.TargetID,
				periodStart: period.Start(now),
			}
			if entry, ok := l.usage[key]; ok {
				entry.credits += credits
			}
		}
	}
}

func (l *CreditLimiter) appLimits(ctx context.Context, appID string) ([]*model.CreditLimit, error) {
	now := l.now()

	l.mu.Lock()
	cached, ok := l.limits[appID]
	l.mu.Unlock()

	if ok && now.Sub(cached.fetchedAt) < creditLimitCacheTTL {
		return cached.limits, nil
	}

	limits, err := l.limitStore.CreditLimitsByApp(ctx, appID)
	if err != nil {
		if ok {
			// Keep enforcing the last known limits rather than failing every
			// execution while the database is unavailable.
			return cached.limits, nil
		}
		return nil, fmt.Errorf("failed to get credit limits: %w", err)
	}

	l.mu.Lock()
	l.limits[appID] = cachedCreditLimits{limits: limits, fetchedAt: now}
	l.pruneLocked(now)
	l.mu.Unlock()

	return limits, nil
}

func (l *CreditLimiter) used(ctx context.Context, key creditUsageKey, now time.Time) (int, error) {
	l.mu.Lock()
	entry, ok := l.usage[key]
	if ok {
		entry.checkedAt = now
		if now.Sub(entry.fetchedAt) < creditUsageCacheTTL {
			credits := entry.credits
			l.mu.Unlock()
			return credits, nil
		}
	}
	l.mu.Unlock()

	credits, err := l.usageStore.UsageCreditsUsedByTargetSince(ctx, key.appID, key.scope, key.targetID, key.periodStart)
	if err != nil {
		if ok {
			l.mu.Lock()
			credits := entry.credits
			l.mu.Unlock()
			return credits, nil
		}
		return 0, fmt.Errorf("failed to get credits used: %w", err)
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if entry, ok := l.usage[key]; ok {
		entry.credits = credits
		entry.fetchedAt = now
		entry.checkedAt = now
	} else {
		l.usage[key] = &cachedCreditUsage{credits: credits, fetchedAt: now, checkedAt: now}
	}

	return credits, nil
}

func (l *CreditLimiter) markReported(key creditUsageKey) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry, ok := l.usage[key]
	if !ok || entry.reported {
		return false
	}
	entry.reported = true
	return true
}

// pruneLocked drops usage of past periods and of servers and users that
// weren't seen for a while, and limits of apps that weren't seen for a while.
func (l *CreditLimiter) pruneLocked(now time.Time) {
	if now.Sub(l.lastPrune) < creditPruneInterval {
		return
	}
	l.lastPrune = now

	for key, entry := range l.usage {
		if now.Sub(entry.checkedAt) > creditUsageIdleExpiry {
			delete(l.usage, key)
		}
	}
	for appID, cached := range l.limits {
		if now.Sub(cached.fetchedAt) > creditUsageIdleExpiry {
			delete(l.limits, appID)
		}
	}
}

func creditLimitTargets(guildID string, userID string) []CreditLimitTarget {
	return []CreditLimitTarget{
		{Scope: model.CreditLimitScopeGuild, TargetID: guildID},
		{Scope: model.CreditLimitScopeUser, TargetID: userID},
	}
}

// creditLimitMessage is what users see when they run into a limit.
func creditLimitMessage(exceeded *CreditLimitExceeded) string {
	var period string
	switch exceeded.Limit.Period {
	case model.CreditLimitPeriodDay:
		period = "today"
	default:
		period = "this month"
	}

	switch exceeded.Limit.Scope {
	case model.CreditLimitScopeGuild:
		return fmt.Sprintf("This server has reached its usage limit for %s. Try again later.", period)
	default:
		return fmt.Sprintf("You have reached your usage limit for %s. Try again later.", period)
	}
}

func logCreditLimitError(appID string, err error) {
	slog.Error(
		"Failed to check credit limits",
		slog.String("app_id", appID),
		slog.String("error", err.Error()),
	)
}
