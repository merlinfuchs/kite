package store

import (
	"context"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

type OAuthCodeStore interface {
	CreateOAuthCode(ctx context.Context, code *model.OAuthCode) error
	// UseOAuthCode marks the code as exchanged for the token with the ID. It
	// returns ErrNotFound if the code doesn't exist, expired or was used.
	UseOAuthCode(ctx context.Context, codeHash string, apiTokenID string, now time.Time) (*model.OAuthCode, error)
	OAuthCode(ctx context.Context, codeHash string) (*model.OAuthCode, error)
	DeleteExpiredOAuthCodes(ctx context.Context, expiredBefore time.Time, batchSize int) (int64, error)
}
