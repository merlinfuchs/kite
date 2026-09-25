package usage

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/core/plan"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"gopkg.in/guregu/null.v4"
)

const (
	// Only the current month is ever read, the rest is slack for support.
	UsageRecordExpiry = 40 * 24 * time.Hour
	LogEntryExpiry    = 30 * 24 * time.Hour

	// Buttons stop working once these expire, so they count from last use, not creation.
	ResumePointExpiry              = 90 * 24 * time.Hour
	FlowMessageInstanceExpiry      = 90 * 24 * time.Hour
	DashboardMessageInstanceExpiry = 360 * 24 * time.Hour
	ShareCodeExpiry                = 90 * 24 * time.Hour

	// Tombstones only need to outlive the engine and gateway polls.
	DeletedEntityExpiry = 24 * time.Hour

	cleanupBatchSize = 5000
	// Caps a single tick so a large backlog can't stall the credit sweep, the
	// rest is picked up by the next tick.
	cleanupMaxBatches = 100
)

type UsageManager struct {
	appStore             store.AppStore
	usageStore           store.UsageStore
	logStore             store.LogStore
	resumePointStore     store.ResumePointStore
	messageInstanceStore store.MessageInstanceStore
	shareCodeStore       store.ShareCodeStore
	deletedEntityStore   store.DeletedEntityStore

	planManager *plan.PlanManager
}

func NewUsageManager(
	appStore store.AppStore,
	usageStore store.UsageStore,
	logStore store.LogStore,
	resumePointStore store.ResumePointStore,
	messageInstanceStore store.MessageInstanceStore,
	shareCodeStore store.ShareCodeStore,
	deletedEntityStore store.DeletedEntityStore,
	planManager *plan.PlanManager,
) *UsageManager {
	return &UsageManager{
		appStore:             appStore,
		usageStore:           usageStore,
		logStore:             logStore,
		resumePointStore:     resumePointStore,
		messageInstanceStore: messageInstanceStore,
		shareCodeStore:       shareCodeStore,
		deletedEntityStore:   deletedEntityStore,
		planManager:          planManager,
	}
}

func (m *UsageManager) Run(ctx context.Context) {
	// Each sweep sums the whole month so far, millions of rows by the end of
	// it. Apps may overrun their credits by up to one interval before being
	// disabled.
	ticker := time.NewTicker(5 * time.Minute)
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
				if err := m.cleanupResumePoints(ctx); err != nil {
					slog.Error(
						"Failed to cleanup resume points",
						slog.String("error", err.Error()),
					)
				}
				if err := m.cleanupMessageInstances(ctx); err != nil {
					slog.Error(
						"Failed to cleanup message instances",
						slog.String("error", err.Error()),
					)
				}
				if err := m.cleanupShareCodes(ctx); err != nil {
					slog.Error(
						"Failed to cleanup share codes",
						slog.String("error", err.Error()),
					)
				}
				if err := m.cleanupDeletedEntities(ctx); err != nil {
					slog.Error(
						"Failed to cleanup deleted entities",
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

	start, end := startAndEndOfMonth(time.Now().UTC())

	creditsUsed, err := m.usageStore.AllUsageCreditsUsedBetween(ctx, start, end)
	if err != nil {
		return fmt.Errorf("failed to get all usage credits used: %w", err)
	}

	// An app can only be over its limit if it is over the allowance every app
	// gets for free, so the rest need no entitlement lookup at all. That is
	// the large majority of them.
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

	err := deleteInBatches(ctx, func(ctx context.Context) (int64, error) {
		return m.usageStore.DeleteUsageRecordsBefore(ctx, expiry, cleanupBatchSize)
	})
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

func (m *UsageManager) cleanupResumePoints(ctx context.Context) error {
	now := time.Now().UTC()

	return deleteInBatches(ctx, func(ctx context.Context) (int64, error) {
		return m.resumePointStore.DeleteStaleResumePoints(ctx, now, now.Add(-ResumePointExpiry), cleanupBatchSize)
	})
}

func (m *UsageManager) cleanupMessageInstances(ctx context.Context) error {
	now := time.Now().UTC()

	return deleteInBatches(ctx, func(ctx context.Context) (int64, error) {
		return m.messageInstanceStore.DeleteUnusedMessageInstances(
			ctx,
			now.Add(-FlowMessageInstanceExpiry),
			now.Add(-DashboardMessageInstanceExpiry),
			cleanupBatchSize,
		)
	})
}

func (m *UsageManager) cleanupShareCodes(ctx context.Context) error {
	usedBefore := time.Now().UTC().Add(-ShareCodeExpiry)

	return deleteInBatches(ctx, func(ctx context.Context) (int64, error) {
		return m.shareCodeStore.DeleteUnusedShareCodes(ctx, usedBefore, cleanupBatchSize)
	})
}

func (m *UsageManager) cleanupDeletedEntities(ctx context.Context) error {
	expiry := time.Now().UTC().Add(-DeletedEntityExpiry)

	return deleteInBatches(ctx, func(ctx context.Context) (int64, error) {
		return m.deletedEntityStore.DeleteDeletedEntitiesBefore(ctx, expiry, cleanupBatchSize)
	})
}

func deleteInBatches(ctx context.Context, deleteBatch func(ctx context.Context) (int64, error)) error {
	for range cleanupMaxBatches {
		batchCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		deleted, err := deleteBatch(batchCtx)
		cancel()
		if err != nil {
			return err
		}
		if deleted < cleanupBatchSize {
			return nil
		}
	}
	return nil
}

func startAndEndOfMonth(t time.Time) (time.Time, time.Time) {
	year, month, _ := t.Date()
	start := time.Date(year, month, 1, 0, 0, 0, 0, t.Location())
	end := start.AddDate(0, 1, 0).Add(-time.Nanosecond)
	return start, end
}
