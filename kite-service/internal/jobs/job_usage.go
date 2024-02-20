package jobs

import (
	"context"
	"fmt"
	"time"

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
		river.PeriodicInterval(1*time.Hour),
		func() (river.JobArgs, *river.InsertOpts) {
			return a, nil
		},
		&river.PeriodicJobOpts{RunOnStart: true},
	)
}

type GuildUsagePopulateWorker struct {
	river.WorkerDefaults[GuildUsagePopulateArgs]

	deploymentMetrics store.DeploymentMetricStore
}

func (w *GuildUsagePopulateWorker) Work(ctx context.Context, job *river.Job[GuildUsagePopulateArgs]) error {
	fmt.Println("yeet")
	return nil
}
