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
	// DeleteUsageRecordsBefore deletes up to batchSize usage records created
	// before the given time and returns how many were deleted.
	DeleteUsageRecordsBefore(ctx context.Context, before time.Time, batchSize int) (int64, error)
	// RollupUsageRecordsBefore sums the usage records of every whole day before
	// the given midnight into daily rollups, which outlive the records.
	RollupUsageRecordsBefore(ctx context.Context, before time.Time) (int64, error)
	// The analytics methods cover both rolled up days and live usage records.
	UsageAnalyticsTotalsBetween(ctx context.Context, appID string, start time.Time, end time.Time) ([]model.UsageAnalyticsTotal, error)
	// bucket is a date_trunc unit: hour, day or month.
	UsageAnalyticsSeriesBetween(ctx context.Context, appID string, start time.Time, end time.Time, bucket string) ([]model.UsageAnalyticsBucket, error)
	UsageAnalyticsTopSourcesBetween(ctx context.Context, appID string, start time.Time, end time.Time, perType int) ([]model.UsageAnalyticsSource, error)
}
