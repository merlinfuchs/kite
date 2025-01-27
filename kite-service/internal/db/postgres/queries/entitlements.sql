-- name: GetEntitlements :many
SELECT * FROM entitlements WHERE app_id = $1;

-- name: GetEntitlementsWithSubscription :many
SELECT sqlc.embed(entitlements), sqlc.embed(subscriptions) FROM entitlements 
LEFT JOIN subscriptions ON entitlements.subscription_id = subscriptions.id 
WHERE entitlements.app_id = $1;

-- name: UpsertSubscriptionEntitlement :one
INSERT INTO entitlements (
    id,
    type,
    subscription_id,
    app_id,
    feature_set_id,
    created_at,
    updated_at,
    ends_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8) 
ON CONFLICT (subscription_id) DO UPDATE SET 
    feature_set_id = EXCLUDED.feature_set_id,
    updated_at = EXCLUDED.updated_at,
    ends_at = EXCLUDED.ends_at
RETURNING *;

-- name: UpdateSubscriptionEntitlement :one
UPDATE entitlements SET
    feature_set_id = $2,
    updated_at = $3,
    ends_at = $4
WHERE subscription_id = $1 RETURNING *;
