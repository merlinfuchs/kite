-- name: GetSubscriptions :many
SELECT * FROM subscriptions WHERE user_id = $1 ORDER BY created_at DESC;

-- name: GetSubscriptionsByAppID :many
SELECT subscriptions.* FROM subscriptions LEFT JOIN entitlements ON subscriptions.id = entitlements.subscription_id WHERE entitlements.app_id = $1 ORDER BY subscriptions.created_at DESC;

-- name: GetAllSubscriptions :many
SELECT * FROM subscriptions ORDER BY created_at DESC;

-- name: GetSubscription :one
SELECT * FROM subscriptions WHERE id = $1;

-- name: UpsertLemonSqueezySubscription :one
INSERT INTO subscriptions (
    id,
    display_name,
    source,
    status,
    status_formatted,
    renews_at,
    trial_ends_at,
    ends_at,
    created_at,
    updated_at,
    user_id,
    lemonsqueezy_subscription_id,
    lemonsqueezy_customer_id,
    lemonsqueezy_order_id,
    lemonsqueezy_product_id,
    lemonsqueezy_variant_id
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
ON CONFLICT (lemonsqueezy_subscription_id) DO UPDATE SET 
    display_name = EXCLUDED.display_name,
    status = EXCLUDED.status,
    status_formatted = EXCLUDED.status_formatted,
    renews_at = EXCLUDED.renews_at,
    trial_ends_at = EXCLUDED.trial_ends_at,
    ends_at = EXCLUDED.ends_at,
    updated_at = EXCLUDED.updated_at,
    lemonsqueezy_customer_id = EXCLUDED.lemonsqueezy_customer_id,
    lemonsqueezy_order_id = EXCLUDED.lemonsqueezy_order_id,
    lemonsqueezy_product_id = EXCLUDED.lemonsqueezy_product_id,
    lemonsqueezy_variant_id = EXCLUDED.lemonsqueezy_variant_id
RETURNING *;
