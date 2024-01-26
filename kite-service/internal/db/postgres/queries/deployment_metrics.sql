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