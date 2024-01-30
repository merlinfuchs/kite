package store

import (
	"context"
	"time"

	"github.com/merlinfuchs/kite/kite-service/pkg/model"
)

type DeploymentStore interface {
	UpsertDeployment(ctx context.Context, deployment model.Deployment) (*model.Deployment, error)
	DeleteDeployment(ctx context.Context, id string, guildID string) error
	GetDeployments(ctx context.Context) ([]model.Deployment, error)
	GetDeployment(ctx context.Context, id string, guildID string) (*model.Deployment, error)
	GetDeploymentsForGuild(ctx context.Context, guildID string) ([]model.Deployment, error)
	GetDeploymentsWithUndeployedChanges(ctx context.Context) ([]model.Deployment, error)
	GetDeploymentIDs(ctx context.Context) ([]model.PartialDeployment, error)
	UpdateDeploymentsDeployedAtForGuild(ctx context.Context, guildID string, deployedAt time.Time) error
	GetGuildIDsWithDeployment(ctx context.Context) ([]string, error)
}
