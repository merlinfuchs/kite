package store

import (
	"context"
	"time"

	"github.com/merlinfuchs/kite/kite-service/pkg/model"
)

type DeploymentMetricStore interface {
	CreateDeploymentMetricEntry(ctx context.Context, entry model.DeploymentMetricEntry) error
	GetDeploymentsMetricsSummary(ctx context.Context, guildID string, startAt time.Time, endAt time.Time) (model.DeploymentMetricsSummary, error)
	GetDeploymentEventMetrics(ctx context.Context, deploymentID string, startAt time.Time, groupBy time.Duration) ([]model.DeploymentEventMetricEntry, error)
	GetDeploymentsEventMetrics(ctx context.Context, guildID string, startAt time.Time, groupBy time.Duration) ([]model.DeploymentEventMetricEntry, error)
	GetDeploymentCallMetrics(ctx context.Context, deploymentID string, startAt time.Time, groupBy time.Duration) ([]model.DeploymentCallMetricEntry, error)
	GetDeploymentsCallMetrics(ctx context.Context, guildID string, startAt time.Time, groupBy time.Duration) ([]model.DeploymentCallMetricEntry, error)
}
