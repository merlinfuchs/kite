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
	// Kept after they expire, so a code used again soon after its first use
	// is still found and revokes its token.
	UsedOAuthCodeExpiry = time.Hour

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
	apiTokenStore        store.APITokenStore
	oauthCodeStore       store.OAuthCodeStore

	planManager    *plan.PlanManager
	creditNotifier CreditNotifier
}

// CreditNotifier tells the owner of an app about its credits.
type CreditNotifier interface {
	CreditsLow(app *model.App, used int, limit int, resetAt time.Time)
	CreditsUsedUp(app *model.App, limit int, resetAt time.Time)
}

// creditsWarnPercent is the share of its credits after which the owner of an
// app gets a DM, once a month.
const creditsWarnPercent = 80

func NewUsageManager(
	appStore store.AppStore,
	usageStore store.UsageStore,
	logStore store.LogStore,
	resumePointStore store.ResumePointStore,
	messageInstanceStore store.MessageInstanceStore,
	shareCodeStore store.ShareCodeStore,
	itemVersionStore store.ItemVersionStore,
	assetStore store.AssetStore,
	apiTokenStore store.APITokenStore,
	oauthCodeStore store.OAuthCodeStore,
	planManager *plan.PlanManager,
	creditNotifier CreditNotifier,
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
		apiTokenStore:        apiTokenStore,
		oauthCodeStore:       oauthCodeStore,
		planManager:          planManager,
		creditNotifier:       creditNotifier,
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
				if err := m.cleanupOAuth(ctx); err != nil {
					slog.Error(
						"Failed to cleanup OAuth codes and tokens",
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

	paused, err := m.appStore.CreditsPausedAppIDs(ctx)
	if err != nil {
		return fmt.Errorf("failed to get apps paused for credits: %w", err)
	}

	// An app can only be near its limit if it is near the allowance every app
	// gets for free, so the rest need no entitlement lookup at all. That is
	// the large majority of them. Paused apps are always checked, to unpause
	// them once they're under their limit.
	floor := m.planManager.DefaultFeatures().UsageCreditsPerMonth * creditsWarnPercent / 100

	isPaused := make(map[string]bool, len(paused))
	for _, appID := range paused {
		isPaused[appID] = true
	}

	candidates := paused
	for appID, used := range creditsUsed {
		if used >= floor && !isPaused[appID] {
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

	resetAt := start.AddDate(0, 1, 0)
	for _, appID := range candidates {
		limit := features[appID].UsageCreditsPerMonth
		used := creditsUsed[appID]
		over := used >= limit
		switch {
		case over && !isPaused[appID]:
			m.pauseApp(ctx, appID, limit, resetAt)
		case !over && isPaused[appID]:
			m.unpauseApp(ctx, appID)
		case !over && used*100 >= limit*creditsWarnPercent:
			m.warnApp(ctx, appID, used, limit, start, resetAt)
		}
	}

	return nil
}

// pauseApp and unpauseApp are separate functions so their context is released
// when the app is done rather than accumulating until the whole sweep returns.
func (m *UsageManager) pauseApp(ctx context.Context, appID string, limit int, resetAt time.Time) {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	paused, err := m.appStore.PauseAppForCredits(ctx, appID, time.Now().UTC())
	if err != nil {
		slog.Error(
			"Failed to pause app with no credits",
			slog.String("app_id", appID),
			slog.String("error", err.Error()),
		)
		return
	}

	if paused {
		m.createLogEntry(ctx, appID, model.LogLevelError, fmt.Sprintf(
			"Your app used all of its %d credits for this month, so its commands and events are paused. It runs again by itself when the next month begins, or a few minutes after you upgrade your plan.",
			limit,
		))

		if app := m.appForNotice(ctx, appID); app != nil {
			m.creditNotifier.CreditsUsedUp(app, limit, resetAt)
		}
	}
}

// warnApp tells the owner that the app used most of its credits, once in the
// month that started at monthStart.
func (m *UsageManager) warnApp(ctx context.Context, appID string, used int, limit int, monthStart time.Time, resetAt time.Time) {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	claimed, err := m.appStore.ClaimAppCreditsWarning(ctx, appID, monthStart, time.Now().UTC())
	if err != nil {
		slog.Error(
			"Failed to claim credits warning of app",
			slog.String("app_id", appID),
			slog.String("error", err.Error()),
		)
		return
	}
	if !claimed {
		return
	}

	if app := m.appForNotice(ctx, appID); app != nil {
		m.creditNotifier.CreditsLow(app, used, limit, resetAt)
	}
}

// appForNotice returns the app if its owner wants DMs about its credits.
func (m *UsageManager) appForNotice(ctx context.Context, appID string) *model.App {
	app, err := m.appStore.App(ctx, appID)
	if err != nil {
		slog.Error(
			"Failed to get app for credits notice",
			slog.String("app_id", appID),
			slog.String("error", err.Error()),
		)
		return nil
	}
	if !app.CreditNoticesEnabled {
		return nil
	}
	return app
}

func (m *UsageManager) unpauseApp(ctx context.Context, appID string) {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	unpaused, err := m.appStore.UnpauseAppForCredits(ctx, appID, time.Now().UTC())
	if err != nil {
		slog.Error(
			"Failed to unpause app with credits",
			slog.String("app_id", appID),
			slog.String("error", err.Error()),
		)
		return
	}

	if unpaused {
		m.createLogEntry(ctx, appID, model.LogLevelInfo, "Your app has credits again, its commands and events run again.")
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

// cleanupOAuth deletes old codes, and the tokens of OAuth clients whose
// refresh token expired, which clients replace by connecting again.
func (m *UsageManager) cleanupOAuth(ctx context.Context) error {
	now := time.Now().UTC()

	err := deleteInBatches(ctx, func(ctx context.Context) (int64, error) {
		return m.oauthCodeStore.DeleteExpiredOAuthCodes(ctx, now.Add(-UsedOAuthCodeExpiry), cleanupBatchSize)
	})
	if err != nil {
		return fmt.Errorf("failed to delete expired OAuth codes: %w", err)
	}

	return deleteInBatches(ctx, func(ctx context.Context) (int64, error) {
		return m.apiTokenStore.DeleteExpiredOAuthAPITokens(ctx, now, cleanupBatchSize)
	})
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
