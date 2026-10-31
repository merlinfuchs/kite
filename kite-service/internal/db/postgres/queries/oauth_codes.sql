-- name: CreateOAuthCode :exec
INSERT INTO oauth_codes (
    code_hash,
    user_id,
    client_id,
    client_name,
    redirect_uri,
    code_challenge,
    scopes,
    app_ids,
    resource,
    expires_at,
    created_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);

-- name: UseOAuthCode :one
-- Only one of two uses of a code in parallel finds it.
UPDATE oauth_codes SET api_token_id = @api_token_id
WHERE code_hash = @code_hash AND api_token_id IS NULL AND expires_at > @now
RETURNING *;

-- name: GetOAuthCode :one
SELECT * FROM oauth_codes WHERE code_hash = $1;

-- name: DeleteExpiredOAuthCodes :execrows
DELETE FROM oauth_codes WHERE code_hash IN (
    SELECT expired.code_hash FROM oauth_codes expired
    WHERE expired.expires_at < @expired_before
    LIMIT @batch_size
);
