package store

import (
	"context"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

type UsageStore interface {
	CreateUsageRecord(ctx context.Context, record model.UsageRecord) error
	UsageRecordsBetween(ctx context.Context, appID string, start time.Time, end time.Time) ([]model.UsageRecord, error)
	UsageCreditsUsedBetween(ctx context.Context, appID string, start time.Time, end time.Time) (int, error)
	UsageCreditsUsedByTypeBetween(ctx context.Context, appID string, start time.Time, end time.Time) ([]model.UsageCreditsUsedByType, error)
	UsageCreditsUsedByDayBetween(ctx context.Context, appID string, start time.Time, end time.Time) ([]model.UsageCreditsUsedByDay, error)
	AllUsageCreditsUsedBetween(ctx context.Context, start time.Time, end time.Time) (map[string]int, error)
	DeleteUsageRecordsBefore(ctx context.Context, before time.Time) error
}
