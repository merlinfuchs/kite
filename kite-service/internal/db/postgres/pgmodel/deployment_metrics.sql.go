// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.22.0
// source: deployment_metrics.sql

package pgmodel

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createDeploymentMetricEntry = `-- name: CreateDeploymentMetricEntry :exec
INSERT INTO deployment_metrics (
    deployment_id,
    type,
    metadata,
    event_type,
    event_success,
    event_execution_time,
    event_total_time,
    call_type,
    call_success,
    call_total_time,
    timestamp
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
    $11
)
`

type CreateDeploymentMetricEntryParams struct {
	DeploymentID       string
	Type               string
	Metadata           []byte
	EventType          string
	EventSuccess       bool
	EventExecutionTime int64
	EventTotalTime     int64
	CallType           string
	CallSuccess        bool
	CallTotalTime      int64
	Timestamp          pgtype.Timestamp
}

func (q *Queries) CreateDeploymentMetricEntry(ctx context.Context, arg CreateDeploymentMetricEntryParams) error {
	_, err := q.db.Exec(ctx, createDeploymentMetricEntry,
		arg.DeploymentID,
		arg.Type,
		arg.Metadata,
		arg.EventType,
		arg.EventSuccess,
		arg.EventExecutionTime,
		arg.EventTotalTime,
		arg.CallType,
		arg.CallSuccess,
		arg.CallTotalTime,
		arg.Timestamp,
	)
	return err
}

const getDeploymentCallMetrics = `-- name: GetDeploymentCallMetrics :many
SELECT
	generate_series::timestamp AS timestamp,
	COALESCE(total_count, 0) AS total_count,
	COALESCE(success_count, 0) AS success_count,
	COALESCE(avg_total_time, 0) AS avg_total_time
FROM (
	SELECT
		date_trunc($2::text, timestamp)::timestamp AS trunc_timestamp,
		COUNT(*) AS total_count,
		SUM(
			CASE WHEN call_success = TRUE THEN
				1
			ELSE
				0
			END) AS success_count, AVG(call_total_time) AS avg_total_time
	FROM
		deployment_metrics
	WHERE
		deployment_metrics.deployment_id = $1
		AND timestamp >= date_trunc($2::text, $3::timestamp)
		AND TYPE = 'CALL_EXECUTED'
	GROUP BY
		trunc_timestamp) AS y
	RIGHT JOIN generate_series(date_trunc($2::text, $3::timestamp), $4::timestamp, ($5::text)::interval) 
ON trunc_timestamp = generate_series
`

type GetDeploymentCallMetricsParams struct {
	DeploymentID string
	Precision    string
	StartAt      pgtype.Timestamp
	EndAt        pgtype.Timestamp
	SeriesStep   string
}

type GetDeploymentCallMetricsRow struct {
	Timestamp    pgtype.Timestamp
	TotalCount   int64
	SuccessCount int64
	AvgTotalTime float64
}

