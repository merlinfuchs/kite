-- name: CreateLogEntry :exec
INSERT INTO logs (app_id, message, level, created_at) VALUES ($1, $2, $3, $4) RETURNING *;

-- name: GetLogEntriesByApp :many
SELECT * FROM logs WHERE app_id = $1 ORDER BY created_at DESC LIMIT $2;

-- name: GetLogEntriesByAppBefore :many
SELECT * FROM logs WHERE app_id = $1 AND id < $2 ORDER BY created_at DESC LIMIT $3;