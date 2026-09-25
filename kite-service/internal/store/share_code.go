package store

import (
	"context"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

type ShareCodeStore interface {
	CreateShareCode(ctx context.Context, shareCode *model.ShareCode) (*model.ShareCode, error)
	ShareCode(ctx context.Context, code string) (*model.ShareCode, error)
}
