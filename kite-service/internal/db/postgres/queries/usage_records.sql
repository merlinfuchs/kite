-- name: CreateUsageRecord :exec
INSERT INTO usage_records (
    type,
    app_id,
    command_id,
    event_listener_id,
    message_id,
    credits_used,
    created_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: GetUsageRecordsByAppBetween :many
SELECT * FROM usage_records WHERE app_id = @app_id AND created_at BETWEEN @start_at AND @end_at ORDER BY created_at DESC;

-- name: GetUsageCreditsUsedByAppBetween :one
SELECT COALESCE(SUM(credits_used), 0)::int FROM usage_records WHERE app_id = @app_id AND created_at BETWEEN @start_at AND @end_at;

-- name: GetUsageCreditsUsedByTypeBetween :many
SELECT type, SUM(credits_used) FROM usage_records WHERE app_id = @app_id AND created_at BETWEEN @start_at AND @end_at GROUP BY type;

-- name: GetUsageCreditsUsedByDayBetween :many
SELECT 
    d.dt as date, 
    coalesce(u.credits_used, 0) as credits_used 
FROM (
    SELECT dt::date 
    FROM generate_series(@start_at::timestamp, @end_at::timestamp, '1 day'::interval) dt
) d
LEFT JOIN (
    SELECT DATE(created_at) as date, SUM(credits_used) as credits_used 
    FROM usage_records 
    WHERE app_id = @app_id AND created_at BETWEEN @start_at AND @end_at 
    GROUP BY DATE(created_at)
) u ON d.dt = u.date;

-- name: GetAllUsageCreditsUsedBetween :many
SELECT app_id, SUM(credits_used) FROM usage_records WHERE created_at BETWEEN @start_at AND @end_at GROUP BY app_id;

-- name: DeleteUsageRecordsBefore :exec
DELETE FROM usage_records WHERE created_at < @before_at;
