package oauth

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"
)

type memoryTokenStore struct {
	// Methods the tests don't use panic.
	store.APITokenStore

	tokens []*model.APIToken
}

func (s *memoryTokenStore) CreateAPIToken(ctx context.Context, token *model.APIToken) (*model.APIToken, error) {
	copied := *token
	s.tokens = append(s.tokens, &copied)
	return token, nil
}

func (s *memoryTokenStore) CountAPITokensByUser(ctx context.Context, userID string, now time.Time) (int, error) {
	count := 0
	for _, token := range s.tokens {
		if token.UserID == userID && (!token.ExpiresAt.Valid || token.ExpiresAt.Time.After(now)) {
			count++
		}
	}
	return count, nil
}

func (s *memoryTokenStore) DeleteAPIToken(ctx context.Context, userID string, id string) error {
	for i, token := range s.tokens {
		if token.UserID == userID && token.ID == id {
			s.tokens = append(s.tokens[:i], s.tokens[i+1:]...)
			return nil
		}
	}
	return store.ErrNotFound
}

func (s *memoryTokenStore) RefreshAPIToken(ctx context.Context, refresh store.APITokenRefresh) (*model.APIToken, error) {
	for _, token := range s.tokens {
		if token.RefreshKeyHash.String == refresh.RefreshKeyHash && token.OAuthClientID.String == refresh.ClientID && token.ExpiresAt.Time.After(refresh.Now) {
			token.KeyHash = refresh.NewKeyHash
			token.RefreshKeyHash = null.StringFrom(refresh.NewRefreshKeyHash)
			token.AccessExpiresAt = null.TimeFrom(refresh.AccessExpiresAt)
			token.ExpiresAt = null.TimeFrom(refresh.ExpiresAt)
			copied := *token
			return &copied, nil
		}
	}
	return nil, store.ErrNotFound
}

func (s *memoryTokenStore) byKey(key string) *model.APIToken {
	for _, token := range s.tokens {
		if token.KeyHash == util.HashKey(key) {
			return token
		}
	}
	return nil
}

type memoryCodeStore struct {
	codes map[string]*model.OAuthCode
}

func (s *memoryCodeStore) CreateOAuthCode(ctx context.Context, code *model.OAuthCode) error {
	copied := *code
	s.codes[code.CodeHash] = &copied
	return nil
}

func (s *memoryCodeStore) UseOAuthCode(ctx context.Context, codeHash string, apiTokenID string, now time.Time) (*model.OAuthCode, error) {
	code, ok := s.codes[codeHash]
	if !ok || code.APITokenID.Valid || !code.ExpiresAt.After(now) {
		return nil, store.ErrNotFound
	}
	code.APITokenID = null.StringFrom(apiTokenID)
	copied := *code
	return &copied, nil
}

func (s *memoryCodeStore) OAuthCode(ctx context.Context, codeHash string) (*model.OAuthCode, error) {
	code, ok := s.codes[codeHash]
	if !ok {
		return nil, store.ErrNotFound
	}
	copied := *code
	return &copied, nil
}

func (s *memoryCodeStore) DeleteExpiredOAuthCodes(ctx context.Context, expiredBefore time.Time, batchSize int) (int64, error) {
	return 0, nil
}

type memoryAppStore struct {
	store.AppStore
}

func (s *memoryAppStore) AppsByUser(ctx context.Context, userID string) ([]*model.App, error) {
	return []*model.App{{ID: "app"}}, nil
}

type memoryClients map[string]*Client

func (m memoryClients) Client(ctx context.Context, id string) (*Client, error) {
	client, ok := m[id]
	if !ok {
		return nil, invalidClient("failed to fetch the client's metadata document")
	}
	return client, nil
}

const (
	clientID    = "https://claude.ai/oauth/mcp-oauth-client-metadata"
	redirectURI = "https://claude.ai/api/mcp/auth_callback"
	resource    = "https://kite.onl/api/mcp"
	verifier    = "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
)

