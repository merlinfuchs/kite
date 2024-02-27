// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0
// source: guild_usage.sql

package pgmodel

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createGuildUsageEntry = `-- name: CreateGuildUsageEntry :exec
INSERT INTO guild_usage (
    guild_id,
    total_event_count,
    success_event_count,
    total_event_execution_time,
    avg_event_execution_time,
    total_event_total_time,
    avg_event_total_time,
    total_call_count,
    success_call_count,
    total_call_total_time,
    avg_call_total_time,
    period_starts_at,
    period_ends_at
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    $9,
    $10,
    $11,
    $12,
    $13
)
`

type CreateGuildUsageEntryParams struct {
	GuildID                 string
	TotalEventCount         int32
	SuccessEventCount       int32
	TotalEventExecutionTime int64
	AvgEventExecutionTime   int64
	TotalEventTotalTime     int64
	AvgEventTotalTime       int64
	TotalCallCount          int32
	SuccessCallCount        int32
	TotalCallTotalTime      int64
	AvgCallTotalTime        int64
	PeriodStartsAt          pgtype.Timestamp
	PeriodEndsAt            pgtype.Timestamp
}

func (q *Queries) CreateGuildUsageEntry(ctx context.Context, arg CreateGuildUsageEntryParams) error {
	_, err := q.db.Exec(ctx, createGuildUsageEntry,
		arg.GuildID,
		arg.TotalEventCount,
		arg.SuccessEventCount,
		arg.TotalEventExecutionTime,
		arg.AvgEventExecutionTime,
		arg.TotalEventTotalTime,
		arg.AvgEventTotalTime,
		arg.TotalCallCount,
		arg.SuccessCallCount,
		arg.TotalCallTotalTime,
		arg.AvgCallTotalTime,
		arg.PeriodStartsAt,
		arg.PeriodEndsAt,
	)
	return err
}

const getGuildUsageAndLimits = `-- name: GetGuildUsageAndLimits :one
SELECT guild_entitlements_resolved_view.guild_id, feature_monthly_execution_time_limit, guild_usage_month_view.guild_id, total_event_count, success_event_count, total_event_execution_time, avg_event_execution_time, total_event_total_time, avg_event_total_time, total_call_count, success_call_count, total_call_total_time, avg_call_total_time FROM guild_entitlements_resolved_view 
LEFT JOIN guild_usage_month_view  
ON guild_usage_month_view.guild_id = guild_entitlements_resolved_view.guild_id 
WHERE guild_entitlements_resolved_view.guild_id = $1
`

type GetGuildUsageAndLimitsRow struct {
	GuildID                          string
	FeatureMonthlyExecutionTimeLimit int32
	GuildID_2                        pgtype.Text
	TotalEventCount                  pgtype.Int8
	SuccessEventCount                pgtype.Int8
	TotalEventExecutionTime          pgtype.Int8
	AvgEventExecutionTime            pgtype.Float8
	TotalEventTotalTime              pgtype.Int8
	AvgEventTotalTime                pgtype.Float8
	TotalCallCount                   pgtype.Int8
	SuccessCallCount                 pgtype.Int8
	TotalCallTotalTime               pgtype.Int8
	AvgCallTotalTime                 pgtype.Float8
}

func (q *Queries) GetGuildUsageAndLimits(ctx context.Context, guildID string) (GetGuildUsageAndLimitsRow, error) {
	row := q.db.QueryRow(ctx, getGuildUsageAndLimits, guildID)
	var i GetGuildUsageAndLimitsRow
	err := row.Scan(
		&i.GuildID,
		&i.FeatureMonthlyExecutionTimeLimit,
		&i.GuildID_2,
		&i.TotalEventCount,
		&i.SuccessEventCount,
		&i.TotalEventExecutionTime,
		&i.AvgEventExecutionTime,
		&i.TotalEventTotalTime,
		&i.AvgEventTotalTime,
		&i.TotalCallCount,
		&i.SuccessCallCount,
		&i.TotalCallTotalTime,
		&i.AvgCallTotalTime,
	)
	return i, err
}

const getGuildUsageSummary = `-- name: GetGuildUsageSummary :one
SELECT guild_id, total_event_count, success_event_count, total_event_execution_time, avg_event_execution_time, total_event_total_time, avg_event_total_time, total_call_count, success_call_count, total_call_total_time, avg_call_total_time FROM guild_usage_month_view WHERE guild_id = $1
`

func (q *Queries) GetGuildUsageSummary(ctx context.Context, guildID string) (GuildUsageMonthView, error) {
	row := q.db.QueryRow(ctx, getGuildUsageSummary, guildID)
	var i GuildUsageMonthView
	err := row.Scan(
		&i.GuildID,
		&i.TotalEventCount,
		&i.SuccessEventCount,
		&i.TotalEventExecutionTime,
		&i.AvgEventExecutionTime,
		&i.TotalEventTotalTime,
		&i.AvgEventTotalTime,
		&i.TotalCallCount,
		&i.SuccessCallCount,
		&i.TotalCallTotalTime,
		&i.AvgCallTotalTime,
	)
	return i, err
}

const getLastGuildUsageEntry = `-- name: GetLastGuildUsageEntry :one
SELECT id, guild_id, total_event_count, success_event_count, total_event_execution_time, avg_event_execution_time, total_event_total_time, avg_event_total_time, total_call_count, success_call_count, total_call_total_time, avg_call_total_time, period_starts_at, period_ends_at FROM guild_usage WHERE guild_id = $1 ORDER BY period_ends_at DESC LIMIT 1
`

func (q *Queries) GetLastGuildUsageEntry(ctx context.Context, guildID string) (GuildUsage, error) {
	row := q.db.QueryRow(ctx, getLastGuildUsageEntry, guildID)
	var i GuildUsage
	err := row.Scan(
		&i.ID,
		&i.GuildID,
		&i.TotalEventCount,
		&i.SuccessEventCount,
		&i.TotalEventExecutionTime,
		&i.AvgEventExecutionTime,
		&i.TotalEventTotalTime,
		&i.AvgEventTotalTime,
		&i.TotalCallCount,
		&i.SuccessCallCount,
		&i.TotalCallTotalTime,
		&i.AvgCallTotalTime,
		&i.PeriodStartsAt,
		&i.PeriodEndsAt,
	)
	return i, err
}
