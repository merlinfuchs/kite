package provider

import "context"

// IntegrationProvider gives flows the credentials the app connected its
// integrations with.
type IntegrationProvider interface {
	// Credential returns the credential of an integration, or ErrNotFound if
	// the app didn't connect it.
	Credential(ctx context.Context, integrationID string) (string, error)
}

type MockIntegrationProvider struct{}

func (p *MockIntegrationProvider) Credential(ctx context.Context, integrationID string) (string, error) {
	return "", ErrNotFound
}
