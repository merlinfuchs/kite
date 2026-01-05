-- name: CreateLogEntry :exec
INSERT INTO logs (app_id, message, level, command_id, event_listener_id, message_id, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING *;

-- name: GetLogEntriesByApp :many
SELECT * FROM logs 
WHERE 
    app_id = $1 AND 
    (sqlc.narg(before_id)::bigint IS NULL OR id < sqlc.narg(before_id)::bigint) 
ORDER BY created_at DESC LIMIT $2;

-- name: GetLogEntriesByCommand :many
SELECT * FROM logs 
WHERE 
    app_id = $1 AND 
    command_id = $2 AND 
    (sqlc.narg(before_id)::bigint IS NULL OR id < sqlc.narg(before_id)::bigint) 
ORDER BY created_at DESC LIMIT $3;

-- name: GetLogEntriesByEvent :many
SELECT * FROM logs 
WHERE 
    app_id = $1 AND 
    event_listener_id = $2 AND 
    (sqlc.narg(before_id)::bigint IS NULL OR id < sqlc.narg(before_id)::bigint) 
ORDER BY created_at DESC LIMIT $3;

-- name: GetLogEntriesByMessage :many
SELECT * FROM logs WHERE app_id = $1 AND message_id = $2 AND (sqlc.narg(before_id)::bigint IS NULL OR id < sqlc.narg(before_id)::bigint) ORDER BY created_at DESC LIMIT $3;

-- name: GetLogSummary :one
SELECT COUNT(*) AS total_entries,
       COALESCE(SUM(CASE WHEN level = 'error' THEN 1 ELSE 0 END), 0)::bigint AS total_errors,
       COALESCE(SUM(CASE WHEN level = 'warn' THEN 1 ELSE 0 END), 0)::bigint AS total_warnings,
       COALESCE(SUM(CASE WHEN level = 'info' THEN 1 ELSE 0 END), 0)::bigint AS total_infos,
       COALESCE(SUM(CASE WHEN level = 'debug' THEN 1 ELSE 0 END), 0)::bigint AS total_debugs
FROM logs WHERE app_id = @app_id AND created_at >= @start_at AND created_at < @end_at;

-- name: DeleteLogEntriesBefore :exec
DELETE FROM logs WHERE created_at < @before_at;
