package wire

import (
	"errors"
	"slices"
	"strings"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"gopkg.in/guregu/null.v4"
)

// MaxAPITokenDays is how long a token can last at most, unless it doesn't
// expire at all.
const MaxAPITokenDays = 365

// MaxAPITokenApps is how many apps a token that can't use all of them can
// have.
const MaxAPITokenApps = 100

// APIToken never includes the token itself, it's only shown when created.
type APIToken struct {
	ID     string   `json:"id"`
	Name   string   `json:"name"`
	Scopes []string `json:"scopes"`
	// The apps the token can use, null for all apps of the user.
	AppIDs     []string  `json:"app_ids" tstype:"string[] | null"`
	LastUsedAt null.Time `json:"last_used_at"`
	ExpiresAt  null.Time `json:"expires_at"`
	CreatedAt  time.Time `json:"created_at"`
	// The OAuth client the token is for, like Claude, null for tokens the
	// user made. It's the URL of the client's metadata document.
	OAuthClientID null.String `json:"oauth_client_id"`
}

type APITokenListResponse = []*APIToken

type APITokenCreateRequest struct {
	Name string `json:"name"`
	// How many days the token lasts, null for a token that doesn't expire.
	ExpiresInDays null.Int `json:"expires_in_days"`
	// Like kite:read, see model.APITokenScopes.
	Scopes []string `json:"scopes"`
	// The apps the token can use, null for all apps of the user.
	AppIDs []string `json:"app_ids" tstype:"string[] | null"`
}

func (req APITokenCreateRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Name,
			validation.Required,
			validation.Length(1, 100),
			validation.NewStringRule(func(s string) bool { return strings.TrimSpace(s) != "" }, "must not be blank"),
		),
		validation.Field(&req.ExpiresInDays, validation.When(req.ExpiresInDays.Valid,
			// Min skips 0 as empty, Required doesn't.
			validation.Required,
			validation.Min(int64(1)),
			validation.Max(int64(MaxAPITokenDays)),
		)),
		validation.Field(&req.Scopes, apiTokenScopesRules(req.Scopes)...),
		validation.Field(&req.AppIDs, apiTokenAppIDsRules(req.AppIDs)...),
	)
}

func apiTokenScopesRules(scopes []string) []validation.Rule {
	return []validation.Rule{
		validation.Required,
		validation.Each(validation.In(stringsToAny(model.APITokenScopes)...)),
		// Clients need it to find the apps, even for the Discord API.
		validation.By(func(any) error {
			if !slices.Contains(scopes, model.APITokenScopeKiteRead) && !slices.Contains(scopes, model.APITokenScopeKiteWrite) {
				return errors.New("must include kite:read or kite:write")
			}
			return nil
		}),
	}
}

func apiTokenAppIDsRules(appIDs []string) []validation.Rule {
	// Null is all apps, so an empty list must not become null.
	return []validation.Rule{validation.When(appIDs != nil,
		validation.Required,
		validation.Length(1, MaxAPITokenApps),
	)}
}

func stringsToAny(s []string) []any {
	res := make([]any, len(s))
	for i, v := range s {
		res[i] = v
	}
	return res
}

type APITokenCreateResponse struct {
	APIToken *APIToken `json:"api_token"`
	// The token to send as "Authorization: Bearer <token>". It can't be read
	// again later.
	Token string `json:"token"`
}

type APITokenCurrentGetResponse = APIToken

type APITokenDeleteResponse struct{}

func APITokenToWire(token *model.APIToken) *APIToken {
	if token == nil {
		return nil
	}

	return &APIToken{
		ID:         token.ID,
		Name:       token.Name,
		Scopes:     token.Scopes,
		AppIDs:     token.AppIDs,
		LastUsedAt: token.LastUsedAt,
		ExpiresAt:  token.ExpiresAt,
		CreatedAt:  token.CreatedAt,

		OAuthClientID: token.OAuthClientID,
	}
}
