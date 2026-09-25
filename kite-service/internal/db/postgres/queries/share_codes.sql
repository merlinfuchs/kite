-- name: CreateShareCode :exec
INSERT INTO share_codes (
    code,
    type,
    data,
    creator_user_id,
    app_id,
    created_at,
    last_used_at
) VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: ShareCode :one
SELECT * FROM share_codes WHERE code = $1;

-- name: TouchShareCode :exec
UPDATE share_codes SET last_used_at = @used_at WHERE code = @code;

-- name: DeleteUnusedShareCodes :execrows
-- Batched so a large backlog doesn't hold one long transaction.
DELETE FROM share_codes WHERE code IN (
    SELECT stale.code FROM share_codes stale
    WHERE stale.last_used_at < @used_before
    LIMIT @batch_size
);
