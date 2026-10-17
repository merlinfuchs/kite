package flowtest

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

// secretStore only implements what the handler uses.
type secretStore struct {
	store.AppSecretStore
	secrets []*model.AppSecret
}

func (s *secretStore) AppSecretsByNames(ctx context.Context, appID string, names []string) ([]*model.AppSecret, error) {
	var res []*model.AppSecret
	for _, secret := range s.secrets {
		for _, name := range names {
			if secret.AppID == appID && secret.Name == name {
				res = append(res, secret)
			}
		}
	}
	return res, nil
}

func request(t *testing.T, h *FlowTestHandler, body string) (int, map[string]any) {
	t.Helper()

	mux := http.NewServeMux()
	mux.Handle("POST /test", handler.APIHandler(func(c *handler.Context) error {
		c.App = &model.App{ID: "app"}
		return handler.TypedWithBody(h.HandleHTTPRequestTest)(c)
	}))

	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	var res map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	return rec.Code, res
}

// Without an egress proxy a test could read the host's internal network, so
// the instance has to opt in by configuring one.
func TestHTTPRequestTestNeedsProxy(t *testing.T) {
	h := NewFlowTestHandler(nil, &secretStore{}, nil)

	code, res := request(t, h, `{"data": {"url": "http://127.0.0.1"}}`)
	assert.Equal(t, http.StatusForbidden, code)
	assert.Equal(t, "http_request_test_disabled", res["error"].(map[string]any)["code"])
}

func TestHTTPRequestTest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"auth": "` + r.Header.Get("Authorization") + `", "body": ` + string(body) + `}`))
	}))
	t.Cleanup(srv.Close)

	crypt, err := util.NewSymmetricCrypt(strings.Repeat("ab", 32))
	require.NoError(t, err)
	encrypted, err := crypt.EncryptString("s3cr3t-value")
	require.NoError(t, err)

	secrets := &secretStore{secrets: []*model.AppSecret{
		{AppID: "app", Name: "API_KEY", ValueEncrypted: encrypted},
		// Another app's secret with the same name isn't used.
		{AppID: "other", Name: "API_KEY", ValueEncrypted: encrypted},
	}}
	h := NewFlowTestHandler(srv.Client(), secrets, crypt)

	code, res := request(t, h, `{
		"data": {
			"url": "`+srv.URL+`",
			"method": "POST",
			"headers": [{"key": "Authorization", "value": "Bearer {{secrets.API_KEY}}"}],
			"body_type": "json",
			"body": "{\"n\": {{arg('n')}}}",
			"response_transform": "response.data().body.n"
		},
		"values": {"arg('n')": "5"}
	}`)
	require.Equal(t, http.StatusOK, code, res)

	data := res["data"].(map[string]any)
	assert.Empty(t, data["error"])
	assert.Equal(t, "POST", data["method"])
	assert.Equal(t, `{"n": 5}`, data["request_body"])
	assert.Equal(t, 5.0, data["result"])

	response := data["response"].(map[string]any)
	assert.Equal(t, 200.0, response["status_code"])
	assert.Contains(t, response["body"], "Bearer [secret]")

	raw, _ := json.Marshal(res)
	assert.NotContains(t, string(raw), "s3cr3t")
}

func TestHTTPRequestTestValidation(t *testing.T) {
	h := NewFlowTestHandler(http.DefaultClient, &secretStore{}, nil)

	code, _ := request(t, h, `{"values": {}}`)
	assert.Equal(t, http.StatusBadRequest, code, "the request is required")

	values := map[string]string{}
	for i := 0; i < 51; i++ {
		values[strings.Repeat("a", i+1)] = "x"
	}
	body, _ := json.Marshal(map[string]any{"data": map[string]any{"url": "http://example.com"}, "values": values})
	code, _ = request(t, h, string(body))
	assert.Equal(t, http.StatusBadRequest, code, "too many test values")
}
