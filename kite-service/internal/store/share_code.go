package store

import (
	"context"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

type ShareCodeStore interface {
	// CreateShareCode returns the existing code instead if the app already shared the same data.
	CreateShareCode(ctx context.Context, shareCode *model.ShareCode) (string, error)
	ShareCode(ctx context.Context, code string) (*model.ShareCode, error)
	TouchShareCode(ctx context.Context, code string, usedAt time.Time) error
	// DeleteUnusedShareCodes deletes up to batchSize share codes last used before usedBefore.
	DeleteUnusedShareCodes(ctx context.Context, usedBefore time.Time, batchSize int) (int64, error)
}
