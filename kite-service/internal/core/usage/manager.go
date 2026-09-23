package usage

import (
	"context"
	"fmt"
	"log/slog"
	"maps"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/core/plan"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"gopkg.in/guregu/null.v4"
)

const (
	UsageRecordExpiry = 3 * 30 * 24 * time.Hour
	LogEntryExpiry    = 30 * 24 * time.Hour

	// Usage rows get created_at before their insert runs, and the insert has a
	// 30s timeout, so a row stamped longer ago than this is either committed
	// or never will be.
	usageSettleDelay = 5 * time.Minute
)

type UsageManager struct {
	appStore   store.AppStore
	usageStore store.UsageStore
	logStore   store.LogStore

	planManager *plan.PlanManager

	// Credits per app used from settledMonth up to settledUntil, so the sweep
	// only has to read rows added since instead of the whole month. Only
	// touched from the Run goroutine.
	settledMonth time.Time
	settledUntil time.Time
	settled      map[string]int
}

func NewUsageManager(
	appStore store.AppStore,
	usageStore store.UsageStore,
	logStore store.LogStore,
	planManager *plan.PlanManager,
) *UsageManager {
	return &UsageManager{
		appStore:    appStore,
		usageStore:  usageStore,
		logStore:    logStore,
		planManager: planManager,
	}
}

func (m *UsageManager) Run(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	cleanupTicker := time.NewTicker(1 * time.Hour)

	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				if err := m.disableAppsWithNoCredits(ctx); err != nil {
					slog.Error(
						"Failed to disable apps with no credits",
						slog.String("error", err.Error()),
					)
				}
			case <-cleanupTicker.C:
				if err := m.cleanupUsageRecords(ctx); err != nil {
					slog.Error(
						"Failed to cleanup usage records",
						slog.String("error", err.Error()),
					)
				}
				if err := m.cleanupLogEntries(ctx); err != nil {
					slog.Error(
						"Failed to cleanup log entries",
						slog.String("error", err.Error()),
					)
				}
			}
		}
	}()
}

func (m *UsageManager) disableAppsWithNoCredits(ctx context.Context) error {
	// Run's context has no deadline, and this shares its goroutine with the
	// cleanup tickers, so a stuck query would stall those too.
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	creditsUsed, err := m.creditsUsedThisMonth(ctx, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("failed to get all usage credits used: %w", err)
	}

	// An app can only be over its limit if it is over the allowance every app
	// gets for free, so the rest need no entitlement lookup at all. That is
	// the large majority of them. Apps that are already disabled stay in, as
	// DisableApp is a no-op for them.
	floor := m.planManager.DefaultFeatures().UsageCreditsPerMonth

	var candidates []string
	for appID, used := range creditsUsed {
		if used >= floor {
			candidates = append(candidates, appID)
		}
	}

	if len(candidates) == 0 {
		return nil
	}

	features, err := m.planManager.AppFeaturesForApps(ctx, candidates)
	if err != nil {
		return fmt.Errorf("failed to get features for apps: %w", err)
	}

	for _, appID := range candidates {
		if creditsUsed[appID] < features[appID].UsageCreditsPerMonth {
			continue
		}

		m.disableApp(ctx, appID)
	}

	return nil
}

// creditsUsedThisMonth returns the credits every app has used in the month of
// now. Summing the whole month took a scan of millions of rows every minute,
// so rows older than usageSettleDelay are added to m.settled once and only the
// recent tail is read each time.
func (m *UsageManager) creditsUsedThisMonth(ctx context.Context, now time.Time) (map[string]int, error) {
	start, end := startAndEndOfMonth(now)

	if !start.Equal(m.settledMonth) {
		m.settledMonth = start
		m.settledUntil = start
		m.settled = make(map[string]int)
	}

	settleUntil := now.Add(-usageSettleDelay)
	if settleUntil.After(m.settledUntil) {
		newlySettled, err := m.usageStore.AllUsageCreditsUsedBetween(ctx, m.settledUntil, settleUntil)
		if err != nil {
			return nil, err
		}

		for appID, used := range newlySettled {
			m.settled[appID] += used
		}
		m.settledUntil = settleUntil
	}

	recent, err := m.usageStore.AllUsageCreditsUsedBetween(ctx, m.settledUntil, end)
	if err != nil {
		return nil, err
	}

	res := maps.Clone(m.settled)
	for appID, used := range recent {
		res[appID] += used
	}

	return res, nil
}

// disableApp is a separate function so its context is released when the app is
// done rather than accumulating until the whole sweep returns.
func (m *UsageManager) disableApp(ctx context.Context, appID string) {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	err := m.appStore.DisableApp(ctx, store.AppDisableOpts{
		ID:             appID,
		DisabledReason: null.StringFrom("No credits remaining"),
		UpdatedAt:      time.Now().UTC(),
	})
	if err != nil {
		slog.Error(
			"Failed to disable app with no credits",
			slog.String("app_id", appID),
			slog.String("error", err.Error()),
		)
	}
}

func (m *UsageManager) cleanupUsageRecords(ctx context.Context) error {
	expiry := time.Now().UTC().Add(-UsageRecordExpiry)

	err := m.usageStore.DeleteUsageRecordsBefore(ctx, expiry)
	if err != nil {
		return fmt.Errorf("failed to delete usage records: %w", err)
	}

	return nil
}

func (m *UsageManager) cleanupLogEntries(ctx context.Context) error {
	expiry := time.Now().UTC().Add(-LogEntryExpiry)

	err := m.logStore.DeleteLogEntriesBefore(ctx, expiry)
	if err != nil {
		return fmt.Errorf("failed to delete log entries: %w", err)
	}
	return nil
}

// startAndEndOfMonth returns the start of t's month and the start of the next.
func startAndEndOfMonth(t time.Time) (time.Time, time.Time) {
	year, month, _ := t.Date()
	start := time.Date(year, month, 1, 0, 0, 0, 0, t.Location())
	return start, start.AddDate(0, 1, 0)
}
