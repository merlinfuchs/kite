-- name: CreateGuildUsageEntry :exec
INSERT INTO guild_usage (
    guild_id,
    total_event_count,
    success_event_count,
    total_event_execution_time,
    avg_event_execution_time,
    total_event_total_time,
    avg_event_total_time,
    total_call_count,
    success_call_count,
    total_call_total_time,
    avg_call_total_time,
    period_starts_at,
    period_ends_at
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
);

-- name: GetLastGuildUsageEntry :one
SELECT * FROM guild_usage WHERE guild_id = $1 ORDER BY period_ends_at DESC LIMIT 1;

-- name: GetGuildUsageSummary :one
SELECT * FROM guild_usage_month_view WHERE guild_id = $1;

-- name: GetGuildUsageAndLimits :one
SELECT * FROM guild_entitlements_resolved_view 
LEFT JOIN guild_usage_month_view  
ON guild_usage_month_view.guild_id = guild_entitlements_resolved_view.guild_id 
WHERE guild_entitlements_resolved_view.guild_id = $1;