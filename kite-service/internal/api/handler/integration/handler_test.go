package integration

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// memoryStore keeps credentials and integrations like the database, where
// removing an integration removes its credential.
type memoryStore struct {
	store.AppSecretStore
	credentials  map[string]*model.AppSecret
	integrations map[string]*model.AppIntegration
}

func newMemoryStore() *memoryStore {
	return &memoryStore{
		credentials:  map[string]*model.AppSecret{},
		integrations: map[string]*model.AppIntegration{},
	}
}

func (s *memoryStore) AppIntegrationCredentials(ctx context.Context, appID string) ([]*model.AppSecret, error) {
	var res []*model.AppSecret
	for _, secret := range s.credentials {
		res = append(res, secret)
	}
	return res, nil
}

func (s *memoryStore) AppIntegrationCredential(ctx context.Context, appID string, integrationID string) (*model.AppSecret, error) {
	if secret, ok := s.credentials[integrationID]; ok {
		return secret, nil
	}
	return nil, store.ErrNotFound
}

func (s *memoryStore) AppIntegrations(ctx context.Context, appID string) ([]*model.AppIntegration, error) {
	var res []*model.AppIntegration
	for _, integration := range s.integrations {
		res = append(res, integration)
	}
	return res, nil
}

func (s *memoryStore) SetAppIntegrationEnabled(ctx context.Context, integration *model.AppIntegration) (*model.AppIntegration, error) {
	s.integrations[integration.IntegrationID] = integration
	return integration, nil
}

func (s *memoryStore) UpdateAppIntegrationEnabled(ctx context.Context, integration *model.AppIntegration) (*model.AppIntegration, error) {
	if _, ok := s.integrations[integration.IntegrationID]; !ok {
		return nil, store.ErrNotFound
	}
	s.integrations[integration.IntegrationID] = integration
	return integration, nil
}

func (s *memoryStore) ConnectAppIntegration(ctx context.Context, integration *model.AppIntegration, secret *model.AppSecret) error {
	existing, ok := s.integrations[integration.IntegrationID]
	if !ok {
		integration.Enabled = true
		s.integrations[integration.IntegrationID] = integration
		existing = integration
	}
	secret.AppIntegrationID = existing.ID
	secret.IntegrationID = existing.IntegrationID
	s.credentials[existing.IntegrationID] = secret
	return nil
}

func (s *memoryStore) DeleteAppIntegration(ctx context.Context, appID string, integrationID string) error {
	if _, ok := s.integrations[integrationID]; !ok {
		return store.ErrNotFound
	}
	delete(s.integrations, integrationID)
	delete(s.credentials, integrationID)
	return nil
}

// roundTripper answers requests without a network.
type roundTripper struct {
	req    *http.Request
	status int
}

func (rt *roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	rt.req = req
	return &http.Response{StatusCode: rt.status, Body: io.NopCloser(strings.NewReader("{}"))}, nil
}

// newHandler returns a handler whose credential checks get the given status.
func newHandler(t *testing.T, s *memoryStore, status int) (*IntegrationHandler, *roundTripper) {
	crypt, err := util.NewSymmetricCrypt(strings.Repeat("ab", 32))
	require.NoError(t, err)
	rt := &roundTripper{status: status}
	h := NewIntegrationHandler(s, s, crypt)
	h.client.Transport = rt
	return h, rt
}

type response struct {
	code  int
	data  any
	error map[string]any
}

