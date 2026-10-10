-- name: GetCreditLimitsByApp :many
SELECT * FROM credit_limits WHERE app_id = $1 ORDER BY scope, target_id NULLS FIRST, period;

-- name: GetCreditLimit :one
SELECT * FROM credit_limits WHERE app_id = $1 AND id = $2;

-- name: CountCreditLimitsByApp :one
SELECT COUNT(*) FROM credit_limits WHERE app_id = $1;

-- name: CreateCreditLimit :one
INSERT INTO credit_limits (
    id,
    app_id,
    scope,
    target_id,
    period,
    credits,
    message,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING *;

-- name: UpdateCreditLimit :one
UPDATE credit_limits SET
    scope = $3,
    target_id = $4,
    period = $5,
    credits = $6,
    message = $7,
    updated_at = $8
WHERE app_id = $1 AND id = $2 RETURNING *;

-- name: DeleteCreditLimit :execrows
DELETE FROM credit_limits WHERE app_id = $1 AND id = $2;

-- name: GetCreditLimitSettings :one
SELECT * FROM credit_limit_settings WHERE app_id = $1;

-- name: UpsertCreditLimitSettings :one
INSERT INTO credit_limit_settings (
    app_id,
    message,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4
)
ON CONFLICT (app_id) DO UPDATE SET
    message = EXCLUDED.message,
    updated_at = EXCLUDED.updated_at
RETURNING *;
