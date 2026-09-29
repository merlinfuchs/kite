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
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/internal/util"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
)

type IntegrationHandler struct {
	appSecretStore store.AppSecretStore
	tokenCrypt     *util.SymmetricCrypt
	// Checks credentials. Integrations' URLs come from their definitions,
	// never from users.
	client *http.Client
}

func NewIntegrationHandler(appSecretStore store.AppSecretStore, tokenCrypt *util.SymmetricCrypt) *IntegrationHandler {
	return &IntegrationHandler{
		appSecretStore: appSecretStore,
		tokenCrypt:     tokenCrypt,
		client: &http.Client{
			Timeout: 10 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
}

func (h *IntegrationHandler) HandleAppIntegrationList(c *handler.Context) (*wire.AppIntegrationListResponse, error) {
	credentials, err := h.appSecretStore.AppIntegrationCredentials(c.Context(), c.App.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get integrations: %w", err)
	}

	res := make([]*wire.AppIntegration, len(credentials))
	for i, credential := range credentials {
		res[i] = wire.AppIntegrationToWire(credential)
	}
	return &res, nil
}

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

	secret, err := h.appSecretStore.SetAppIntegrationCredential(c.Context(), &model.AppSecret{
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

	return wire.AppIntegrationToWire(secret), nil
}

func (h *IntegrationHandler) HandleAppIntegrationDisconnect(c *handler.Context) (*wire.AppIntegrationDisconnectResponse, error) {
	integration, err := credentialIntegration(c.Param("integrationID"))
	if err != nil {
		return nil, err
	}

	err = h.appSecretStore.DeleteAppIntegrationCredential(c.Context(), c.App.ID, integration.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, handler.ErrNotFound("not_connected", "The integration isn't connected")
		}
		return nil, fmt.Errorf("failed to delete credential: %w", err)
	}

	return &wire.AppIntegrationDisconnectResponse{}, nil
}

// credentialIntegration returns the integration with the given ID, if apps
// connect it with a credential. The others are always connected.
func credentialIntegration(id string) (flow.Integration, error) {
	integration, ok := flow.GetIntegration(id)
	if !ok {
		return flow.Integration{}, handler.ErrNotFound("unknown_integration", "Integration not found")
	}
	if !integration.NeedsCredential() {
		return flow.Integration{}, handler.ErrBadRequest("always_connected", "The integration is always connected")
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