func serve(t *testing.T, h *IntegrationHandler, method string, path string, body string) response {
	withApp := func(next func(c *handler.Context) error) http.Handler {
		return handler.APIHandler(func(c *handler.Context) error {
			c.App = &model.App{ID: "app"}
			return next(c)
		})
	}

	mux := http.NewServeMux()
	mux.Handle("GET /integrations", withApp(handler.Typed(h.HandleAppIntegrationList)))
	mux.Handle("PATCH /integrations/{integrationID}", withApp(handler.TypedWithBody(h.HandleAppIntegrationUpdate)))
	mux.Handle("PUT /integrations/{integrationID}", withApp(handler.TypedWithBody(h.HandleAppIntegrationConnect)))
	mux.Handle("DELETE /integrations/{integrationID}", withApp(handler.Typed(h.HandleAppIntegrationRemove)))

	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	var res struct {
		Data  any            `json:"data"`
		Error map[string]any `json:"error"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	return response{code: rec.Code, data: res.Data, error: res.Error}
}

func TestConnectChecksCredential(t *testing.T) {
	s := newMemoryStore()
	h, rt := newHandler(t, s, http.StatusOK)

	res := serve(t, h, http.MethodPut, "/integrations/cookie_api", `{"credential":"k3y"}`)
	require.Equal(t, http.StatusOK, res.code)
	assert.Equal(t, "https://api.cookie-api.com/api/time/current-time", rt.req.URL.String())
	assert.Equal(t, "k3y", rt.req.Header.Get("Authorization"))

	// Connecting enables the integration.
	require.Contains(t, s.credentials, "cookie_api")
	assert.NotEqual(t, "k3y", s.credentials["cookie_api"].ValueEncrypted)
	assert.True(t, s.integrations["cookie_api"].Enabled)
}

func TestConnectRejectsRefusedCredential(t *testing.T) {
	s := newMemoryStore()
	h, _ := newHandler(t, s, http.StatusUnauthorized)

	res := serve(t, h, http.MethodPut, "/integrations/cookie_api", `{"credential":"k3y"}`)
	assert.Equal(t, http.StatusBadRequest, res.code)
	assert.Equal(t, "invalid_credential", res.error["code"])
	assert.Empty(t, s.credentials)
	assert.Empty(t, s.integrations)
}

func TestConnectIntegrationWithoutCredential(t *testing.T) {
	h, _ := newHandler(t, newMemoryStore(), http.StatusOK)

	res := serve(t, h, http.MethodPut, "/integrations/discord", `{"credential":"k3y"}`)
	assert.Equal(t, http.StatusBadRequest, res.code)
	assert.Equal(t, "no_credential", res.error["code"])

	res = serve(t, h, http.MethodPut, "/integrations/nope", `{"credential":"k3y"}`)
	assert.Equal(t, http.StatusNotFound, res.code)
}

func TestUpdateIntegration(t *testing.T) {
	s := newMemoryStore()
	h, _ := newHandler(t, s, http.StatusOK)

	res := serve(t, h, http.MethodPatch, "/integrations/roblox", `{"enabled":false}`)
	require.Equal(t, http.StatusOK, res.code)
	assert.False(t, s.integrations["roblox"].Enabled)

	// A body without the value doesn't disable the integration.
	res = serve(t, h, http.MethodPatch, "/integrations/roblox", `{}`)
	assert.Equal(t, http.StatusBadRequest, res.code)

	res = serve(t, h, http.MethodPatch, "/integrations/discord", `{"enabled":false}`)
	assert.Equal(t, http.StatusBadRequest, res.code)
	assert.Equal(t, "always_enabled", res.error["code"])

	// Integrations that need a credential are enabled by connecting them.
	res = serve(t, h, http.MethodPatch, "/integrations/cookie_api", `{"enabled":true}`)
	assert.Equal(t, http.StatusBadRequest, res.code)
	assert.Equal(t, "not_connected", res.error["code"])
}

// Disabling an integration keeps its credential.
func TestDisableConnectedIntegration(t *testing.T) {
	s := newMemoryStore()
	h, _ := newHandler(t, s, http.StatusOK)
	require.Equal(t, http.StatusOK, serve(t, h, http.MethodPut, "/integrations/cookie_api", `{"credential":"k3y"}`).code)

	res := serve(t, h, http.MethodPatch, "/integrations/cookie_api", `{"enabled":false}`)
	require.Equal(t, http.StatusOK, res.code)
	assert.False(t, s.integrations["cookie_api"].Enabled)
	assert.Contains(t, s.credentials, "cookie_api")

	// Replacing the credential doesn't enable it again, and keeps the row.
	id := s.integrations["cookie_api"].ID
	res = serve(t, h, http.MethodPut, "/integrations/cookie_api", `{"credential":"n3w"}`)
	require.Equal(t, http.StatusOK, res.code)
	assert.False(t, s.integrations["cookie_api"].Enabled)
	assert.Equal(t, id, s.integrations["cookie_api"].ID)
	assert.Equal(t, id, s.credentials["cookie_api"].AppIntegrationID)
}

func TestRemoveIntegration(t *testing.T) {
	s := newMemoryStore()
	h, _ := newHandler(t, s, http.StatusOK)
	require.Equal(t, http.StatusOK, serve(t, h, http.MethodPut, "/integrations/cookie_api", `{"credential":"k3y"}`).code)

	res := serve(t, h, http.MethodDelete, "/integrations/cookie_api", "")
	require.Equal(t, http.StatusOK, res.code)
	assert.Empty(t, s.credentials)
	assert.Empty(t, s.integrations)

	res = serve(t, h, http.MethodDelete, "/integrations/cookie_api", "")
	assert.Equal(t, http.StatusNotFound, res.code)
	assert.Equal(t, "not_set_up", res.error["code"])
}

func TestListIntegrations(t *testing.T) {
	s := newMemoryStore()
	h, _ := newHandler(t, s, http.StatusOK)
	require.Equal(t, http.StatusOK, serve(t, h, http.MethodPut, "/integrations/cookie_api", `{"credential":"k3y"}`).code)
	require.Equal(t, http.StatusOK, serve(t, h, http.MethodPatch, "/integrations/roblox", `{"enabled":false}`).code)

	res := serve(t, h, http.MethodGet, "/integrations", "")
	require.Equal(t, http.StatusOK, res.code)

	enabled := map[string]bool{}
	for _, item := range res.data.([]any) {
		entry := item.(map[string]any)
		enabled[entry["integration_id"].(string)] = entry["enabled"].(bool)
	}
	assert.Equal(t, map[string]bool{"discord": true, "roblox": false, "cookie_api": true}, enabled)
}
