-- name: CreateSession :one
INSERT INTO sessions (key_hash, user_id, created_at, expires_at) VALUES ($1, $2, $3, $4) RETURNING *;

-- name: GetSession :one
SELECT * FROM sessions WHERE key_hash = $1;

-- name: DeleteSession :exec
DELETE FROM sessions WHERE key_hash = $1;