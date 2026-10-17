package provider

import (
	"context"

	"gopkg.in/guregu/null.v4"
)

// IntegrationProvider gives flows the app's setup of its integrations.
type IntegrationProvider interface {
	// Credential returns the credential of an integration, or ErrNotFound if
	// the app didn't connect it.
	Credential(ctx context.Context, integrationID string) (string, error)
	// Choice returns whether the app turned an integration without a
	// credential on or off, or null if it didn't.
	Choice(ctx context.Context, integrationID string) (null.Bool, error)
}

type MockIntegrationProvider struct{}

func (p *MockIntegrationProvider) Credential(ctx context.Context, integrationID string) (string, error) {
	return "", ErrNotFound
}

func (p *MockIntegrationProvider) Choice(ctx context.Context, integrationID string) (null.Bool, error) {
	return null.Bool{}, nil
}
