-- name: GetSession :one
SELECT * FROM sessions WHERE token_hash = $1;

-- name: DeleteSession :exec
DELETE FROM sessions WHERE token_hash = $1;

-- name: CreateSession :exec
INSERT INTO sessions (
    token_hash,
    type,
    user_id,
    guild_ids,
    access_token,
    retrieval_code,
    created_at,
    expires_at
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7
);