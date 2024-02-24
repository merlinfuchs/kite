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

const JobKindPopulateGuildUsage = "guild_usage_populate"

type GuildUsagePopulateArgs struct{}

func (GuildUsagePopulateArgs) Kind() string { return JobKindPopulateGuildUsage }

func (GuildUsagePopulateArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		UniqueOpts: river.UniqueOpts{
			ByArgs:   true,
			ByPeriod: 1 * time.Hour,
		},
	}
}

func (a GuildUsagePopulateArgs) PeriodicJob() *river.PeriodicJob {
	return river.NewPeriodicJob(
		river.PeriodicInterval(15*time.Minute),
		func() (river.JobArgs, *river.InsertOpts) {
			return a, nil
		},
		&river.PeriodicJobOpts{RunOnStart: true},
	)
}

type GuildUsagePopulateWorker struct {
	river.WorkerDefaults[GuildUsagePopulateArgs]

	guilds            store.GuildStore
	deploymentMetrics store.DeploymentMetricStore
	guildUsages       store.GuildUsageStore
}

func (w *GuildUsagePopulateWorker) Work(ctx context.Context, job *river.Job[GuildUsagePopulateArgs]) error {
	guildIDs, err := w.guilds.GetDistinctGuildIDs(ctx)
	if err != nil {
		return fmt.Errorf("Failed to get distinct guild IDs: %v", err)
	}

	failedCount := 0

	for _, guildID := range guildIDs {
		lastUsage, err := w.guildUsages.GetLastGuildUsageEntry(ctx, guildID)
		if err != nil && err != store.ErrNotFound {
			slog.With(logattr.Error(err)).Error("Failed to get last guild usage entry")
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
			guildID,
			lastUsagePeriodEnd,
			now,
		)
		if err != nil {
			slog.With(logattr.Error(err)).Error("Failed to get deployments metrics summary")
			failedCount++
			continue
		}

		if err := w.guildUsages.CreateGuildUsageEntry(ctx, model.GuildUsageEntry{
			GuildID:                 guildID,
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
			PeriodStartsAt:          lastUsagePeriodEnd,
			PeriodEndsAt:            now,
		}); err != nil {
			slog.With(logattr.Error(err)).Error("Failed to insert guild usage entry")
			continue
		}
	}

	slog.
		With("total", len(guildIDs)).
		With("failed", failedCount).
		Info("Finished populating guild usage entries")
	return nil
}
