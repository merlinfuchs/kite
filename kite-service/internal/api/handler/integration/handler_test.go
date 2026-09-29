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
	h := NewIntegrationHandler(s, crypt)
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

func TestConnectAlwaysConnectedIntegration(t *testing.T) {
	code, res, _, _ := connect(t, "discord", http.StatusOK)
	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, "always_connected", res["error"].(map[string]any)["code"])

	code, _, _, _ = connect(t, "nope", http.StatusOK)
	assert.Equal(t, http.StatusNotFound, code)
}
