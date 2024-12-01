package apikey

import (
	"fmt"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/internal/util"
)

type APIKeyHandler struct {
	store store.APIKeyStore
}

func NewAPIKeyHandler(store store.APIKeyStore) *APIKeyHandler {
	return &APIKeyHandler{store: store}
}

func (h *APIKeyHandler) HandleAPIKeyCreate(c *handler.Context, req wire.APIKeyCreateRequest) (*wire.APIKeyCreateResponse, error) {
	key := util.SecureKey()
	keyHash := util.HashKey(key)

	apiKey, err := h.store.CreateAPIKey(c.Context(), &model.APIKey{
		ID:            util.UniqueID(),
		Type:          model.APIKeyType(req.Type),
		Name:          req.Name,
		Key:           key,
		KeyHash:       keyHash,
		AppID:         c.App.ID,
		CreatorUserID: c.Session.UserID,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create api key: %w", err)
	}

	return wire.APIKeyToWire(apiKey), nil
}
