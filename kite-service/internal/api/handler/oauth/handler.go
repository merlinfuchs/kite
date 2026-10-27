// Package oauth is an OAuth 2.1 authorization server, so clients like Claude
// and ChatGPT can get API tokens for the MCP server without users making one.
// Clients are identified by the URL of their metadata document (CIMD) and
// need PKCE. There's no dynamic client registration.
package oauth

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/handler/apitoken"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/internal/util"
)

const (
	// MetadataPath and TokenPath are on the API, which is the issuer.
	MetadataPath = "/.well-known/oauth-authorization-server"
	TokenPath    = "/v1/oauth/token"

	codeExpiry        = time.Minute
	accessTokenExpiry = time.Hour
	// grantExpiry is how long a client stays connected without being used,
	// as every refresh pushes it back.
	grantExpiry = 30 * 24 * time.Hour

	maxStateLength = 2048
)

// scopes are the ones clients can get with OAuth. The Discord API scopes are
// only for API tokens, as the directories of Claude and ChatGPT don't allow a
// tool that can call any endpoint of another service.
var scopes = []string{model.APITokenScopeKiteRead, model.APITokenScopeKiteWrite}

type OAuthHandlerConfig struct {
	// Issuer is the URL of the API, which issues the tokens.
	Issuer string
	// AuthorizationEndpoint is the consent page on the website.
	AuthorizationEndpoint string
	// Resources are the servers tokens can be for, like the MCP server. The
	// first one is used for clients that don't say.
	Resources           []string
	MaxAPITokensPerUser int
}

type OAuthHandler struct {
	config        OAuthHandlerConfig
	clients       ClientFetcher
	apiTokenStore store.APITokenStore
	codeStore     store.OAuthCodeStore
	appStore      store.AppStore
}

func NewOAuthHandler(
	config OAuthHandlerConfig,
	clients ClientFetcher,
	apiTokenStore store.APITokenStore,
	codeStore store.OAuthCodeStore,
	appStore store.AppStore,
) *OAuthHandler {
	config.Issuer = strings.TrimSuffix(config.Issuer, "/")
	return &OAuthHandler{
		config:        config,
		clients:       clients,
		apiTokenStore: apiTokenStore,
		codeStore:     codeStore,
		appStore:      appStore,
	}
}

// metadata is the authorization server metadata of RFC 8414.
type metadata struct {
	Issuer                                     string   `json:"issuer"`
	AuthorizationEndpoint                      string   `json:"authorization_endpoint"`
	TokenEndpoint                              string   `json:"token_endpoint"`
	ScopesSupported                            []string `json:"scopes_supported"`
	ResponseTypesSupported                     []string `json:"response_types_supported"`
	GrantTypesSupported                        []string `json:"grant_types_supported"`
	TokenEndpointAuthMethodsSupported          []string `json:"token_endpoint_auth_methods_supported"`
	CodeChallengeMethodsSupported              []string `json:"code_challenge_methods_supported"`
	ClientIDMetadataDocumentSupported          bool     `json:"client_id_metadata_document_supported"`
	AuthorizationResponseIssParameterSupported bool     `json:"authorization_response_iss_parameter_supported"`
}

func (h *OAuthHandler) HandleMetadata(c *handler.Context) error {
	c.SetHeader("Cache-Control", "public, max-age=3600")
	return c.JSON(http.StatusOK, metadata{
		Issuer:                            h.config.Issuer,
		AuthorizationEndpoint:             h.config.AuthorizationEndpoint,
		TokenEndpoint:                     h.config.Issuer + TokenPath,
		ScopesSupported:                   scopes,
		ResponseTypesSupported:            []string{"code"},
		GrantTypesSupported:               []string{"authorization_code", "refresh_token"},
		TokenEndpointAuthMethodsSupported: []string{"none"},
		CodeChallengeMethodsSupported:     []string{"S256"},
		// Claude only uses metadata documents if this and the none auth
		// method are set.
		ClientIDMetadataDocumentSupported: true,
		// ChatGPT uses one client ID for all users with it, instead of one
		// per connection.
		AuthorizationResponseIssParameterSupported: true,
	})
}

// authorizeRequest is an authorization request that was checked.
type authorizeRequest struct {
	client        *Client
	redirectURI   string
	state         string
	scopes        []string
	codeChallenge string
	resource      string
}

// checkAuthorizeRequest checks the authorization request of a client, which
// the consent page passes on in the query string of its requests.
func (h *OAuthHandler) checkAuthorizeRequest(c *handler.Context) (*authorizeRequest, error) {
	client, err := h.clients.Client(c.Context(), c.Query("client_id"))
	if err != nil {
		return nil, err
	}
	if !client.AllowsRedirectURI(c.Query("redirect_uri")) {
		return nil, handler.ErrBadRequest("invalid_redirect_uri", "redirect_uri isn't one of the client's redirect URIs")
	}

	if c.Query("response_type") != "code" {
		return nil, handler.ErrBadRequest("unsupported_response_type", "response_type must be code")
	}
	if c.Query("code_challenge_method") != "S256" {
		return nil, handler.ErrBadRequest("invalid_request", "code_challenge_method must be S256")
	}
	if !validCodeChallenge(c.Query("code_challenge")) {
		return nil, handler.ErrBadRequest("invalid_request", "code_challenge must be the base64url SHA-256 of the code verifier")
	}
	if len(c.Query("state")) > maxStateLength {
		return nil, handler.ErrBadRequest("invalid_request", fmt.Sprintf("state must be at most %d characters", maxStateLength))
	}

	resource, ok := h.resource(c.Query("resource"))
	if !ok {
		return nil, handler.ErrBadRequest("invalid_target", "resource isn't a server of Kite")
	}

	return &authorizeRequest{
		client:        client,
		redirectURI:   c.Query("redirect_uri"),
		state:         c.Query("state"),
		scopes:        requestedScopes(c.Query("scope")),
		codeChallenge: c.Query("code_challenge"),
		resource:      resource,
	}, nil
}

