package wire

import (
	"strings"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"gopkg.in/guregu/null.v4"
)

// MaxAPITokenDays is how long a token can last at most, unless it doesn't
// expire at all.
const MaxAPITokenDays = 365

// APIToken never includes the token itself, it's only shown when created.
type APIToken struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Scopes     []string  `json:"scopes"`
	LastUsedAt null.Time `json:"last_used_at"`
	ExpiresAt  null.Time `json:"expires_at"`
	CreatedAt  time.Time `json:"created_at"`
}

type APITokenListResponse = []*APIToken

type APITokenCreateRequest struct {
	Name string `json:"name"`
	// How many days the token lasts, null for a token that doesn't expire.
	ExpiresInDays null.Int `json:"expires_in_days"`
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
	)
}

type APITokenCreateResponse struct {
	APIToken *APIToken `json:"api_token"`
	// The token to send as "Authorization: Bearer <token>". It can't be read
	// again later.
	Token string `json:"token"`
}

type APITokenDeleteResponse struct{}

func APITokenToWire(token *model.APIToken) *APIToken {
	if token == nil {
		return nil
	}

	return &APIToken{
		ID:         token.ID,
		Name:       token.Name,
		Scopes:     token.Scopes,
		LastUsedAt: token.LastUsedAt,
		ExpiresAt:  token.ExpiresAt,
		CreatedAt:  token.CreatedAt,
	}
}
