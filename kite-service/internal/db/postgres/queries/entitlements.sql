-- name: GetEntitlements :many
SELECT * FROM entitlements WHERE app_id = $1 ORDER BY created_at DESC;

-- name: GetActiveEntitlements :many
SELECT * FROM entitlements WHERE app_id = $1 AND (ends_at IS NULL OR ends_at > $2) ORDER BY created_at DESC;

-- Batch form of GetActiveEntitlements. The usage manager checks every app with
-- usage this month, which was one round trip each.
-- name: GetActiveEntitlementsForApps :many
SELECT * FROM entitlements
WHERE app_id = ANY(@app_ids::text[]) AND (ends_at IS NULL OR ends_at > @ends_at)
ORDER BY created_at DESC;

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

-- Updates every entitlement held by the subscription. :exec, not :one, because
-- a webhook can arrive for a subscription that has no entitlement yet (no
-- app_id in the checkout metadata), and no rows to update is not an error.
-- name: UpdateSubscriptionEntitlement :exec
UPDATE entitlements SET
    plan_id = $2,
    updated_at = $3,
    ends_at = $4
WHERE subscription_id = $1;