func challenge(verifier string) string {
	hash := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

type testSetup struct {
	tokens  *memoryTokenStore
	codes   *memoryCodeStore
	handler *OAuthHandler
}

func newTestSetup() *testSetup {
	tokens := &memoryTokenStore{}
	codes := &memoryCodeStore{codes: map[string]*model.OAuthCode{}}
	clients := memoryClients{clientID: {
		ID:           clientID,
		Name:         "Claude",
		Domain:       "claude.ai",
		RedirectURIs: []string{redirectURI},
	}}
	return &testSetup{
		tokens: tokens,
		codes:  codes,
		handler: NewOAuthHandler(OAuthHandlerConfig{
			Issuer:                "https://api.kite.onl/",
			AuthorizationEndpoint: "https://kite.onl/oauth/authorize",
			Resources:             []string{resource},
			MaxAPITokensPerUser:   2,
		}, clients, tokens, codes, &memoryAppStore{}),
	}
}

func authorizeParams() map[string]string {
	return map[string]string{
		"response_type":         "code",
		"client_id":             clientID,
		"redirect_uri":          redirectURI,
		"state":                 "some-state",
		"code_challenge":        challenge(verifier),
		"code_challenge_method": "S256",
		"resource":              resource,
	}
}

// request sends a request to the route as the logged in user, and returns
// the status and the response.
func request(t *testing.T, h handler.HandlerFunc, req *http.Request) (int, map[string]any) {
	t.Helper()

	rec := httptest.NewRecorder()
	handler.APIHandler(func(c *handler.Context) error {
		c.Session = &model.Session{UserID: "user"}
		return h(c)
	}).ServeHTTP(rec, req)

	var res map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res), rec.Body.String())
	return rec.Code, res
}

func errorCode(res map[string]any) string {
	if e, ok := res["error"].(map[string]any); ok {
		return e["code"].(string)
	}
	return ""
}

func authorizeQuery(params map[string]string) string {
	query := url.Values{}
	for key, value := range params {
		query.Set(key, value)
	}
	return query.Encode()
}

func (s *testSetup) authorizeGet(t *testing.T, params map[string]string) (int, map[string]any) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/v1/oauth/authorize?"+authorizeQuery(params), nil)
	return request(t, handler.Typed(s.handler.HandleOAuthAuthorizeGet), req)
}

func (s *testSetup) authorize(t *testing.T, params map[string]string, approved bool, scopes []string) (int, map[string]any) {
	t.Helper()
	body, err := json.Marshal(map[string]any{
		"approved": approved,
		"scopes":   scopes,
		"app_ids":  nil,
	})
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/v1/oauth/authorize?"+authorizeQuery(params), strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	return request(t, handler.TypedWithBody(s.handler.HandleOAuthAuthorize), req)
}

// code approves the client and returns the code it gets.
func (s *testSetup) code(t *testing.T) string {
	t.Helper()
	status, res := s.authorize(t, authorizeParams(), true, []string{model.APITokenScopeKiteWrite})
	require.Equal(t, http.StatusOK, status, res)
	redirect, err := url.Parse(res["data"].(map[string]any)["redirect_url"].(string))
	require.NoError(t, err)
	return redirect.Query().Get("code")
}

func (s *testSetup) token(t *testing.T, form url.Values) (int, map[string]any) {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/v1/oauth/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return request(t, s.handler.HandleOAuthToken, req)
}

func exchangeForm(code string) url.Values {
	return url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"client_id":     {clientID},
		"redirect_uri":  {redirectURI},
		"code_verifier": {verifier},
		"resource":      {resource},
	}
}

func TestMetadata(t *testing.T) {
	s := newTestSetup()
	status, res := request(t, s.handler.HandleMetadata, httptest.NewRequest(http.MethodGet, "/.well-known/oauth-authorization-server", nil))
	require.Equal(t, http.StatusOK, status)

	assert.Equal(t, "https://api.kite.onl", res["issuer"])
	assert.Equal(t, "https://kite.onl/oauth/authorize", res["authorization_endpoint"])
	assert.Equal(t, "https://api.kite.onl/v1/oauth/token", res["token_endpoint"])
	assert.Equal(t, true, res["client_id_metadata_document_supported"])
	assert.Equal(t, []any{"none"}, res["token_endpoint_auth_methods_supported"])
	assert.Equal(t, []any{"S256"}, res["code_challenge_methods_supported"])
	assert.NotContains(t, res, "registration_endpoint")
}

