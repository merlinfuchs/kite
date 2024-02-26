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
SELECT 
    SUM(total_event_count) AS total_event_count,
    SUM(success_event_count) AS success_event_count,
    SUM(total_event_execution_time) AS total_event_execution_time,
    AVG(avg_event_execution_time) AS avg_event_execution_time,
    SUM(total_event_total_time) AS total_event_total_time,
    AVG(avg_event_total_time) AS avg_event_total_time,
    SUM(total_call_count) AS total_call_count,
    SUM(success_call_count) AS success_call_count,
    SUM(total_call_total_time) AS total_call_total_time,
    AVG(avg_call_total_time) AS avg_call_total_time
FROM guild_usage 
WHERE 
    guild_id = $1 AND 
    period_starts_at >= @start_at AND 
    period_ends_at <= @end_at;

-- name: GetGuildUsageAndLimits :one
SELECT * FROM guild_usage_month_view 
INNER JOIN guild_entitlements_resolved_view 
ON guild_usage_month_view.guild_id = guild_entitlements_resolved_view.guild_id 
WHERE guild_usage_month_view.guild_id = $1;