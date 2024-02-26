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

func (c *Client) CreateGuildUsageEntry(ctx context.Context, entry model.GuildUsageEntry) error {
	err := c.Q.CreateGuildUsageEntry(ctx, pgmodel.CreateGuildUsageEntryParams{
		GuildID:                 string(entry.GuildID),
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

func (c *Client) GetLastGuildUsageEntry(ctx context.Context, guildID distype.Snowflake) (*model.GuildUsageEntry, error) {
	row, err := c.Q.GetLastGuildUsageEntry(ctx, string(guildID))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	return &model.GuildUsageEntry{
		GuildID:                 distype.Snowflake(row.GuildID),
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

func (c *Client) GetGuildUsageSummary(ctx context.Context, guildID distype.Snowflake, startAt, endAt time.Time) (*model.GuildUsageSummary, error) {
	row, err := c.Q.GetGuildUsageSummary(ctx, pgmodel.GetGuildUsageSummaryParams{
		GuildID: string(guildID),
		StartAt: timeToTimestamp(startAt),
		EndAt:   timeToTimestamp(endAt),
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return &model.GuildUsageSummary{}, nil
		}
		return nil, err
	}

	return &model.GuildUsageSummary{
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

func (c *Client) GetGuildUsageAndLimits(ctx context.Context, guildID distype.Snowflake) (*model.GuildUsageAndLimits, error) {
	row, err := c.Q.GetGuildUsageAndLimits(ctx, string(guildID))
	if err != nil {
		return nil, err
	}

	return &model.GuildUsageAndLimits{
		Usage: model.GuildUsageSummary{
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
		},
		Limits: model.GuildEntitlementResolved{
			MonthlyCpuTimeLimit: int(row.FeatureMonthlyCpuTimeLimit),
		},
	}, nil
}
