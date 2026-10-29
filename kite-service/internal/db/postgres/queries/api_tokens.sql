-- name: CreateAPIToken :one
INSERT INTO api_tokens (
    id,
    user_id,
    name,
    key_hash,
    scopes,
    app_ids,
    expires_at,
    created_at,
    oauth_client_id,
    oauth_resource,
    refresh_key_hash,
    access_expires_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) RETURNING *;

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

-- name: RefreshAPIToken :one
-- Replaces both keys, so the old refresh token can't be used again. Only one
-- of two refreshes with the same token in parallel finds it.
UPDATE api_tokens SET
    key_hash = @key_hash,
    refresh_key_hash = @new_refresh_key_hash,
    access_expires_at = @access_expires_at,
    expires_at = @expires_at
WHERE refresh_key_hash = @refresh_key_hash
    AND oauth_client_id = @oauth_client_id
    AND expires_at > @now
RETURNING *;

-- name: DeleteExpiredOAuthAPITokens :execrows
-- Their refresh tokens expired, so clients have to connect again, which makes
-- a new token.
DELETE FROM api_tokens WHERE id IN (
    SELECT expired.id FROM api_tokens expired
    WHERE expired.oauth_client_id IS NOT NULL AND expired.expires_at < @expired_before
    LIMIT @batch_size
);
