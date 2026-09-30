package integration

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/core/appintegration"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/internal/util"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
)

type IntegrationHandler struct {
	appSecretStore      store.AppSecretStore
	appIntegrationStore store.AppIntegrationStore
	tokenCrypt          *util.SymmetricCrypt
	// Checks credentials. Integrations' URLs come from their definitions,
	// never from users.
	client *http.Client
}

func NewIntegrationHandler(appSecretStore store.AppSecretStore, appIntegrationStore store.AppIntegrationStore, tokenCrypt *util.SymmetricCrypt) *IntegrationHandler {
	return &IntegrationHandler{
		appSecretStore:      appSecretStore,
		appIntegrationStore: appIntegrationStore,
		tokenCrypt:          tokenCrypt,
		client: &http.Client{
			Timeout: 10 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
}

func (h *IntegrationHandler) HandleAppIntegrationList(c *handler.Context) (*wire.AppIntegrationListResponse, error) {
	states, err := appintegration.States(c.Context(), h.appSecretStore, h.appIntegrationStore, c.App.ID)
	if err != nil {
		return nil, err
	}

	res := make([]*wire.AppIntegration, len(states))
	for i, state := range states {
		res[i] = &wire.AppIntegration{
			IntegrationID:       state.Integration.ID,
			Enabled:             state.Enabled,
			CredentialUpdatedAt: state.CredentialUpdatedAt,
		}
	}
	return &res, nil
}

// HandleAppIntegrationUpdate enables or disables an integration. Integrations
// that need a credential are enabled by connecting them, and keep it while
// they're disabled.
func (h *IntegrationHandler) HandleAppIntegrationUpdate(c *handler.Context, req wire.AppIntegrationUpdateRequest) (*wire.AppIntegrationUpdateResponse, error) {
	integration, ok := flow.GetIntegration(c.Param("integrationID"))
	if !ok {
		return nil, handler.ErrNotFound("unknown_integration", "Integration not found")
	}
	if integration.Availability == flow.AvailabilityAlways {
		return nil, handler.ErrBadRequest("always_enabled", "The integration is always enabled")
	}
	choice := &model.AppIntegration{
		AppID:         c.App.ID,
		IntegrationID: integration.ID,
		Enabled:       *req.Enabled,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}
	// Integrations that need a credential have a row once it's entered.
	set := h.appIntegrationStore.SetAppIntegrationEnabled
	if integration.NeedsCredential() {
		set = h.appIntegrationStore.UpdateAppIntegrationEnabled
	}
	if _, err := set(c.Context(), choice); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, handler.ErrBadRequest("not_connected", "Enable the integration by entering its credential first")
		}
		return nil, fmt.Errorf("failed to update integration: %w", err)
	}
	return &wire.AppIntegrationUpdateResponse{}, nil
}

// HandleAppIntegrationConnect sets the credential of an integration, which
// enables it unless the app disabled it before.
func (h *IntegrationHandler) HandleAppIntegrationConnect(c *handler.Context, req wire.AppIntegrationConnectRequest) (*wire.AppIntegrationConnectResponse, error) {
	integration, err := credentialIntegration(c.Param("integrationID"))
	if err != nil {
		return nil, err
	}

	if err := h.checkCredential(c.Context(), integration, req.Credential); err != nil {
		return nil, err
	}

	value, err := h.tokenCrypt.EncryptString(req.Credential)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt credential: %w", err)
	}

	err = h.appIntegrationStore.ConnectAppIntegration(c.Context(), &model.AppSecret{
		ID:             util.UniqueID(),
		AppID:          c.App.ID,
		IntegrationID:  integration.ID,
		ValueEncrypted: value,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to save credential: %w", err)
	}
	return &wire.AppIntegrationConnectResponse{}, nil
}

// HandleAppIntegrationRemove removes an integration the app set up, with its
// credential, so it's back to its default.
func (h *IntegrationHandler) HandleAppIntegrationRemove(c *handler.Context) (*wire.AppIntegrationRemoveResponse, error) {
	integration, ok := flow.GetIntegration(c.Param("integrationID"))
	if !ok {
		return nil, handler.ErrNotFound("unknown_integration", "Integration not found")
	}

	err := h.appIntegrationStore.DeleteAppIntegration(c.Context(), c.App.ID, integration.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, handler.ErrNotFound("not_set_up", "The app didn't set up the integration")
		}
		return nil, fmt.Errorf("failed to remove integration: %w", err)
	}

	return &wire.AppIntegrationRemoveResponse{}, nil
}

// credentialIntegration returns the integration with the given ID, if apps
// connect it with a credential.
func credentialIntegration(id string) (flow.Integration, error) {
	integration, ok := flow.GetIntegration(id)
	if !ok {
		return flow.Integration{}, handler.ErrNotFound("unknown_integration", "Integration not found")
	}
	if !integration.NeedsCredential() {
		return flow.Integration{}, handler.ErrBadRequest("no_credential", "The integration doesn't need a credential")
	}
	return integration, nil
}

// checkCredential calls the integration's test endpoint, if it has one, and
// rejects credentials the integration refuses. Other failures don't say
// whether the credential is wrong, so they don't stop the app from saving it.
func (h *IntegrationHandler) checkCredential(ctx context.Context, integration flow.Integration, credential string) error {
	if integration.TestPath == "" {
		return nil
	}

	req, err := integration.NewRequest(ctx, http.MethodGet, integration.TestPath, credential, nil)
	if err != nil {
		return fmt.Errorf("failed to create test request: %w", err)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return nil
	}
	resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return handler.ErrBadRequest("invalid_credential", fmt.Sprintf("%s didn't accept the %s", integration.Name, strings.ToLower(integration.Auth.Label)))
	}
	return nil
}
