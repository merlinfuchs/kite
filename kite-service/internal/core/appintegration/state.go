package appintegration

import (
	"context"
	"fmt"

	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"gopkg.in/guregu/null.v4"
)

// CredentialStore is what States needs of store.AppSecretStore.
type CredentialStore interface {
	AppIntegrationCredentials(ctx context.Context, appID string) ([]*model.AppSecret, error)
}

// ChoiceStore is what States needs of store.AppIntegrationStore.
type ChoiceStore interface {
	AppIntegrations(ctx context.Context, appID string) ([]*model.AppIntegration, error)
}

// State is whether an app can use an integration.
type State struct {
	Integration flow.Integration
	Enabled     bool
	// When the app last set the credential, for integrations that need one.
	CredentialUpdatedAt null.Time
}

// States returns the state of every integration for the app.
func States(ctx context.Context, credentials CredentialStore, choices ChoiceStore, appID string) ([]State, error) {
	secrets, err := credentials.AppIntegrationCredentials(ctx, appID)
	if err != nil {
		return nil, fmt.Errorf("failed to get integration credentials: %w", err)
	}
	rows, err := choices.AppIntegrations(ctx, appID)
	if err != nil {
		return nil, fmt.Errorf("failed to get integrations: %w", err)
	}

	updatedAt := make(map[string]null.Time, len(secrets))
	for _, secret := range secrets {
		updatedAt[secret.IntegrationID] = null.TimeFrom(secret.UpdatedAt)
	}
	choice := make(map[string]null.Bool, len(rows))
	for _, row := range rows {
		choice[row.IntegrationID] = null.BoolFrom(row.Enabled)
	}

	integrations := flow.Integrations()
	res := make([]State, len(integrations))
	for i, integration := range integrations {
		res[i] = State{
			Integration: integration,
			// A row without its credential can't run, whatever it says.
			Enabled: integration.Enabled(choice[integration.ID]) &&
				(!integration.NeedsCredential() || updatedAt[integration.ID].Valid),
			CredentialUpdatedAt: updatedAt[integration.ID],
		}
	}
	return res, nil
}
