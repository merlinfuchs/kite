package store

import (
	"context"

	"github.com/merlinfuchs/kite/kite-service/pkg/model"
)

type DeploymentStore interface {
	UpsertDeployment(ctx context.Context, deployment model.Deployment) (*model.Deployment, error)
	DeleteDeployment(ctx context.Context, deployment model.Deployment) error
	GetDeploymentForGuild(ctx context.Context, id string, guildID string) (*model.Deployment, error)
	GetDeploymentsForGuild(ctx context.Context, guildID string) ([]model.Deployment, error)
	GetGuildIDsWithDeployment(ctx context.Context) ([]string, error)
}
