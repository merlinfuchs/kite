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
SELECT * FROM usage_records WHERE app_id = $1 AND created_at BETWEEN $2 AND $3 ORDER BY created_at DESC;

-- name: GetUsageCreditsUsedByAppBetween :one
SELECT SUM(credits_used) FROM usage_records WHERE app_id = $1 AND created_at BETWEEN $2 AND $3;

-- name: GetAllUsageCreditsUsedBetween :many
SELECT app_id, SUM(credits_used) FROM usage_records WHERE created_at BETWEEN $1 AND $2 GROUP BY app_id;
