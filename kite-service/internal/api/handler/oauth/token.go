package oauth

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/handler/apitoken"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/internal/util"
	"gopkg.in/guregu/null.v4"
)

const maxTokenRequestSize = 16 * 1024

// codeVerifierPattern is a PKCE code verifier, see RFC 7636.
var codeVerifierPattern = regexp.MustCompile(`^[A-Za-z0-9\-._~]{43,128}$`)

// tokenError is an error response of the token endpoint, see RFC 6749. All
// of them are 400s, as clients don't authenticate.
type tokenError struct {
	Code        string `json:"error"`
	Description string `json:"error_description"`
}

func (e *tokenError) Error() string {
	return e.Code + ": " + e.Description
}

func errToken(code string, description string) *tokenError {
	return &tokenError{Code: code, Description: description}
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

func newTokenResponse(token *model.APIToken, accessToken string, refreshToken string) *tokenResponse {
	return &tokenResponse{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(accessTokenExpiry.Seconds()),
		RefreshToken: refreshToken,
		Scope:        strings.Join(token.Scopes, " "),
	}
}

// HandleOAuthToken is the token endpoint, where clients exchange codes and
// refresh tokens for API tokens. It answers like RFC 6749 says, not like the
// rest of the API.
func (h *OAuthHandler) HandleOAuthToken(c *handler.Context) error {
	c.SetHeader("Cache-Control", "no-store")

	resp, err := h.token(c)
	if err != nil {
		var terr *tokenError
		if errors.As(err, &terr) {
			return c.JSON(http.StatusBadRequest, terr)
		}
		return err
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *OAuthHandler) token(c *handler.Context) (*tokenResponse, error) {
	form, err := c.ParseForm(maxTokenRequestSize)
	if err != nil {
		var herr *handler.Error
		if errors.As(err, &herr) {
			return nil, errToken("invalid_request", herr.Message)
		}
		return nil, err
	}

	if form.Get("client_id") == "" {
		form.Set("client_id", assertionClientID(form))
	}

	switch form.Get("grant_type") {
	case "authorization_code":
		return h.exchangeCode(c, form)
	case "refresh_token":
		return h.refresh(c, form)
	default:
		return nil, errToken("unsupported_grant_type", "grant_type must be authorization_code or refresh_token")
	}
}

func (h *OAuthHandler) exchangeCode(c *handler.Context, form url.Values) (*tokenResponse, error) {
	for _, param := range []string{"code", "client_id", "redirect_uri", "code_verifier"} {
		if form.Get(param) == "" {
			return nil, errToken("invalid_request", param+" is missing")
		}
	}

	now := time.Now().UTC()
	codeHash := util.HashKey(form.Get("code"))
	tokenID := util.UniqueID()

	// The code is used up even if the checks below fail, so it can't be
	// guessed with.
	code, err := h.codeStore.UseOAuthCode(c.Context(), codeHash, tokenID, now)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			h.revokeReusedCode(c, codeHash)
			return nil, errToken("invalid_grant", "the code is invalid, expired or was already used")
		}
		return nil, fmt.Errorf("failed to use OAuth code: %w", err)
	}

	if code.ClientID != form.Get("client_id") {
		return nil, errToken("invalid_grant", "the code is for a different client")
	}
	if code.RedirectURI != form.Get("redirect_uri") {
		return nil, errToken("invalid_grant", "redirect_uri isn't the one of the authorization request")
	}
	if !verifyCodeChallenge(form.Get("code_verifier"), code.CodeChallenge) {
		return nil, errToken("invalid_grant", "code_verifier doesn't match the code_challenge")
	}
	if resource := form.Get("resource"); resource != "" {
		if r, ok := h.resource(resource); !ok || r != code.Resource {
			return nil, errToken("invalid_target", "resource isn't the one of the authorization request")
		}
	}

	// Checked again, as the user can approve more clients before they get
	// their tokens.
	if err := apitoken.CheckTokenLimit(c.Context(), h.apiTokenStore, code.UserID, h.config.MaxAPITokensPerUser, now); err != nil {
		var herr *handler.Error
		if errors.As(err, &herr) {
			return nil, errToken("invalid_grant", herr.Message)
		}
		return nil, err
	}

	accessToken, accessKeyHash := apitoken.NewKey()
	refreshToken, refreshKeyHash := apitoken.NewKey()
	token, err := h.apiTokenStore.CreateAPIToken(c.Context(), &model.APIToken{
		ID:        tokenID,
		UserID:    code.UserID,
		Name:      code.ClientName,
		KeyHash:   accessKeyHash,
		Scopes:    code.Scopes,
		AppIDs:    code.AppIDs,
		ExpiresAt: null.TimeFrom(now.Add(grantExpiry)),
		CreatedAt: now,

		OAuthClientID:   null.StringFrom(code.ClientID),
		OAuthResource:   null.StringFrom(code.Resource),
		RefreshKeyHash:  null.StringFrom(refreshKeyHash),
		AccessExpiresAt: null.TimeFrom(now.Add(accessTokenExpiry)),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create API token: %w", err)
	}

	return newTokenResponse(token, accessToken, refreshToken), nil
}

// revokeReusedCode deletes the token a code was exchanged for if it's used
// again, as either the first use or this one wasn't the client. See RFC 6749
// section 4.1.2.
func (h *OAuthHandler) revokeReusedCode(c *handler.Context, codeHash string) {
	code, err := h.codeStore.OAuthCode(c.Context(), codeHash)
	if err != nil || !code.APITokenID.Valid {
		return
	}

	slog.Warn(
		"OAuth code used again, revoking its token",
		slog.String("user_id", code.UserID),
		slog.String("client_id", code.ClientID),
	)
	err = h.apiTokenStore.DeleteAPIToken(c.Context(), code.UserID, code.APITokenID.String)
	if err != nil && !errors.Is(err, store.ErrNotFound) {
		slog.Error(
			"Failed to revoke token of reused OAuth code",
			slog.String("api_token_id", code.APITokenID.String),
			slog.String("error", err.Error()),
		)
	}
}

// refresh gives the client new keys for its token. A refresh token that was
// used before isn't treated as stolen, as clients like Claude Code can refresh
// from several processes at once, and that would disconnect them.
func (h *OAuthHandler) refresh(c *handler.Context, form url.Values) (*tokenResponse, error) {
	for _, param := range []string{"refresh_token", "client_id"} {
		if form.Get(param) == "" {
			return nil, errToken("invalid_request", param+" is missing")
		}
	}

	now := time.Now().UTC()
	accessToken, accessKeyHash := apitoken.NewKey()
	refreshToken, refreshKeyHash := apitoken.NewKey()
	token, err := h.apiTokenStore.RefreshAPIToken(c.Context(), store.APITokenRefresh{
		ClientID:          form.Get("client_id"),
		RefreshKeyHash:    util.HashKey(form.Get("refresh_token")),
		NewKeyHash:        accessKeyHash,
		NewRefreshKeyHash: refreshKeyHash,
		AccessExpiresAt:   now.Add(accessTokenExpiry),
		ExpiresAt:         now.Add(grantExpiry),
		Now:               now,
	})
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, errToken("invalid_grant", "the refresh token is invalid, expired or was revoked")
		}
		return nil, fmt.Errorf("failed to refresh API token: %w", err)
	}

	return newTokenResponse(token, accessToken, refreshToken), nil
}

func verifyCodeChallenge(verifier string, challenge string) bool {
	if !codeVerifierPattern.MatchString(verifier) {
		return false
	}
	hash := sha256.Sum256([]byte(verifier))
	expected := base64.RawURLEncoding.EncodeToString(hash[:])
	return subtle.ConstantTimeCompare([]byte(expected), []byte(challenge)) == 1
}

// assertionClientID is the client of a request that signs in with a JWT
// instead of sending its client_id, like ChatGPT can (RFC 7523). Kite treats
// all clients as public, so the JWT isn't checked, like a client_id wouldn't
// be. The code and PKCE prove who the client is.
func assertionClientID(form url.Values) string {
	if form.Get("client_assertion_type") != "urn:ietf:params:oauth:client-assertion-type:jwt-bearer" {
		return ""
	}
	parts := strings.Split(form.Get("client_assertion"), ".")
	if len(parts) != 3 {
		return ""
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return ""
	}
	var claims struct {
		Subject string `json:"sub"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return ""
	}
	return claims.Subject
}
