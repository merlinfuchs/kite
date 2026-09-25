package store

import (
	"context"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

type ShareCodeStore interface {
	CreateShareCode(ctx context.Context, shareCode *model.ShareCode) error
	ShareCode(ctx context.Context, code string) (*model.ShareCode, error)
	TouchShareCode(ctx context.Context, code string, usedAt time.Time) error
	// DeleteUnusedShareCodes deletes up to batchSize share codes last used before usedBefore.
	DeleteUnusedShareCodes(ctx context.Context, usedBefore time.Time, batchSize int) (int64, error)
}
