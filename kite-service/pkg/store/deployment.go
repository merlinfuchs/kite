package store

import (
	"context"
	"time"

	"github.com/merlinfuchs/kite/kite-service/pkg/model"
)

type DeploymentStore interface {
	UpsertDeployment(ctx context.Context, deployment model.Deployment) (*model.Deployment, error)
	DeleteDeployment(ctx context.Context, id string, appID string) error
	GetDeployments(ctx context.Context) ([]model.Deployment, error)
	GetDeployment(ctx context.Context, id string, appID string) (*model.Deployment, error)
	GetDeploymentsForApp(ctx context.Context, appID string) ([]model.Deployment, error)
	GetDeploymentsWithUndeployedChanges(ctx context.Context) ([]model.Deployment, error)
	GetDeploymentIDs(ctx context.Context) ([]model.PartialDeployment, error)
	UpdateDeploymentsDeployedAtForApp(ctx context.Context, appID string, deployedAt time.Time) error
	GetAppIDsWithDeployment(ctx context.Context) ([]string, error)
}
