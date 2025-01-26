-- name: GetEntitlements :many
SELECT * FROM entitlements WHERE app_id = $1;

-- name: UpsertSubscriptionEntitlement :one
INSERT INTO entitlements (
    id,
    type,
    subscription_id,
    app_id,
    feature_usage_credits_per_month,
    feature_max_collaborator,
    created_at,
    updated_at,
    ends_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) 
ON CONFLICT (subscription_id) DO UPDATE SET 
    feature_usage_credits_per_month = EXCLUDED.feature_usage_credits_per_month,
    feature_max_collaborator = EXCLUDED.feature_max_collaborator,
    updated_at = EXCLUDED.updated_at,
    ends_at = EXCLUDED.ends_at
RETURNING *;