// resource is the resource a token is for, if it's one of Kite's.
func (h *OAuthHandler) resource(resource string) (string, bool) {
	if resource == "" {
		return h.config.Resources[0], true
	}
	for _, r := range h.config.Resources {
		if strings.TrimSuffix(resource, "/") == strings.TrimSuffix(r, "/") {
			return r, true
		}
	}
	return "", false
}

// requestedScopes are the scopes a client asks for, which the user can grant.
// Unknown ones are left out, and it's all scopes if it asks for none of them.
// Write includes read, so the user can grant read instead.
func requestedScopes(scope string) []string {
	requested := model.APIToken{Scopes: strings.Fields(scope)}
	var granted []string
	for _, s := range scopes {
		resource, access, _ := strings.Cut(s, ":")
		if requested.HasAccess(resource, access) {
			granted = append(granted, s)
		}
	}
	if len(granted) == 0 {
		return slices.Clone(scopes)
	}
	return granted
}

// validCodeChallenge reports whether the challenge can be an S256 one, the
// base64url SHA-256 of a code verifier.
func validCodeChallenge(challenge string) bool {
	hash, err := base64.RawURLEncoding.DecodeString(challenge)
	return err == nil && len(hash) == sha256.Size
}

// HandleOAuthAuthorizeGet checks the authorization request of the consent
// page's URL, and returns what the page shows.
func (h *OAuthHandler) HandleOAuthAuthorizeGet(c *handler.Context) (*wire.OAuthAuthorizeGetResponse, error) {
	req, err := h.checkAuthorizeRequest(c)
	if err != nil {
		return nil, err
	}

	return &wire.OAuthAuthorizeGetResponse{
		Client: wire.OAuthClient{
			ID:     req.client.ID,
			Name:   req.client.Name,
			Domain: req.client.Domain,
		},
		Scopes: req.scopes,
	}, nil
}

// HandleOAuthAuthorize is the user approving or denying a client on the
// consent page. It returns where to send the user, back to the client with a
// code if they approved.
func (h *OAuthHandler) HandleOAuthAuthorize(c *handler.Context, body wire.OAuthAuthorizeRequest) (*wire.OAuthAuthorizeResponse, error) {
	req, err := h.checkAuthorizeRequest(c)
	if err != nil {
		return nil, err
	}

	if !body.Approved {
		return &wire.OAuthAuthorizeResponse{
			RedirectURL: h.redirectURL(req, url.Values{"error": {"access_denied"}}),
		}, nil
	}

	for _, scope := range body.Scopes {
		if !slices.Contains(req.scopes, scope) {
			return nil, handler.ErrBadRequest("invalid_scope", fmt.Sprintf("the client didn't ask for %s", scope))
		}
	}
	if err := apitoken.CheckAppIDs(c.Context(), h.appStore, c.Session.UserID, body.AppIDs); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	if err := apitoken.CheckTokenLimit(c.Context(), h.apiTokenStore, c.Session.UserID, h.config.MaxAPITokensPerUser, now); err != nil {
		return nil, err
	}

	code := util.SecureURLKey()
	err = h.codeStore.CreateOAuthCode(c.Context(), &model.OAuthCode{
		CodeHash:      util.HashKey(code),
		UserID:        c.Session.UserID,
		ClientID:      req.client.ID,
		ClientName:    req.client.Name,
		RedirectURI:   req.redirectURI,
		CodeChallenge: req.codeChallenge,
		Scopes:        apitoken.SortedUnique(body.Scopes),
		AppIDs:        apitoken.SortedUnique(body.AppIDs),
		Resource:      req.resource,
		ExpiresAt:     now.Add(codeExpiry),
		CreatedAt:     now,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create OAuth code: %w", err)
	}

	return &wire.OAuthAuthorizeResponse{
		RedirectURL: h.redirectURL(req, url.Values{"code": {code}}),
	}, nil
}

// redirectURL is the client's redirect URI with the response, and the state
// and issuer, which tells clients the response is from Kite (RFC 9207).
func (h *OAuthHandler) redirectURL(req *authorizeRequest, params url.Values) string {
	// It was parsed when checked.
	u, _ := url.Parse(req.redirectURI)
	query := u.Query()
	for key, values := range params {
		query[key] = values
	}
	if req.state != "" {
		query.Set("state", req.state)
	}
	query.Set("iss", h.config.Issuer)
	u.RawQuery = query.Encode()
	return u.String()
}
