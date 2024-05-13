// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0
// source: app_usage.sql

package pgmodel

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createAppUsageEntry = `-- name: CreateAppUsageEntry :exec
INSERT INTO app_usage (
    app_id,
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

type CreateAppUsageEntryParams struct {
	AppID                   string
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

func (q *Queries) CreateAppUsageEntry(ctx context.Context, arg CreateAppUsageEntryParams) error {
	_, err := q.db.Exec(ctx, createAppUsageEntry,
		arg.AppID,
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

const getAppUsageAndLimits = `-- name: GetAppUsageAndLimits :one
SELECT app_entitlements_resolved_view.app_id, feature_monthly_execution_time_limit, app_usage_month_view.app_id, total_event_count, success_event_count, total_event_execution_time, avg_event_execution_time, total_event_total_time, avg_event_total_time, total_call_count, success_call_count, total_call_total_time, avg_call_total_time FROM app_entitlements_resolved_view 
LEFT JOIN app_usage_month_view  
ON app_usage_month_view.app_id = app_entitlements_resolved_view.app_id 
WHERE app_entitlements_resolved_view.app_id = $1
`

type GetAppUsageAndLimitsRow struct {
	AppID                            string
	FeatureMonthlyExecutionTimeLimit int32
	AppID_2                          pgtype.Text
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

func (q *Queries) GetAppUsageAndLimits(ctx context.Context, appID string) (GetAppUsageAndLimitsRow, error) {
	row := q.db.QueryRow(ctx, getAppUsageAndLimits, appID)
	var i GetAppUsageAndLimitsRow
	err := row.Scan(
		&i.AppID,
		&i.FeatureMonthlyExecutionTimeLimit,
		&i.AppID_2,
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

const getAppUsageSummary = `-- name: GetAppUsageSummary :one
SELECT app_id, total_event_count, success_event_count, total_event_execution_time, avg_event_execution_time, total_event_total_time, avg_event_total_time, total_call_count, success_call_count, total_call_total_time, avg_call_total_time FROM app_usage_month_view WHERE app_id = $1
`

func (q *Queries) GetAppUsageSummary(ctx context.Context, appID string) (AppUsageMonthView, error) {
	row := q.db.QueryRow(ctx, getAppUsageSummary, appID)
	var i AppUsageMonthView
	err := row.Scan(
		&i.AppID,
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

const getLastAppUsageEntry = `-- name: GetLastAppUsageEntry :one
SELECT id, app_id, total_event_count, success_event_count, total_event_execution_time, avg_event_execution_time, total_event_total_time, avg_event_total_time, total_call_count, success_call_count, total_call_total_time, avg_call_total_time, period_starts_at, period_ends_at FROM app_usage WHERE app_id = $1 ORDER BY period_ends_at DESC LIMIT 1
`

func (q *Queries) GetLastAppUsageEntry(ctx context.Context, appID string) (AppUsage, error) {
	row := q.db.QueryRow(ctx, getLastAppUsageEntry, appID)
	var i AppUsage
	err := row.Scan(
		&i.ID,
		&i.AppID,
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