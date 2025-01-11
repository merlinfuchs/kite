package store

import (
	"context"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

type SuspendPointStore interface {
	CreateSuspendPoint(ctx context.Context, suspendPoint *model.SuspendPoint) (*model.SuspendPoint, error)
	DeleteSuspendPoint(ctx context.Context, id string) error
	DeleteExpiredSuspendPoints(ctx context.Context, timestamp time.Time) error
	SuspendPoint(ctx context.Context, id string) (*model.SuspendPoint, error)
}
