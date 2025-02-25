-- name: GetEntitlements :many
SELECT * FROM entitlements WHERE app_id = $1 ORDER BY created_at DESC;

-- name: UpsertSubscriptionEntitlement :one
INSERT INTO entitlements (
    id,
    type,
    subscription_id,
    app_id,
    lemonsqueezy_product_id,
    lemonsqueezy_variant_id,
    created_at,
    updated_at,
    ends_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) 
ON CONFLICT (subscription_id, app_id) DO UPDATE SET 
    lemonsqueezy_product_id = EXCLUDED.lemonsqueezy_product_id,
    lemonsqueezy_variant_id = EXCLUDED.lemonsqueezy_variant_id,
    updated_at = EXCLUDED.updated_at,
    ends_at = EXCLUDED.ends_at
RETURNING *;

-- name: UpdateSubscriptionEntitlement :one
UPDATE entitlements SET
    lemonsqueezy_product_id = $2,
    lemonsqueezy_variant_id = $3,
    updated_at = $4,
    ends_at = $5
WHERE subscription_id = $1 RETURNING *;
