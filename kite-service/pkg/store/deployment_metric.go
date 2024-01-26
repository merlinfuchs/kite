package store

import (
	"context"
	"time"

	"github.com/merlinfuchs/kite/kite-service/pkg/model"
)

type DeploymentMetricStore interface {
	CreateDeploymentMetricEntry(ctx context.Context, entry model.DeploymentMetricEntry) error
	GetDeploymentMetricEntries(ctx context.Context, args GetDeploymentMetricEntriesArgs) ([]model.DeploymentMetricEntry, error)
}

type GetDeploymentMetricEntriesArgs struct {
	DeploymentID string
	GuildID      string
	StartTime    time.Time
	EndTime      time.Time
	Limit        int
}
