-- name: CreateAppUsageEntry :exec
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
);

-- name: GetLastAppUsageEntry :one
SELECT * FROM app_usage WHERE app_id = $1 ORDER BY period_ends_at DESC LIMIT 1;

-- name: GetAppUsageSummary :one
SELECT * FROM app_usage_month_view WHERE app_id = $1;

-- name: GetAppUsageAndLimits :one
SELECT * FROM app_entitlements_resolved_view 
LEFT JOIN app_usage_month_view  
ON app_usage_month_view.app_id = app_entitlements_resolved_view.app_id 
WHERE app_entitlements_resolved_view.app_id = $1;