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
) u ON d.dt = u.date
ORDER BY d.dt;

-- Only enabled apps and apps stopped for running out of credits, which the
-- credit sweep starts again once they are under their limit. Other disabled
-- apps would only add work, their usage rows stay for the rest of the month.
-- name: GetAllUsageCreditsUsedBetween :many
SELECT u.app_id, SUM(u.credits_used) FROM usage_records u
JOIN apps a ON a.id = u.app_id AND (a.enabled OR a.disabled_reason = @no_credits_reason)
WHERE u.created_at BETWEEN @start_at AND @end_at
GROUP BY u.app_id;

-- name: DeleteUsageRecordsBefore :execrows
-- Batched so a large backlog doesn't hold one long transaction.
DELETE FROM usage_records WHERE id IN (
    SELECT expired.id FROM usage_records expired
    WHERE expired.created_at < @before_at
    LIMIT @batch_size
);

-- name: HasAppCommandUsage :one
SELECT EXISTS(SELECT 1 FROM usage_records WHERE app_id = $1 AND command_id IS NOT NULL);
