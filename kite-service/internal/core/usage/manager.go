package usage

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/core/plan"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/internal/util"
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
	// How long a deleted command, event listener or message can be restored.
	DeletedItemVersionExpiry = 30 * 24 * time.Hour
	// Files are uploaded before their asset is created, so a newer file
	// without an asset may belong to an upload that is still in progress.
	UnusedAssetObjectMinAge = time.Hour

	cleanupBatchSize = 5000
	// Caps a single tick so a large backlog can't stall the credit sweep, the
	// rest is picked up by the next tick.
	cleanupMaxBatches = 100
	// Shorter than the cleanup interval, so asset cleanups never overlap.
	assetCleanupTimeout = 30 * time.Minute
)

type UsageManager struct {
	appStore             store.AppStore
	usageStore           store.UsageStore
	logStore             store.LogStore
	resumePointStore     store.ResumePointStore
	messageInstanceStore store.MessageInstanceStore
	shareCodeStore       store.ShareCodeStore
	itemVersionStore     store.ItemVersionStore
	assetStore           store.AssetStore

	planManager *plan.PlanManager
}

func NewUsageManager(
	appStore store.AppStore,
	usageStore store.UsageStore,
	logStore store.LogStore,
	resumePointStore store.ResumePointStore,
	messageInstanceStore store.MessageInstanceStore,
	shareCodeStore store.ShareCodeStore,
	itemVersionStore store.ItemVersionStore,
	assetStore store.AssetStore,
	planManager *plan.PlanManager,
) *UsageManager {
	return &UsageManager{
		appStore:             appStore,
		usageStore:           usageStore,
		logStore:             logStore,
		resumePointStore:     resumePointStore,
		messageInstanceStore: messageInstanceStore,
		shareCodeStore:       shareCodeStore,
		itemVersionStore:     itemVersionStore,
		assetStore:           assetStore,
		planManager:          planManager,
	}
}

func (m *UsageManager) Run(ctx context.Context) {
	// Each sweep sums the whole month so far, millions of rows by the end of
	// it. Apps may overrun their credits by up to one interval before being
	// disabled, and start again up to one interval after they have credits.
	ticker := time.NewTicker(5 * time.Minute)
	cleanupTicker := time.NewTicker(1 * time.Hour)

	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				if err := m.updateAppsForCredits(ctx); err != nil {
					slog.Error(
						"Failed to update apps for credits",
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
				if err := m.cleanupDeletedItemVersions(ctx); err != nil {
					slog.Error(
						"Failed to cleanup deleted item versions",
						slog.String("error", err.Error()),
					)
				}
				// Listing and deleting files can take a while, so it doesn't
				// hold up the credit checks of the other ticker.
				go func() {
					ctx, cancel := context.WithTimeout(ctx, assetCleanupTimeout)
					defer cancel()
					if err := m.cleanupAssets(ctx); err != nil {
						slog.Error(
							"Failed to cleanup assets",
							slog.String("error", err.Error()),
						)
					}
				}()
			}
		}
	}()
}

func (m *UsageManager) updateAppsForCredits(ctx context.Context) error {
	// Run's context has no deadline, and this shares its goroutine with the
	// cleanup tickers, so a stuck query would stall those too.
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	start, end := util.StartAndEndOfMonth(time.Now().UTC())

	creditsUsed, err := m.usageStore.AllUsageCreditsUsedBetween(ctx, start, end)
	if err != nil {
		return fmt.Errorf("failed to get all usage credits used: %w", err)
	}

	stopped, err := m.appStore.AppIDsDisabledWithReason(ctx, model.AppDisabledReasonNoCredits)
	if err != nil {
		return fmt.Errorf("failed to get apps with no credits: %w", err)
	}

	// An app can only be over its limit if it is over the allowance every app
	// gets for free, so the rest need no entitlement lookup at all. That is
	// the large majority of them. Stopped apps are always checked, to start
	// them again once they're under their limit.
	floor := m.planManager.DefaultFeatures().UsageCreditsPerMonth

	isStopped := make(map[string]bool, len(stopped))
	for _, appID := range stopped {
		isStopped[appID] = true
	}

	candidates := stopped
	for appID, used := range creditsUsed {
		if used >= floor && !isStopped[appID] {
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
		limit := features[appID].UsageCreditsPerMonth
		over := creditsUsed[appID] >= limit
		if over && !isStopped[appID] {
			m.disableApp(ctx, appID, limit)
		} else if !over && isStopped[appID] {
			m.enableApp(ctx, appID)
		}
	}

	return nil
}

// disableApp and enableApp are separate functions so their context is released
// when the app is done rather than accumulating until the whole sweep returns.
func (m *UsageManager) disableApp(ctx context.Context, appID string, limit int) {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	disabled, err := m.appStore.DisableApp(ctx, store.AppDisableOpts{
		ID:             appID,
		DisabledReason: null.StringFrom(model.AppDisabledReasonNoCredits),
		UpdatedAt:      time.Now().UTC(),
	})
	if err != nil {
		slog.Error(
			"Failed to disable app with no credits",
			slog.String("app_id", appID),
			slog.String("error", err.Error()),
		)
		return
	}

	if disabled {
		m.createLogEntry(ctx, appID, model.LogLevelError, fmt.Sprintf(
			"Your app used all of its %d credits for this month and was stopped. %s",
			limit, model.AppNoCreditsRestartNote,
		))
	}
}

func (m *UsageManager) enableApp(ctx context.Context, appID string) {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	enabled, err := m.appStore.EnableAppDisabledWithReason(ctx, appID, model.AppDisabledReasonNoCredits, time.Now().UTC())
	if err != nil {
		slog.Error(
			"Failed to enable app with credits",
			slog.String("app_id", appID),
			slog.String("error", err.Error()),
		)
		return
	}

	if enabled {
		m.createLogEntry(ctx, appID, model.LogLevelInfo, "Your app has credits again and was started.")
	}
}

func (m *UsageManager) createLogEntry(ctx context.Context, appID string, level model.LogLevel, message string) {
	err := m.logStore.CreateLogEntry(ctx, model.LogEntry{
		AppID:     appID,
		Level:     level,
		Message:   message,
		CreatedAt: time.Now().UTC(),
	})
	if err != nil {
		slog.Error(
			"Failed to create credits log entry",
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

func (m *UsageManager) cleanupDeletedItemVersions(ctx context.Context) error {
	deletedBefore := time.Now().UTC().Add(-DeletedItemVersionExpiry)

	// Batches count items, which have up to MaxItemVersions versions each, so
	// it keeps going until nothing is left instead of until a batch is short.
	for range cleanupMaxBatches {
		batchCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		deleted, err := m.itemVersionStore.DeleteDeletedItemVersions(batchCtx, deletedBefore, cleanupBatchSize/model.MaxItemVersions)
		cancel()
		if err != nil {
			return err
		}
		if deleted == 0 {
			return nil
		}
	}
	return nil
}

func (m *UsageManager) cleanupAssets(ctx context.Context) error {
	if err := m.assetStore.DeleteExpiredAssets(ctx, time.Now().UTC()); err != nil {
		return fmt.Errorf("failed to delete expired assets: %w", err)
	}

	if err := m.assetStore.DeleteUnusedObjects(ctx, UnusedAssetObjectMinAge); err != nil {
		return fmt.Errorf("failed to delete unused asset objects: %w", err)
	}

	return nil
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
