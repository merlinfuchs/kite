-- name: CreateLogEntry :exec
INSERT INTO logs (app_id, message, level, command_id, event_listener_id, message_id, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING *;

-- name: GetLogEntriesByApp :many
SELECT * FROM logs WHERE app_id = $1 ORDER BY created_at DESC LIMIT $2;

-- name: GetLogEntriesByAppBefore :many
SELECT * FROM logs WHERE app_id = $1 AND id < $2 ORDER BY created_at DESC LIMIT $3;

-- name: GetLogSummary :one
SELECT COUNT(*) AS total_entries,
       SUM(CASE WHEN level = 'error' THEN 1 ELSE 0 END) AS total_errors,
       SUM(CASE WHEN level = 'warn' THEN 1 ELSE 0 END) AS total_warnings,
       SUM(CASE WHEN level = 'info' THEN 1 ELSE 0 END) AS total_infos,
       SUM(CASE WHEN level = 'debug' THEN 1 ELSE 0 END) AS total_debugs
FROM logs WHERE app_id = @app_id AND created_at >= @start_at AND created_at < @end_at;
