package jobs

import (
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/merlinfuchs/kite/kite-service/pkg/store"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
)

type Client struct {
	*river.Client[pgx.Tx]
}

func NewClient(db *pgxpool.Pool, workers *river.Workers, periodicJobs []*river.PeriodicJob) (Client, error) {
	riverClient, err := river.NewClient(riverpgxv5.New(db), &river.Config{
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {MaxWorkers: 100},
		},
		Workers:      workers,
		PeriodicJobs: periodicJobs,
		ErrorHandler: &ErrorHandler{},
	})
	if err != nil {
		return Client{}, fmt.Errorf("Failed to create river client: %v", err)
	}

	return Client{
		Client: riverClient,
	}, nil
}

func DefaultWorkers(guilds store.GuildStore, deploymentMetrics store.DeploymentMetricStore, guildUsages store.GuildUsageStore) *river.Workers {
	workers := river.NewWorkers()
	river.AddWorker(workers, &GuildUsagePopulateWorker{
		guilds:            guilds,
		deploymentMetrics: deploymentMetrics,
		guildUsages:       guildUsages,
	})
	return workers
}

func DefaultPeriodicJobs() []*river.PeriodicJob {
	return []*river.PeriodicJob{
		GuildUsagePopulateArgs{}.PeriodicJob(),
	}
}
