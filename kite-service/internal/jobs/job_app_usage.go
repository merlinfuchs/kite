package jobs

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/merlinfuchs/kite/kite-service/internal/logging/logattr"
	"github.com/merlinfuchs/kite/kite-service/pkg/model"
	"github.com/merlinfuchs/kite/kite-service/pkg/store"
	"github.com/riverqueue/river"
)

const JobKindPopulateAppUsage = "app_usage_populate"

type AppUsagePopulateArgs struct{}

func (AppUsagePopulateArgs) Kind() string { return JobKindPopulateAppUsage }

func (AppUsagePopulateArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		UniqueOpts: river.UniqueOpts{
			ByArgs:   true,
			ByPeriod: 15 * time.Minute,
		},
	}
}

func (a AppUsagePopulateArgs) PeriodicJob() *river.PeriodicJob {
	return river.NewPeriodicJob(
		river.PeriodicInterval(15*time.Minute),
		func() (river.JobArgs, *river.InsertOpts) {
			return a, nil
		},
		&river.PeriodicJobOpts{RunOnStart: true},
	)
}

type AppUsagePopulateWorker struct {
	river.WorkerDefaults[AppUsagePopulateArgs]

	apps              store.AppStore
	deploymentMetrics store.DeploymentMetricStore
	appUsages         store.AppUsageStore
}

func (w *AppUsagePopulateWorker) Work(ctx context.Context, job *river.Job[AppUsagePopulateArgs]) error {
	appIDs, err := w.apps.GetDistinctAppIDs(ctx)
	if err != nil {
		return fmt.Errorf("Failed to get distinct app IDs: %v", err)
	}

	failedCount := 0

	for _, appID := range appIDs {
		lastUsage, err := w.appUsages.GetLastAppUsageEntry(ctx, appID)
		if err != nil && err != store.ErrNotFound {
			slog.With(logattr.Error(err)).Error("Failed to get last app usage entry")
			failedCount++
			continue
		}

		var lastUsagePeriodEnd time.Time
		if lastUsage != nil {
			lastUsagePeriodEnd = lastUsage.PeriodEndsAt
		}

		now := time.Now().UTC()

		summary, err := w.deploymentMetrics.GetDeploymentsMetricsSummary(
			ctx,
			appID,
			lastUsagePeriodEnd,
			now,
		)
		if err != nil {
			slog.With(logattr.Error(err)).Error("Failed to get deployments metrics summary")
			failedCount++
			continue
		}

		if err := w.appUsages.CreateAppUsageEntry(ctx, model.AppUsageEntry{
			AppID:                   appID,
			TotalEventCount:         summary.TotalEventCount,
			SuccessEventCount:       summary.SuccessEventCount,
			TotalEventExecutionTime: summary.TotalEventExecutionTime,
			AvgEventExecutionTime:   summary.AvgEventExecutionTime,
			TotalEventTotalTime:     summary.TotalEventTotalTime,
			AvgEventTotalTime:       summary.AvgEventTotalTime,
			TotalCallCount:          summary.TotalCallCount,
			SuccessCallCount:        summary.SuccessCallCount,
			TotalCallTotalTime:      summary.TotalCallTotalTime,
			AvgCallTotalTime:        summary.AvgCallTotalTime,
			PeriodStartsAt:          summary.FirstEntryAt,
			PeriodEndsAt:            now,
		}); err != nil {
			slog.With(logattr.Error(err)).Error("Failed to insert app usage entry")
			continue
		}
	}

	slog.
		With("total", len(appIDs)).
		With("failed", failedCount).
		Info("Finished populating app usage entries")
	return nil
}
