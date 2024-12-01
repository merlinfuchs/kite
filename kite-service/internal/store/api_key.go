package store

import (
	"context"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

type APIKeyStore interface {
	CreateAPIKey(ctx context.Context, key *model.APIKey) (*model.APIKey, error)
	APIKeysByApp(ctx context.Context, appID string) ([]*model.APIKey, error)
	APIKeyByKeyHash(ctx context.Context, keyHash string) (*model.APIKey, error)
}
