-- name: CreateAPIToken :one
INSERT INTO api_tokens (
    id,
    user_id,
    name,
    key_hash,
    scopes,
    expires_at,
    created_at
) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING *;

-- name: GetAPITokensByUser :many
SELECT * FROM api_tokens WHERE user_id = $1 ORDER BY created_at DESC;

-- name: CountAPITokensByUser :one
-- Expired tokens don't count, they only stay so users see they expired.
SELECT COUNT(*) FROM api_tokens
WHERE user_id = @user_id AND (expires_at IS NULL OR expires_at > @now);

-- name: GetAPITokenByKeyHash :one
SELECT * FROM api_tokens WHERE key_hash = $1;

-- name: UpdateAPITokenLastUsedAt :exec
-- Requests in parallel may all find it due, only the first one writes it.
UPDATE api_tokens SET last_used_at = @last_used_at
WHERE id = @id AND (last_used_at IS NULL OR last_used_at < @last_used_before);

-- name: DeleteAPIToken :execrows
DELETE FROM api_tokens WHERE id = @id AND user_id = @user_id;
