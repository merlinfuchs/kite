package store

import (
	"context"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

type UsageStore interface {
	CreateUsageRecord(ctx context.Context, record model.UsageRecord) error
	UsageRecordsBetween(ctx context.Context, appID string, start time.Time, end time.Time) ([]model.UsageRecord, error)
	UsageCreditsUsedBetween(ctx context.Context, appID string, start time.Time, end time.Time) (uint32, error)
}