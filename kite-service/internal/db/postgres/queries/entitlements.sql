-- name: GetEntitlements :many
SELECT * FROM entitlements WHERE app_id = $1 ORDER BY created_at DESC;

-- name: GetActiveEntitlements :many
SELECT * FROM entitlements WHERE app_id = $1 AND (ends_at IS NULL OR ends_at > $2) ORDER BY created_at DESC;

-- name: UpsertSubscriptionEntitlement :one
INSERT INTO entitlements (
    id,
    type,
    subscription_id,
    app_id,
    plan_id,
    created_at,
    updated_at,
    ends_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8) 
ON CONFLICT (subscription_id, app_id) DO UPDATE SET 
    plan_id = EXCLUDED.plan_id,
    updated_at = EXCLUDED.updated_at,
    ends_at = EXCLUDED.ends_at
RETURNING *;

-- name: UpdateSubscriptionEntitlement :one
UPDATE entitlements SET
    plan_id = $2,
    updated_at = $3,
    ends_at = $4
WHERE subscription_id = $1 RETURNING *;
