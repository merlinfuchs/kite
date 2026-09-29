package appsecret

import (
	"context"
	"encoding/json"
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

type memoryStore struct {
	secrets []*model.AppSecret
}

func (s *memoryStore) AppSecretsByApp(ctx context.Context, appID string) ([]*model.AppSecret, error) {
	var res []*model.AppSecret
	for _, secret := range s.secrets {
		if secret.AppID == appID {
			res = append(res, secret)
		}
	}
	return res, nil
}

func (s *memoryStore) AppSecret(ctx context.Context, appID string, id string) (*model.AppSecret, error) {
	for _, secret := range s.secrets {
		if secret.AppID == appID && secret.ID == id {
			copied := *secret
			return &copied, nil
		}
	}
	return nil, store.ErrNotFound
}

func (s *memoryStore) AppSecretsByNames(ctx context.Context, appID string, names []string) ([]*model.AppSecret, error) {
	return nil, nil
}

func (s *memoryStore) CountAppSecretsByApp(ctx context.Context, appID string) (int, error) {
	secrets, _ := s.AppSecretsByApp(ctx, appID)
	return len(secrets), nil
}

func (s *memoryStore) CreateAppSecret(ctx context.Context, secret *model.AppSecret) (*model.AppSecret, error) {
	for _, other := range s.secrets {
		if other.AppID == secret.AppID && other.Name == secret.Name {
			return nil, store.ErrAlreadyExists
		}
	}
	s.secrets = append(s.secrets, secret)
	return secret, nil
}

func (s *memoryStore) UpdateAppSecret(ctx context.Context, secret *model.AppSecret) (*model.AppSecret, error) {
	for i, other := range s.secrets {
		if other.AppID == secret.AppID && other.ID == secret.ID {
			s.secrets[i] = secret
			return secret, nil
		}
	}
	return nil, store.ErrNotFound
}

func (s *memoryStore) DeleteAppSecret(ctx context.Context, appID string, id string) error {
	return nil
}

type testSetup struct {
	store   *memoryStore
	crypt   *util.SymmetricCrypt
	handler *AppSecretHandler
}

func newTestSetup(t *testing.T) *testSetup {
	crypt, err := util.NewSymmetricCrypt(strings.Repeat("ab", 32))
	require.NoError(t, err)
	s := &memoryStore{}
	return &testSetup{store: s, crypt: crypt, handler: NewAppSecretHandler(s, crypt)}
}

func (s *testSetup) request(t *testing.T, maxSecrets int, method string, path string, body string, h handler.HandlerFunc) (int, map[string]any) {
	t.Helper()

	mux := http.NewServeMux()
	mux.Handle(method+" /secrets/{secretID}", handler.APIHandler(func(c *handler.Context) error {
		c.App = &model.App{ID: "app"}
		c.Features = model.Features{MaxSecrets: maxSecrets}
		return h(c)
	}))
	mux.Handle(method+" /secrets", handler.APIHandler(func(c *handler.Context) error {
		c.App = &model.App{ID: "app"}
		c.Features = model.Features{MaxSecrets: maxSecrets}
		return h(c)
	}))

	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	var res map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	return rec.Code, res
}

func TestCreateSecretEncryptsAndNeverReturnsValue(t *testing.T) {
	s := newTestSetup(t)

	code, res := s.request(t, 0, http.MethodPost, "/secrets", `{"name":"API_KEY","value":"s3cr3t"}`,
		handler.TypedWithBody(s.handler.HandleAppSecretCreate))
	require.Equal(t, http.StatusOK, code, res)
	assert.NotContains(t, mustJSON(res), "s3cr3t")

	require.Len(t, s.store.secrets, 1)
	assert.NotEqual(t, "s3cr3t", s.store.secrets[0].ValueEncrypted)
	value, err := s.crypt.DecryptString(s.store.secrets[0].ValueEncrypted)
	require.NoError(t, err)
	assert.Equal(t, "s3cr3t", value)
}

func TestCreateSecretLimits(t *testing.T) {
	s := newTestSetup(t)
	create := handler.TypedWithBody(s.handler.HandleAppSecretCreate)

	code, _ := s.request(t, 1, http.MethodPost, "/secrets", `{"name":"A","value":"1"}`, create)
	require.Equal(t, http.StatusOK, code)

	// The plan's cap.
	code, res := s.request(t, 1, http.MethodPost, "/secrets", `{"name":"B","value":"1"}`, create)
	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, "resource_limit", res["error"].(map[string]any)["code"])

	// Names are unique per app.
	code, res = s.request(t, 0, http.MethodPost, "/secrets", `{"name":"A","value":"1"}`, create)
	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, "duplicate_name", res["error"].(map[string]any)["code"])

	// Values are at most 4 KB.
	code, _ = s.request(t, 0, http.MethodPost, "/secrets", `{"name":"C","value":"`+strings.Repeat("a", 4097)+`"}`, create)
	assert.Equal(t, http.StatusBadRequest, code)
}

func TestUpdateSecretKeepsValueWhenNoneGiven(t *testing.T) {
	s := newTestSetup(t)
	code, res := s.request(t, 0, http.MethodPost, "/secrets", `{"name":"API_KEY","value":"old"}`,
		handler.TypedWithBody(s.handler.HandleAppSecretCreate))
	require.Equal(t, http.StatusOK, code)
	id := res["data"].(map[string]any)["id"].(string)
	encrypted := s.store.secrets[0].ValueEncrypted

	code, _ = s.request(t, 0, http.MethodPatch, "/secrets/"+id, `{"name":"RENAMED","value":null}`,
		handler.TypedWithBody(s.handler.HandleAppSecretUpdate))
	require.Equal(t, http.StatusOK, code)
	assert.Equal(t, "RENAMED", s.store.secrets[0].Name)
	assert.Equal(t, encrypted, s.store.secrets[0].ValueEncrypted)

	code, _ = s.request(t, 0, http.MethodPatch, "/secrets/"+id, `{"name":"RENAMED","value":"new"}`,
		handler.TypedWithBody(s.handler.HandleAppSecretUpdate))
	require.Equal(t, http.StatusOK, code)
	value, err := s.crypt.DecryptString(s.store.secrets[0].ValueEncrypted)
	require.NoError(t, err)
	assert.Equal(t, "new", value)
}

func mustJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
