package store

import (
	"context"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

type AppSecretStore interface {
	AppSecretsByApp(ctx context.Context, appID string) ([]*model.AppSecret, error)
	AppSecret(ctx context.Context, appID string, id string) (*model.AppSecret, error)
	// AppSecretsByNames returns the secrets of the app with the given names.
	// Names that don't exist are left out.
	AppSecretsByNames(ctx context.Context, appID string, names []string) ([]*model.AppSecret, error)
	CountAppSecretsByApp(ctx context.Context, appID string) (int, error)
	CreateAppSecret(ctx context.Context, secret *model.AppSecret) (*model.AppSecret, error)
	UpdateAppSecret(ctx context.Context, secret *model.AppSecret) (*model.AppSecret, error)
	DeleteAppSecret(ctx context.Context, appID string, id string) error

	AppIntegrationCredentials(ctx context.Context, appID string) ([]*model.AppSecret, error)
	AppIntegrationCredential(ctx context.Context, appID string, integrationID string) (*model.AppSecret, error)
}
