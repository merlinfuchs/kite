package store

import (
	"context"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

type AppIntegrationStore interface {
	AppIntegrations(ctx context.Context, appID string) ([]*model.AppIntegration, error)
	// SetAppIntegrationEnabled creates or replaces the app's choice.
	SetAppIntegrationEnabled(ctx context.Context, integration *model.AppIntegration) (*model.AppIntegration, error)
	// UpdateAppIntegrationEnabled changes the choice of an integration the app
	// set up, and returns ErrNotFound for others.
	UpdateAppIntegrationEnabled(ctx context.Context, integration *model.AppIntegration) (*model.AppIntegration, error)
	// ConnectAppIntegration sets the credential of an integration. An
	// integration the app hadn't set up is enabled with it.
	ConnectAppIntegration(ctx context.Context, secret *model.AppSecret) error
	// DeleteAppIntegration removes an integration the app set up, with its
	// credential.
	DeleteAppIntegration(ctx context.Context, appID string, integrationID string) error
}
