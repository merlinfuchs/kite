package store

import (
	"context"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

type APITokenStore interface {
	CreateAPIToken(ctx context.Context, token *model.APIToken) (*model.APIToken, error)
	APITokensByUser(ctx context.Context, userID string) ([]*model.APIToken, error)
	CountAPITokensByUser(ctx context.Context, userID string) (int, error)
	APITokenByKeyHash(ctx context.Context, keyHash string) (*model.APIToken, error)
	UpdateAPITokenLastUsedAt(ctx context.Context, id string, lastUsedAt time.Time) error
	// DeleteAPIToken returns ErrNotFound if the user has no token with the ID.
	DeleteAPIToken(ctx context.Context, userID string, id string) error
}