func TestAuthorizeGet(t *testing.T) {
	s := newTestSetup()

	status, res := s.authorizeGet(t, authorizeParams())
	require.Equal(t, http.StatusOK, status, res)
	data := res["data"].(map[string]any)
	assert.Equal(t, map[string]any{"id": clientID, "name": "Claude", "domain": "claude.ai"}, data["client"])
	// All scopes, as it asked for none.
	assert.Len(t, data["scopes"], len(model.APITokenScopes))

	params := authorizeParams()
	params["scope"] = "kite:read offline_access discord:read"
	params["resource"] = resource + "/"
	status, res = s.authorizeGet(t, params)
	require.Equal(t, http.StatusOK, status, res)
	assert.Equal(t, []any{"kite:read", "discord:read"}, res["data"].(map[string]any)["scopes"])

	// Write includes read.
	params["scope"] = "kite:write discord:write"
	status, res = s.authorizeGet(t, params)
	require.Equal(t, http.StatusOK, status, res)
	assert.Equal(t, []any{"kite:read", "kite:write", "discord:read", "discord:write"}, res["data"].(map[string]any)["scopes"])

	tests := []struct {
		param string
		value string
		code  string
	}{
		{"client_id", "https://evil.example/client", "invalid_client"},
		{"redirect_uri", "https://evil.example/callback", "invalid_redirect_uri"},
		{"response_type", "token", "unsupported_response_type"},
		{"code_challenge_method", "plain", "invalid_request"},
		{"code_challenge_method", "", "invalid_request"},
		{"code_challenge", "short", "invalid_request"},
		{"state", strings.Repeat("a", maxStateLength+1), "invalid_request"},
		{"resource", "https://evil.example/mcp", "invalid_target"},
	}
	for _, tt := range tests {
		params := authorizeParams()
		params[tt.param] = tt.value
		status, res := s.authorizeGet(t, params)
		assert.Equal(t, http.StatusBadRequest, status, tt.param)
		assert.Equal(t, tt.code, errorCode(res), tt.param)
	}
}

func TestAuthorize(t *testing.T) {
	s := newTestSetup()

	status, res := s.authorize(t, authorizeParams(), true, []string{model.APITokenScopeKiteWrite})
	require.Equal(t, http.StatusOK, status, res)
	redirect, err := url.Parse(res["data"].(map[string]any)["redirect_url"].(string))
	require.NoError(t, err)
	assert.Equal(t, redirectURI, redirect.Scheme+"://"+redirect.Host+redirect.Path)
	assert.Equal(t, "some-state", redirect.Query().Get("state"))
	assert.Equal(t, "https://api.kite.onl", redirect.Query().Get("iss"))

	code := s.codes.codes[util.HashKey(redirect.Query().Get("code"))]
	require.NotNil(t, code)
	assert.Equal(t, "user", code.UserID)
	assert.Equal(t, "Claude", code.ClientName)
	assert.Equal(t, []string{model.APITokenScopeKiteWrite}, code.Scopes)
	assert.Nil(t, code.AppIDs)
	assert.Equal(t, resource, code.Resource)
}

func TestAuthorizeDenied(t *testing.T) {
	s := newTestSetup()

	status, res := s.authorize(t, authorizeParams(), false, nil)
	require.Equal(t, http.StatusOK, status, res)
	redirect, err := url.Parse(res["data"].(map[string]any)["redirect_url"].(string))
	require.NoError(t, err)
	assert.Equal(t, "access_denied", redirect.Query().Get("error"))
	assert.Equal(t, "some-state", redirect.Query().Get("state"))
	assert.Empty(t, redirect.Query().Get("code"))
	assert.Empty(t, s.codes.codes)
}

func TestAuthorizeChecks(t *testing.T) {
	s := newTestSetup()

	// The client asked for less.
	params := authorizeParams()
	params["scope"] = "kite:read"
	status, res := s.authorize(t, params, true, []string{model.APITokenScopeKiteWrite})
	assert.Equal(t, http.StatusBadRequest, status)
	assert.Equal(t, "invalid_scope", errorCode(res))

	// Like API tokens, it needs a kite scope.
	status, res = s.authorize(t, authorizeParams(), true, []string{model.APITokenScopeDiscordRead})
	assert.Equal(t, http.StatusBadRequest, status)
	assert.Equal(t, "validation_failed", errorCode(res))

	// The user's tokens count.
	s.tokens.tokens = []*model.APIToken{{ID: "a", UserID: "user"}, {ID: "b", UserID: "user"}}
	status, res = s.authorize(t, authorizeParams(), true, []string{model.APITokenScopeKiteWrite})
	assert.Equal(t, http.StatusBadRequest, status)
	assert.Equal(t, "resource_limit", errorCode(res))
	assert.Empty(t, s.codes.codes)
}

