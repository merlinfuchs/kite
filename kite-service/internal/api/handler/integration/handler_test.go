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

type credentialStore struct {
	store.AppSecretStore
	saved *model.AppSecret
}

func (s *credentialStore) SetAppIntegrationCredential(ctx context.Context, secret *model.AppSecret) (*model.AppSecret, error) {
	s.saved = secret
	return secret, nil
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

func connect(t *testing.T, integrationID string, status int) (int, map[string]any, *credentialStore, *roundTripper) {
	crypt, err := util.NewSymmetricCrypt(strings.Repeat("ab", 32))
	require.NoError(t, err)
	s := &credentialStore{}
	rt := &roundTripper{status: status}
	h := NewIntegrationHandler(s, nil, crypt)
	h.client.Transport = rt

	mux := http.NewServeMux()
	mux.Handle("PUT /integrations/{integrationID}", handler.APIHandler(func(c *handler.Context) error {
		c.App = &model.App{ID: "app"}
		return handler.TypedWithBody(h.HandleAppIntegrationConnect)(c)
	}))

	req := httptest.NewRequest(http.MethodPut, "/integrations/"+integrationID, strings.NewReader(`{"credential":"k3y"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	var res map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	return rec.Code, res, s, rt
}

func TestConnectChecksCredential(t *testing.T) {
	code, _, s, rt := connect(t, "cookie_api", http.StatusOK)
	require.Equal(t, http.StatusOK, code)
	assert.Equal(t, "https://api.cookie-api.com/api/time/current-time", rt.req.URL.String())
	assert.Equal(t, "k3y", rt.req.Header.Get("Authorization"))
	require.NotNil(t, s.saved)
	assert.Equal(t, "cookie_api", s.saved.IntegrationID)
	assert.NotEqual(t, "k3y", s.saved.ValueEncrypted)
}

func TestConnectRejectsRefusedCredential(t *testing.T) {
	code, res, s, _ := connect(t, "cookie_api", http.StatusUnauthorized)
	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, "invalid_credential", res["error"].(map[string]any)["code"])
	assert.Nil(t, s.saved)
}

func TestConnectIntegrationWithoutCredential(t *testing.T) {
	code, res, _, _ := connect(t, "discord", http.StatusOK)
	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, "no_credential", res["error"].(map[string]any)["code"])

	code, _, _, _ = connect(t, "nope", http.StatusOK)
	assert.Equal(t, http.StatusNotFound, code)
}

type choiceStore struct {
	saved []*model.AppIntegration
}

func (s *choiceStore) AppIntegrations(ctx context.Context, appID string) ([]*model.AppIntegration, error) {
	return s.saved, nil
}

func (s *choiceStore) SetAppIntegrationEnabled(ctx context.Context, integration *model.AppIntegration) (*model.AppIntegration, error) {
	s.saved = append(s.saved, integration)
	return integration, nil
}

type listCredentialStore struct {
	store.AppSecretStore
}

func (s *listCredentialStore) AppIntegrationCredentials(ctx context.Context, appID string) ([]*model.AppSecret, error) {
	return []*model.AppSecret{{IntegrationID: "cookie_api"}}, nil
}

func serve(t *testing.T, h *IntegrationHandler, method string, path string, body string) (int, map[string]any, []any) {
	mux := http.NewServeMux()
	mux.Handle("GET /integrations", handler.APIHandler(func(c *handler.Context) error {
		c.App = &model.App{ID: "app"}
		return handler.Typed(h.HandleAppIntegrationList)(c)
	}))
	mux.Handle("PATCH /integrations/{integrationID}", handler.APIHandler(func(c *handler.Context) error {
		c.App = &model.App{ID: "app"}
		return handler.TypedWithBody(h.HandleAppIntegrationUpdate)(c)
	}))

	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	var res struct {
		Data  any            `json:"data"`
		Error map[string]any `json:"error"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	list, _ := res.Data.([]any)
	return rec.Code, res.Error, list
}

func TestUpdateIntegration(t *testing.T) {
	choices := &choiceStore{}
	h := NewIntegrationHandler(nil, choices, nil)

	code, _, _ := serve(t, h, http.MethodPatch, "/integrations/roblox", `{"enabled":false}`)
	require.Equal(t, http.StatusOK, code)
	require.Len(t, choices.saved, 1)
	assert.Equal(t, "roblox", choices.saved[0].IntegrationID)
	assert.False(t, choices.saved[0].Enabled)

	// A body without the value doesn't turn the integration off.
	code, _, _ = serve(t, h, http.MethodPatch, "/integrations/roblox", `{}`)
	assert.Equal(t, http.StatusBadRequest, code)
	require.Len(t, choices.saved, 1)

	code, res, _ := serve(t, h, http.MethodPatch, "/integrations/discord", `{"enabled":false}`)
	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, "always_enabled", res["code"])

	code, res, _ = serve(t, h, http.MethodPatch, "/integrations/cookie_api", `{"enabled":true}`)
	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, "needs_credential", res["code"])
}

func TestListIntegrations(t *testing.T) {
	choices := &choiceStore{saved: []*model.AppIntegration{{IntegrationID: "roblox", Enabled: false}}}
	h := NewIntegrationHandler(&listCredentialStore{}, choices, nil)

	code, _, list := serve(t, h, http.MethodGet, "/integrations", "")
	require.Equal(t, http.StatusOK, code)

	enabled := map[string]bool{}
	for _, item := range list {
		entry := item.(map[string]any)
		enabled[entry["integration_id"].(string)] = entry["enabled"].(bool)
	}
	assert.Equal(t, map[string]bool{"discord": true, "roblox": false, "cookie_api": true}, enabled)
}
