// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.24.0
// source: deployment_metrics.sql

package pgmodel

import (
	"context"
	"time"

	"github.com/sqlc-dev/pqtype"
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
	Metadata           pqtype.NullRawMessage
	EventType          string
	EventSuccess       bool
	EventExecutionTime int64
	EventTotalTime     int64
	CallType           string
	CallSuccess        bool
	CallTotalTime      int64
	Timestamp          time.Time
}

func (q *Queries) CreateDeploymentMetricEntry(ctx context.Context, arg CreateDeploymentMetricEntryParams) error {
	_, err := q.db.ExecContext(ctx, createDeploymentMetricEntry,
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
	StartAt      time.Time
	EndAt        time.Time
	SeriesStep   string
}

type GetDeploymentCallMetricsRow struct {
	Timestamp    time.Time
	TotalCount   int64
	SuccessCount int64
	AvgTotalTime float64
}

func (q *Queries) GetDeploymentCallMetrics(ctx context.Context, arg GetDeploymentCallMetricsParams) ([]GetDeploymentCallMetricsRow, error) {
	rows, err := q.db.QueryContext(ctx, getDeploymentCallMetrics,
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
	if err := rows.Close(); err != nil {
		return nil, err
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
	Timestamp    time.Time
	DateTrunc    string
}

type GetDeploymentCallMetricsNoFillRow struct {
	Timestamp    time.Time
	TotalCount   int64
	SuccessCount int64
	AvgTotalTime float64
}

func (q *Queries) GetDeploymentCallMetricsNoFill(ctx context.Context, arg GetDeploymentCallMetricsNoFillParams) ([]GetDeploymentCallMetricsNoFillRow, error) {
	rows, err := q.db.QueryContext(ctx, getDeploymentCallMetricsNoFill, arg.DeploymentID, arg.Timestamp, arg.DateTrunc)
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
	if err := rows.Close(); err != nil {
		return nil, err
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
		AND TYPE = 'EVENT_HANDLED'
	GROUP BY
		trunc_timestamp) AS y
	RIGHT JOIN generate_series(date_trunc($2::text, $3::timestamp), $4::timestamp, ($5::text)::interval) 
ON trunc_timestamp = generate_series
`

type GetDeploymentEventMetricsParams struct {
	DeploymentID string
	Precision    string
	StartAt      time.Time
	EndAt        time.Time
	SeriesStep   string
}

type GetDeploymentEventMetricsRow struct {
	Timestamp        time.Time
	TotalCount       int64
	SuccessCount     int64
	AvgExecutionTime float64
	AvgTotalTime     float64
}

func (q *Queries) GetDeploymentEventMetrics(ctx context.Context, arg GetDeploymentEventMetricsParams) ([]GetDeploymentEventMetricsRow, error) {
	rows, err := q.db.QueryContext(ctx, getDeploymentEventMetrics,
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
	if err := rows.Close(); err != nil {
		return nil, err
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
	Timestamp    time.Time
	DateTrunc    string
}

type GetDeploymentEventMetricsNoFillRow struct {
	Timestamp        time.Time
	TotalCount       int64
	SuccessCount     int64
	AvgExecutionTime float64
	AvgTotalTime     float64
}

func (q *Queries) GetDeploymentEventMetricsNoFill(ctx context.Context, arg GetDeploymentEventMetricsNoFillParams) ([]GetDeploymentEventMetricsNoFillRow, error) {
	rows, err := q.db.QueryContext(ctx, getDeploymentEventMetricsNoFill, arg.DeploymentID, arg.Timestamp, arg.DateTrunc)
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
	if err := rows.Close(); err != nil {
		return nil, err
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
		deployments.guild_id = $1
		AND timestamp >= date_trunc($2::text, $3::timestamp)
		AND TYPE = 'CALL_EXECUTED'
	GROUP BY
		trunc_timestamp) AS y
	RIGHT JOIN generate_series(date_trunc($2::text, $3::timestamp), $4::timestamp, ($5::text)::interval) ON trunc_timestamp = generate_series
`

type GetDeploymentsCallMetricsParams struct {
	GuildID    string
	Precision  string
	StartAt    time.Time
	EndAt      time.Time
	SeriesStep string
}

type GetDeploymentsCallMetricsRow struct {
	Timestamp    time.Time
	TotalCount   int64
	SuccessCount int64
	AvgTotalTime float64
}

func (q *Queries) GetDeploymentsCallMetrics(ctx context.Context, arg GetDeploymentsCallMetricsParams) ([]GetDeploymentsCallMetricsRow, error) {
	rows, err := q.db.QueryContext(ctx, getDeploymentsCallMetrics,
		arg.GuildID,
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
	if err := rows.Close(); err != nil {
		return nil, err
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
		deployments.guild_id = $1
		AND timestamp >= date_trunc($2::text, $3::timestamp)
		AND TYPE = 'EVENT_HANDLED'
	GROUP BY
		trunc_timestamp
) AS y
	RIGHT JOIN generate_series(date_trunc($2::text, $3::timestamp), $4::timestamp, ($5::text)::interval) ON trunc_timestamp = generate_series
`

type GetDeploymentsEventMetricsParams struct {
	GuildID    string
	Precision  string
	StartAt    time.Time
	EndAt      time.Time
	SeriesStep string
}

type GetDeploymentsEventMetricsRow struct {
	Timestamp        time.Time
	TotalCount       int64
	SuccessCount     int64
	AvgExecutionTime float64
	AvgTotalTime     float64
}

func (q *Queries) GetDeploymentsEventMetrics(ctx context.Context, arg GetDeploymentsEventMetricsParams) ([]GetDeploymentsEventMetricsRow, error) {
	rows, err := q.db.QueryContext(ctx, getDeploymentsEventMetrics,
		arg.GuildID,
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
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
