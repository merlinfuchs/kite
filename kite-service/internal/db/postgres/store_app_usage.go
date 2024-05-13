package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/merlinfuchs/dismod/distype"
	"github.com/merlinfuchs/kite/kite-service/internal/db/postgres/pgmodel"
	"github.com/merlinfuchs/kite/kite-service/pkg/model"
	"github.com/merlinfuchs/kite/kite-service/pkg/store"
)

var _ store.AppUsageStore = (*Client)(nil)

func (c *Client) CreateAppUsageEntry(ctx context.Context, entry model.AppUsageEntry) error {
	err := c.Q.CreateAppUsageEntry(ctx, pgmodel.CreateAppUsageEntryParams{
		AppID:                   string(entry.AppID),
		TotalEventCount:         int32(entry.TotalEventCount),
		SuccessEventCount:       int32(entry.SuccessEventCount),
		TotalEventExecutionTime: entry.TotalEventExecutionTime.Microseconds(),
		AvgEventExecutionTime:   entry.AvgEventExecutionTime.Microseconds(),
		TotalEventTotalTime:     entry.TotalEventTotalTime.Microseconds(),
		AvgEventTotalTime:       entry.AvgEventTotalTime.Microseconds(),
		TotalCallCount:          int32(entry.TotalCallCount),
		SuccessCallCount:        int32(entry.SuccessCallCount),
		TotalCallTotalTime:      entry.TotalCallTotalTime.Microseconds(),
		AvgCallTotalTime:        entry.AvgCallTotalTime.Microseconds(),
		PeriodStartsAt:          timeToTimestamp(entry.PeriodStartsAt),
		PeriodEndsAt:            timeToTimestamp(entry.PeriodEndsAt),
	})
	return err

}

func (c *Client) GetLastAppUsageEntry(ctx context.Context, appID distype.Snowflake) (*model.AppUsageEntry, error) {
	row, err := c.Q.GetLastAppUsageEntry(ctx, string(appID))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	return &model.AppUsageEntry{
		AppID:                   distype.Snowflake(row.AppID),
		TotalEventCount:         int(row.TotalEventCount),
		SuccessEventCount:       int(row.SuccessEventCount),
		TotalEventExecutionTime: time.Duration(row.TotalEventExecutionTime) * time.Microsecond,
		AvgEventExecutionTime:   time.Duration(row.AvgEventExecutionTime) * time.Microsecond,
		TotalEventTotalTime:     time.Duration(row.TotalEventTotalTime) * time.Microsecond,
		AvgEventTotalTime:       time.Duration(row.AvgEventTotalTime) * time.Microsecond,
		TotalCallCount:          int(row.TotalCallCount),
		SuccessCallCount:        int(row.SuccessCallCount),
		TotalCallTotalTime:      time.Duration(row.TotalCallTotalTime) * time.Microsecond,
		AvgCallTotalTime:        time.Duration(row.AvgCallTotalTime) * time.Microsecond,
		PeriodStartsAt:          row.PeriodStartsAt.Time,
		PeriodEndsAt:            row.PeriodEndsAt.Time,
	}, nil
}

func (c *Client) GetAppUsageSummary(ctx context.Context, appID distype.Snowflake) (*model.AppUsageSummary, error) {
	row, err := c.Q.GetAppUsageSummary(ctx, string(appID))
	if err != nil {
		if err == pgx.ErrNoRows {
			return &model.AppUsageSummary{}, nil
		}
		return nil, err
	}

	return &model.AppUsageSummary{
		TotalEventCount:         int(row.TotalEventCount),
		SuccessEventCount:       int(row.SuccessEventCount),
		TotalEventExecutionTime: time.Duration(row.TotalEventExecutionTime) * time.Microsecond,
		AvgEventExecutionTime:   time.Duration(row.AvgEventExecutionTime) * time.Microsecond,
		TotalEventTotalTime:     time.Duration(row.TotalEventTotalTime) * time.Microsecond,
		AvgEventTotalTime:       time.Duration(row.AvgEventTotalTime) * time.Microsecond,
		TotalCallCount:          int(row.TotalCallCount),
		SuccessCallCount:        int(row.SuccessCallCount),
		TotalCallTotalTime:      time.Duration(row.TotalCallTotalTime) * time.Microsecond,
		AvgCallTotalTime:        time.Duration(row.AvgCallTotalTime) * time.Microsecond,
	}, nil
}

func (c *Client) GetAppUsageAndLimits(ctx context.Context, appID distype.Snowflake) (*model.AppUsageAndLimits, error) {
	row, err := c.Q.GetAppUsageAndLimits(ctx, string(appID))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	return &model.AppUsageAndLimits{
		AppUsageSummary: model.AppUsageSummary{
			TotalEventCount:         int(row.TotalEventCount.Int64),
			SuccessEventCount:       int(row.SuccessEventCount.Int64),
			TotalEventExecutionTime: time.Duration(row.TotalEventExecutionTime.Int64) * time.Microsecond,
			AvgEventExecutionTime:   time.Duration(row.AvgEventExecutionTime.Float64) * time.Microsecond,
			TotalEventTotalTime:     time.Duration(row.TotalEventTotalTime.Int64) * time.Microsecond,
			AvgEventTotalTime:       time.Duration(row.AvgEventTotalTime.Float64) * time.Microsecond,
			TotalCallCount:          int(row.TotalCallCount.Int64),
			SuccessCallCount:        int(row.SuccessCallCount.Int64),
			TotalCallTotalTime:      time.Duration(row.TotalCallTotalTime.Int64) * time.Microsecond,
			AvgCallTotalTime:        time.Duration(row.AvgCallTotalTime.Float64) * time.Microsecond,
		},
		Limits: model.AppEntitlementResolved{
			MonthlyExecutionTimeLimit: time.Duration(row.FeatureMonthlyExecutionTimeLimit) * time.Millisecond,
		},
	}, nil
}
