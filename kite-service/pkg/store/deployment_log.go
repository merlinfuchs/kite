package store

import (
	"context"
	"time"

	"github.com/merlinfuchs/kite/kite-service/pkg/model"
)

type DeploymentLogStore interface {
	CreateDeploymentLogEntry(ctx context.Context, entry model.DeploymentLogEntry) error
	GetDeploymentLogEntries(ctx context.Context, deploymentID string, guildID string) ([]model.DeploymentLogEntry, error)
	GetDeploymentLogSummary(ctx context.Context, deploymentID string, guildID string, cutoff time.Time) (*model.DeploymentLogSummary, error)
}
