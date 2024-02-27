-- name: UpsertGuildEntitlement :one
INSERT INTO guild_entitlements (
    id,
    guild_id,
    user_id,
    source,
    source_id,
    name,
    description,
    feature_monthly_execution_time_limit,
    feature_monthly_execution_time_additive,
    created_at,
    updated_at,
    valid_from,
    valid_until
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    $9,
    $10,
    $11,
    $12,
    $13
) ON CONFLICT (guild_id, source, source_id) DO UPDATE SET 
    feature_monthly_execution_time_limit = EXCLUDED.feature_monthly_execution_time_limit,
    feature_monthly_execution_time_additive = EXCLUDED.feature_monthly_execution_time_additive,
    updated_at = EXCLUDED.updated_at,
    valid_from = EXCLUDED.valid_from,
    valid_until = EXCLUDED.valid_until
RETURNING *;

-- name: GetGuildEntitlements :many
SELECT * FROM guild_entitlements 
WHERE 
    guild_id = $1 AND
    (valid_from IS NULL OR valid_from <= @valid_at) AND 
    (valid_until IS NULL OR valid_until >= @valid_at);

-- name: GetResolvedGuildEntitlement :one
SELECT * FROM guild_entitlements_resolved_view WHERE guild_id = $1;