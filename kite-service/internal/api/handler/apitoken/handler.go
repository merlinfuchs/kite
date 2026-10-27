package apitoken

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/internal/util"
)

// TokenPrefix makes tokens recognizable, e.g. by secret scanners.
const TokenPrefix = "kite_"

type APITokenHandler struct {
	apiTokenStore       store.APITokenStore
	appStore            store.AppStore
	maxAPITokensPerUser int
}

func NewAPITokenHandler(apiTokenStore store.APITokenStore, appStore store.AppStore, maxAPITokensPerUser int) *APITokenHandler {
	return &APITokenHandler{
		apiTokenStore:       apiTokenStore,
		appStore:            appStore,
		maxAPITokensPerUser: maxAPITokensPerUser,
	}
}

func (h *APITokenHandler) HandleAPITokenList(c *handler.Context) (*wire.APITokenListResponse, error) {
	tokens, err := h.apiTokenStore.APITokensByUser(c.Context(), c.Session.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get API tokens: %w", err)
	}

	res := make([]*wire.APIToken, len(tokens))
	for i, token := range tokens {
		res[i] = wire.APITokenToWire(token)
	}
	return &res, nil
}

func (h *APITokenHandler) HandleAPITokenCreate(c *handler.Context, req wire.APITokenCreateRequest) (*wire.APITokenCreateResponse, error) {
	now := time.Now().UTC()
	count, err := h.apiTokenStore.CountAPITokensByUser(c.Context(), c.Session.UserID, now)
	if err != nil {
		return nil, fmt.Errorf("failed to count API tokens: %w", err)
	}
	if count >= h.maxAPITokensPerUser {
		return nil, handler.ErrBadRequest("resource_limit", fmt.Sprintf("maximum number of API tokens (%d) reached", h.maxAPITokensPerUser))
	}

	if err := CheckAppIDs(c.Context(), h.appStore, c.Session.UserID, req.AppIDs); err != nil {
		return nil, err
	}

	token := model.APIToken{
		ID:        util.UniqueID(),
		UserID:    c.Session.UserID,
		Name:      req.Name,
		Scopes:    SortedUnique(req.Scopes),
		AppIDs:    SortedUnique(req.AppIDs),
		CreatedAt: now,
	}
	if req.ExpiresInDays.Valid {
		token.ExpiresAt.SetValid(now.AddDate(0, 0, int(req.ExpiresInDays.Int64)))
	}

	key := TokenPrefix + util.SecureURLKey()
	token.KeyHash = util.HashKey(key)

	created, err := h.apiTokenStore.CreateAPIToken(c.Context(), &token)
	if err != nil {
		return nil, fmt.Errorf("failed to create API token: %w", err)
	}

	return &wire.APITokenCreateResponse{
		APIToken: wire.APITokenToWire(created),
		Token:    key,
	}, nil
}

// HandleAPITokenCurrentGet is the token of the request, so clients like the
// MCP server know what it can do.
func (h *APITokenHandler) HandleAPITokenCurrentGet(c *handler.Context) (*wire.APITokenCurrentGetResponse, error) {
	if c.APIToken == nil {
		return nil, handler.ErrBadRequest("no_api_token", "The request wasn't made with an API token")
	}
	return wire.APITokenToWire(c.APIToken), nil
}

func (h *APITokenHandler) HandleAPITokenDelete(c *handler.Context) (*wire.APITokenDeleteResponse, error) {
	err := h.apiTokenStore.DeleteAPIToken(c.Context(), c.Session.UserID, c.Param("tokenID"))
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, handler.ErrNotFound("unknown_api_token", "API token not found")
		}
		return nil, fmt.Errorf("failed to delete API token: %w", err)
	}

	return &wire.APITokenDeleteResponse{}, nil
}

// CheckAppIDs returns an error unless all apps of a token are apps of the
// user. Nil app IDs are all apps.
func CheckAppIDs(ctx context.Context, appStore store.AppStore, userID string, appIDs []string) error {
	if appIDs == nil {
		return nil
	}

	apps, err := appStore.AppsByUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get apps: %w", err)
	}
	for _, appID := range appIDs {
		if !slices.ContainsFunc(apps, func(app *model.App) bool { return app.ID == appID }) {
			return handler.ErrNotFound("unknown_app", fmt.Sprintf("App %s not found", appID))
		}
	}
	return nil
}

// SortedUnique keeps nil as nil, as nil app IDs mean all apps.
func SortedUnique(s []string) []string {
	if s == nil {
		return nil
	}
	return slices.Compact(slices.Sorted(slices.Values(s)))
}
