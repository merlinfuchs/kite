package store

import (
	"context"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

type StatsStore interface {
	Stats(ctx context.Context, executionsSince time.Time) (model.Stats, error)
}
