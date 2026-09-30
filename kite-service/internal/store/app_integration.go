package store

import (
	"context"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

type AppIntegrationStore interface {
	AppIntegrations(ctx context.Context, appID string) ([]*model.AppIntegration, error)
	// SetAppIntegrationEnabled creates or replaces the app's choice.
	SetAppIntegrationEnabled(ctx context.Context, integration *model.AppIntegration) (*model.AppIntegration, error)
}
