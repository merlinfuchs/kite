-- name: CreateDeploymentMetricEntry :exec
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
);

-- name: GetDeploymentEventMetrics :many
SELECT
	generate_series::timestamp AS timestamp,
	COALESCE(total_count, 0) AS total_count,
	COALESCE(success_count, 0) AS success_count,
	COALESCE(avg_execution_time, 0) AS avg_execution_time,
	COALESCE(avg_total_time, 0) AS avg_total_time
FROM (
	SELECT
		date_trunc(sqlc.arg (precision)::text, timestamp)::timestamp AS trunc_timestamp,
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
		AND timestamp >= date_trunc(sqlc.arg (precision)::text, sqlc.arg (start_at)::timestamp)
		AND TYPE = 'EVENT_HANDLED'
	GROUP BY
		trunc_timestamp) AS y
	RIGHT JOIN generate_series(date_trunc(sqlc.arg (precision)::text, sqlc.arg (start_at)::timestamp), sqlc.arg (end_at)::timestamp, (sqlc.arg (series_step)::text)::interval) 
ON trunc_timestamp = generate_series;

-- name: GetDeploymentEventMetricsNoFill :many
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
ORDER BY timestamp ASC;

-- name: GetDeploymentCallMetrics :many
SELECT
	generate_series::timestamp AS timestamp,
	COALESCE(total_count, 0) AS total_count,
	COALESCE(success_count, 0) AS success_count,
	COALESCE(avg_total_time, 0) AS avg_total_time
FROM (
	SELECT
		date_trunc(sqlc.arg (precision)::text, timestamp)::timestamp AS trunc_timestamp,
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
		AND timestamp >= date_trunc(sqlc.arg (precision)::text, sqlc.arg (start_at)::timestamp)
		AND TYPE = 'CALL_EXECUTED'
	GROUP BY
		trunc_timestamp) AS y
	RIGHT JOIN generate_series(date_trunc(sqlc.arg (precision)::text, sqlc.arg (start_at)::timestamp), sqlc.arg (end_at)::timestamp, (sqlc.arg (series_step)::text)::interval) 
ON trunc_timestamp = generate_series;

-- name: GetDeploymentCallMetricsNoFill :many
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
ORDER BY timestamp ASC;

-- name: GetDeploymentsEventMetrics :many
SELECT
	generate_series::timestamp AS timestamp,
	COALESCE(total_count, 0) AS total_count,
	COALESCE(success_count, 0) AS success_count,
	COALESCE(avg_execution_time, 0) AS avg_execution_time,
	COALESCE(avg_total_time, 0) AS avg_total_time
FROM (
	SELECT
		date_trunc(sqlc.arg (precision)::text, timestamp)::timestamp AS trunc_timestamp,
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
		AND timestamp >= date_trunc(sqlc.arg (precision)::text, sqlc.arg (start_at)::timestamp)
		AND TYPE = 'EVENT_HANDLED'
	GROUP BY
		trunc_timestamp
) AS y
	RIGHT JOIN generate_series(date_trunc(sqlc.arg (precision)::text, sqlc.arg (start_at)::timestamp), sqlc.arg (end_at)::timestamp, (sqlc.arg (series_step)::text)::interval) ON trunc_timestamp = generate_series;

-- name: GetDeploymentsCallMetrics :many
SELECT
	generate_series::timestamp AS timestamp,
	COALESCE(total_count, 0) AS total_count,
	COALESCE(success_count, 0) AS success_count,
	COALESCE(avg_total_time, 0) AS avg_total_time
FROM (
	SELECT
		date_trunc(sqlc.arg (precision)::text, timestamp)::timestamp AS trunc_timestamp,
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
		AND timestamp >= date_trunc(sqlc.arg (precision)::text, sqlc.arg (start_at)::timestamp)
		AND TYPE = 'CALL_EXECUTED'
	GROUP BY
		trunc_timestamp) AS y
	RIGHT JOIN generate_series(date_trunc(sqlc.arg (precision)::text, sqlc.arg (start_at)::timestamp), sqlc.arg (end_at)::timestamp, (sqlc.arg (series_step)::text)::interval) ON trunc_timestamp = generate_series;