func TestExchangeCode(t *testing.T) {
	s := newTestSetup()
	code := s.code(t)

	status, res := s.token(t, exchangeForm(code))
	require.Equal(t, http.StatusOK, status, res)
	assert.Equal(t, "Bearer", res["token_type"])
	assert.Equal(t, float64(3600), res["expires_in"])
	assert.Equal(t, "kite:write", res["scope"])

	token := s.tokens.byKey(res["access_token"].(string))
	require.NotNil(t, token)
	assert.Equal(t, "user", token.UserID)
	assert.Equal(t, "Claude", token.Name)
	assert.Equal(t, clientID, token.OAuthClientID.String)
	assert.Equal(t, resource, token.OAuthResource.String)
	assert.Equal(t, util.HashKey(res["refresh_token"].(string)), token.RefreshKeyHash.String)
	assert.WithinDuration(t, time.Now().Add(accessTokenExpiry), token.AccessExpiresAt.Time, time.Minute)
	assert.WithinDuration(t, time.Now().Add(grantExpiry), token.ExpiresAt.Time, time.Minute)

	// Using it again revokes the token, as one of the uses wasn't the client.
	status, res = s.token(t, exchangeForm(code))
	assert.Equal(t, http.StatusBadRequest, status)
	assert.Equal(t, "invalid_grant", res["error"])
	assert.Empty(t, s.tokens.tokens)
}

func TestExchangeCodeChecks(t *testing.T) {
	tests := []struct {
		name  string
		param string
		value string
		error string
		// Whether the code is used up, so it doesn't work after.
		usesCode bool
	}{
		{"wrong verifier", "code_verifier", strings.Repeat("a", 43), "invalid_grant", true},
		{"short verifier", "code_verifier", "short", "invalid_grant", true},
		{"wrong client", "client_id", "https://evil.example/client", "invalid_grant", true},
		{"wrong redirect", "redirect_uri", "https://claude.ai/other", "invalid_grant", true},
		{"wrong resource", "resource", "https://evil.example/mcp", "invalid_target", true},
		{"unknown code", "code", "unknown", "invalid_grant", false},
		{"missing verifier", "code_verifier", "", "invalid_request", false},
		{"wrong grant type", "grant_type", "client_credentials", "unsupported_grant_type", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestSetup()
			code := s.code(t)
			form := exchangeForm(code)
			form.Set(tt.param, tt.value)
			status, res := s.token(t, form)
			assert.Equal(t, http.StatusBadRequest, status)
			assert.Equal(t, tt.error, res["error"])

			status, _ = s.token(t, exchangeForm(code))
			if tt.usesCode {
				assert.Equal(t, http.StatusBadRequest, status)
			} else {
				assert.Equal(t, http.StatusOK, status)
			}
		})
	}

	// Expired codes don't work.
	s := newTestSetup()
	code := s.code(t)
	s.codes.codes[util.HashKey(code)].ExpiresAt = time.Now().Add(-time.Second)
	status, res := s.token(t, exchangeForm(code))
	assert.Equal(t, http.StatusBadRequest, status)
	assert.Equal(t, "invalid_grant", res["error"])

	// Only forms.
	req := httptest.NewRequest(http.MethodPost, "/v1/oauth/token", strings.NewReader(`{"grant_type":"authorization_code"}`))
	req.Header.Set("Content-Type", "application/json")
	status, res = request(t, s.handler.HandleOAuthToken, req)
	assert.Equal(t, http.StatusBadRequest, status)
	assert.Equal(t, "invalid_request", res["error"])
}

func TestRefresh(t *testing.T) {
	s := newTestSetup()
	_, res := s.token(t, exchangeForm(s.code(t)))
	oldAccess := res["access_token"].(string)
	oldRefresh := res["refresh_token"].(string)

	refreshForm := func(refreshToken string) url.Values {
		return url.Values{
			"grant_type":    {"refresh_token"},
			"refresh_token": {refreshToken},
			"client_id":     {clientID},
		}
	}

	status, res := s.token(t, refreshForm(oldRefresh))
	require.Equal(t, http.StatusOK, status, res)
	assert.Equal(t, "kite:write", res["scope"])
	assert.NotEqual(t, oldRefresh, res["refresh_token"])
	assert.Nil(t, s.tokens.byKey(oldAccess))
	require.NotNil(t, s.tokens.byKey(res["access_token"].(string)))
	newRefresh := res["refresh_token"].(string)

	// The old refresh token is replaced.
	status, res = s.token(t, refreshForm(oldRefresh))
	assert.Equal(t, http.StatusBadRequest, status)
	assert.Equal(t, "invalid_grant", res["error"])

	// Only for its client.
	form := refreshForm(newRefresh)
	form.Set("client_id", "https://evil.example/client")
	status, res = s.token(t, form)
	assert.Equal(t, http.StatusBadRequest, status)
	assert.Equal(t, "invalid_grant", res["error"])

	// Not after the grant expired.
	s.tokens.tokens[0].ExpiresAt = null.TimeFrom(time.Now().Add(-time.Second))
	status, res = s.token(t, refreshForm(newRefresh))
	assert.Equal(t, http.StatusBadRequest, status)
	assert.Equal(t, "invalid_grant", res["error"])
}