func (q *Queries) GetDeploymentCallMetrics(ctx context.Context, arg GetDeploymentCallMetricsParams) ([]GetDeploymentCallMetricsRow, error) {
	rows, err := q.db.Query(ctx, getDeploymentCallMetrics,
		arg.DeploymentID,
		arg.Precision,
		arg.StartAt,
		arg.EndAt,
		arg.SeriesStep,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetDeploymentCallMetricsRow
	for rows.Next() {
		var i GetDeploymentCallMetricsRow
		if err := rows.Scan(
			&i.Timestamp,
			&i.TotalCount,
			&i.SuccessCount,
			&i.AvgTotalTime,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getDeploymentCallMetricsNoFill = `-- name: GetDeploymentCallMetricsNoFill :many
SELECT 
    date_trunc($3, timestamp)::timestamp as timestamp,
    COUNT(*) AS total_count,
    SUM(CASE WHEN call_success = true THEN 1 ELSE 0 END) AS success_count,
    AVG(call_total_time) AS avg_total_time
FROM deployment_metrics 
WHERE 
    deployment_id = $1 AND 
    timestamp >= $2 AND
    type = 'CALL_EXECUTED' 
GROUP BY date_trunc($3, timestamp)
ORDER BY timestamp ASC
`

type GetDeploymentCallMetricsNoFillParams struct {
	DeploymentID string
	Timestamp    pgtype.Timestamp
	DateTrunc    string
}

type GetDeploymentCallMetricsNoFillRow struct {
	Timestamp    pgtype.Timestamp
	TotalCount   int64
	SuccessCount int64
	AvgTotalTime float64
}

func (q *Queries) GetDeploymentCallMetricsNoFill(ctx context.Context, arg GetDeploymentCallMetricsNoFillParams) ([]GetDeploymentCallMetricsNoFillRow, error) {
	rows, err := q.db.Query(ctx, getDeploymentCallMetricsNoFill, arg.DeploymentID, arg.Timestamp, arg.DateTrunc)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetDeploymentCallMetricsNoFillRow
	for rows.Next() {
		var i GetDeploymentCallMetricsNoFillRow
		if err := rows.Scan(
			&i.Timestamp,
			&i.TotalCount,
			&i.SuccessCount,
			&i.AvgTotalTime,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getDeploymentEventMetrics = `-- name: GetDeploymentEventMetrics :many
SELECT
	generate_series::timestamp AS timestamp,
	COALESCE(total_count, 0) AS total_count,
	COALESCE(success_count, 0) AS success_count,
	COALESCE(avg_execution_time, 0) AS avg_execution_time,
	COALESCE(avg_total_time, 0) AS avg_total_time
FROM (
	SELECT
		date_trunc($2::text, timestamp)::timestamp AS trunc_timestamp,
		COUNT(*) AS total_count,
		SUM(
			CASE WHEN event_success = TRUE THEN
				1
			ELSE
				0
			END) AS success_count,
		AVG(event_execution_time) AS avg_execution_time,
		AVG(event_total_time) AS avg_total_time
	FROM
		deployment_metrics
	WHERE
		deployment_metrics.deployment_id = $1
		AND timestamp >= date_trunc($2::text, $3::timestamp)
		AND type = 'EVENT_HANDLED'
	GROUP BY
		trunc_timestamp) AS y
	RIGHT JOIN generate_series(date_trunc($2::text, $3::timestamp), $4::timestamp, ($5::text)::interval) 
ON trunc_timestamp = generate_series
`

type GetDeploymentEventMetricsParams struct {
	DeploymentID string
	Precision    string
	StartAt      pgtype.Timestamp
	EndAt        pgtype.Timestamp
	SeriesStep   string
}

type GetDeploymentEventMetricsRow struct {
	Timestamp        pgtype.Timestamp
	TotalCount       int64
	SuccessCount     int64
	AvgExecutionTime float64
	AvgTotalTime     float64
}

func (q *Queries) GetDeploymentEventMetrics(ctx context.Context, arg GetDeploymentEventMetricsParams) ([]GetDeploymentEventMetricsRow, error) {
	rows, err := q.db.Query(ctx, getDeploymentEventMetrics,
		arg.DeploymentID,
		arg.Precision,
		arg.StartAt,
		arg.EndAt,
		arg.SeriesStep,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetDeploymentEventMetricsRow
	for rows.Next() {
		var i GetDeploymentEventMetricsRow
		if err := rows.Scan(
			&i.Timestamp,
			&i.TotalCount,
			&i.SuccessCount,
			&i.AvgExecutionTime,
			&i.AvgTotalTime,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getDeploymentEventMetricsNoFill = `-- name: GetDeploymentEventMetricsNoFill :many
SELECT 
    date_trunc($3, timestamp)::timestamp as timestamp,
    COUNT(*) AS total_count,
    SUM(CASE WHEN event_success = true THEN 1 ELSE 0 END) AS success_count,
    AVG(event_execution_time) AS avg_execution_time,
    AVG(event_total_time) AS avg_total_time
FROM deployment_metrics 
WHERE 
    deployment_id = $1 AND 
    timestamp >= $2 AND
    type = 'EVENT_HANDLED' 
GROUP BY date_trunc($3, timestamp)
ORDER BY timestamp ASC
`

type GetDeploymentEventMetricsNoFillParams struct {
	DeploymentID string
	Timestamp    pgtype.Timestamp
	DateTrunc    string
}

type GetDeploymentEventMetricsNoFillRow struct {
	Timestamp        pgtype.Timestamp
	TotalCount       int64
	SuccessCount     int64
	AvgExecutionTime float64
	AvgTotalTime     float64
}

func (q *Queries) GetDeploymentEventMetricsNoFill(ctx context.Context, arg GetDeploymentEventMetricsNoFillParams) ([]GetDeploymentEventMetricsNoFillRow, error) {
	rows, err := q.db.Query(ctx, getDeploymentEventMetricsNoFill, arg.DeploymentID, arg.Timestamp, arg.DateTrunc)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetDeploymentEventMetricsNoFillRow
	for rows.Next() {
		var i GetDeploymentEventMetricsNoFillRow
		if err := rows.Scan(
			&i.Timestamp,
			&i.TotalCount,
			&i.SuccessCount,
			&i.AvgExecutionTime,
			&i.AvgTotalTime,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getDeploymentsCallMetrics = `-- name: GetDeploymentsCallMetrics :many
SELECT
	generate_series::timestamp AS timestamp,
	COALESCE(total_count, 0) AS total_count,
	COALESCE(success_count, 0) AS success_count,
	COALESCE(avg_total_time, 0) AS avg_total_time
FROM (
	SELECT
		date_trunc($2::text, timestamp)::timestamp AS trunc_timestamp,
		COUNT(*) AS total_count,
		SUM(
			CASE WHEN call_success = TRUE THEN
				1
			ELSE
				0
			END) AS success_count, AVG(call_total_time) AS avg_total_time
	FROM
		deployment_metrics
	LEFT JOIN 
		deployments ON deployments.id = deployment_metrics.deployment_id
	WHERE
		deployments.app_id = $1
		AND timestamp >= date_trunc($2::text, $3::timestamp)
		AND TYPE = 'CALL_EXECUTED'
	GROUP BY
		trunc_timestamp) AS y
	RIGHT JOIN generate_series(date_trunc($2::text, $3::timestamp), $4::timestamp, ($5::text)::interval) ON trunc_timestamp = generate_series
`

type GetDeploymentsCallMetricsParams struct {
	AppID      string
	Precision  string
	StartAt    pgtype.Timestamp
	EndAt      pgtype.Timestamp
	SeriesStep string
}

type GetDeploymentsCallMetricsRow struct {
	Timestamp    pgtype.Timestamp
	TotalCount   int64
	SuccessCount int64
	AvgTotalTime float64
}

func (q *Queries) GetDeploymentsCallMetrics(ctx context.Context, arg GetDeploymentsCallMetricsParams) ([]GetDeploymentsCallMetricsRow, error) {
	rows, err := q.db.Query(ctx, getDeploymentsCallMetrics,
		arg.AppID,
		arg.Precision,
		arg.StartAt,
		arg.EndAt,
		arg.SeriesStep,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetDeploymentsCallMetricsRow
	for rows.Next() {
		var i GetDeploymentsCallMetricsRow
		if err := rows.Scan(
			&i.Timestamp,
			&i.TotalCount,
			&i.SuccessCount,
			&i.AvgTotalTime,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getDeploymentsEventMetrics = `-- name: GetDeploymentsEventMetrics :many
SELECT
	generate_series::timestamp AS timestamp,
	COALESCE(total_count, 0) AS total_count,
	COALESCE(success_count, 0) AS success_count,
	COALESCE(avg_execution_time, 0) AS avg_execution_time,
	COALESCE(avg_total_time, 0) AS avg_total_time
FROM (
	SELECT
		date_trunc($2::text, timestamp)::timestamp AS trunc_timestamp,
		COUNT(*) AS total_count,
		SUM(
			CASE WHEN event_success = TRUE THEN
				1
			ELSE
				0
			END) AS success_count,
		AVG(event_execution_time) AS avg_execution_time,
		AVG(event_total_time) AS avg_total_time
	FROM
		deployment_metrics
	LEFT JOIN 
		deployments ON deployments.id = deployment_metrics.deployment_id
	WHERE
		deployments.app_id = $1
		AND timestamp >= date_trunc($2::text, $3::timestamp)
		AND TYPE = 'EVENT_HANDLED'
	GROUP BY
		trunc_timestamp
) AS y
	RIGHT JOIN generate_series(date_trunc($2::text, $3::timestamp), $4::timestamp, ($5::text)::interval) ON trunc_timestamp = generate_series
`

type GetDeploymentsEventMetricsParams struct {
	AppID      string
	Precision  string
	StartAt    pgtype.Timestamp
	EndAt      pgtype.Timestamp
	SeriesStep string
}

type GetDeploymentsEventMetricsRow struct {
	Timestamp        pgtype.Timestamp
	TotalCount       int64
	SuccessCount     int64
	AvgExecutionTime float64
	AvgTotalTime     float64
}

func (q *Queries) GetDeploymentsEventMetrics(ctx context.Context, arg GetDeploymentsEventMetricsParams) ([]GetDeploymentsEventMetricsRow, error) {
	rows, err := q.db.Query(ctx, getDeploymentsEventMetrics,
		arg.AppID,
		arg.Precision,
		arg.StartAt,
		arg.EndAt,
		arg.SeriesStep,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetDeploymentsEventMetricsRow
	for rows.Next() {
		var i GetDeploymentsEventMetricsRow
		if err := rows.Scan(
			&i.Timestamp,
			&i.TotalCount,
			&i.SuccessCount,
			&i.AvgExecutionTime,
			&i.AvgTotalTime,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getDeploymentsMetricsSummary = `-- name: GetDeploymentsMetricsSummary :one
SELECT
	MIN(timestamp)::timestamp AS first_entry_at,
	MAX(timestamp)::timestamp AS last_entry_at,

	COUNT(*) FILTER (WHERE type = 'EVENT_HANDLED') AS total_event_count,
	COALESCE(SUM(
		CASE WHEN event_success = TRUE THEN
			1
		ELSE
			0
		END)  FILTER (WHERE type = 'EVENT_HANDLED'), 0)::bigint  AS success_event_count,
	COALESCE(AVG(event_execution_time) FILTER (WHERE type = 'EVENT_HANDLED'), 0)::double precision AS avg_event_execution_time,
	COALESCE(SUM(event_execution_time) FILTER (WHERE type = 'EVENT_HANDLED'), 0)::bigint AS total_event_execution_time,
	COALESCE(AVG(event_total_time) FILTER (WHERE type = 'EVENT_HANDLED'), 0)::double precision  AS avg_event_total_time,
	COALESCE(SUM(event_total_time) FILTER (WHERE type = 'EVENT_HANDLED'), 0::bigint)::bigint AS total_event_total_time,

	COUNT(*) FILTER (WHERE type = 'CALL_EXECUTED') AS total_call_count,
	COALESCE(SUM(
		CASE WHEN call_success = TRUE THEN
			1
		ELSE
			0
		END) FILTER (WHERE type = 'CALL_EXECUTED'), 0)::bigint AS success_call_count,
	COALESCE(AVG(call_total_time) FILTER (WHERE type = 'CALL_EXECUTED'), 0)::double precision AS avg_call_total_time,
	COALESCE(SUM(call_total_time) FILTER (WHERE type = 'CALL_EXECUTED'), 0)::bigint AS total_call_total_time
FROM
	deployment_metrics
LEFT JOIN 
	deployments ON deployments.id = deployment_metrics.deployment_id
WHERE
	app_id = $1 AND 
	timestamp >= $2 AND 
	timestamp <= $3
GROUP BY
	app_id
`

type GetDeploymentsMetricsSummaryParams struct {
	AppID   string
	StartAt pgtype.Timestamp
	EndAt   pgtype.Timestamp
}

type GetDeploymentsMetricsSummaryRow struct {
	FirstEntryAt            pgtype.Timestamp
	LastEntryAt             pgtype.Timestamp
	TotalEventCount         int64
	SuccessEventCount       int64
	AvgEventExecutionTime   float64
	TotalEventExecutionTime int64
	AvgEventTotalTime       float64
	TotalEventTotalTime     int64
	TotalCallCount          int64
	SuccessCallCount        int64
	AvgCallTotalTime        float64
	TotalCallTotalTime      int64
}

func (q *Queries) GetDeploymentsMetricsSummary(ctx context.Context, arg GetDeploymentsMetricsSummaryParams) (GetDeploymentsMetricsSummaryRow, error) {
	row := q.db.QueryRow(ctx, getDeploymentsMetricsSummary, arg.AppID, arg.StartAt, arg.EndAt)
	var i GetDeploymentsMetricsSummaryRow
	err := row.Scan(
		&i.FirstEntryAt,
		&i.LastEntryAt,
		&i.TotalEventCount,
		&i.SuccessEventCount,
		&i.AvgEventExecutionTime,
		&i.TotalEventExecutionTime,
		&i.AvgEventTotalTime,
		&i.TotalEventTotalTime,
		&i.TotalCallCount,
		&i.SuccessCallCount,
		&i.AvgCallTotalTime,
		&i.TotalCallTotalTime,
	)
	return i, err
}
