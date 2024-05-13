-- name: GetSession :one
SELECT * FROM sessions WHERE token_hash = $1;

-- name: DeleteSession :exec
DELETE FROM sessions WHERE token_hash = $1;

-- name: CreateSession :exec
INSERT INTO sessions (
    token_hash,
    type,
    user_id,
    access_token,
    created_at,
    expires_at
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
);

-- name: CreatePendingSession :exec
INSERT INTO sessions_pending (
    code,
    created_at,
    expires_at
) VALUES (
    $1,
    $2,
    $3
);

-- name: UpdatePendingSession :one
UPDATE sessions_pending SET token = $1, expires_at = $2 WHERE code = $3 RETURNING *;

-- name: GetPendingSession :one
SELECT * FROM sessions_pending WHERE code = $1;

-- name: DeleteExpiredPendingSessions :exec
DELETE FROM sessions_pending WHERE expires_at < $1;