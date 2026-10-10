-- name: CreateUsageRecord :exec
INSERT INTO usage_records (
    type,
    app_id,
    command_id,
    event_listener_id,
    message_id,
    credits_used,
    created_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: GetUsageRecordsByAppBetween :many
SELECT * FROM usage_records WHERE app_id = @app_id AND created_at BETWEEN @start_at AND @end_at ORDER BY created_at DESC;

-- name: GetUsageCreditsUsedByAppBetween :one
SELECT COALESCE(SUM(credits_used), 0)::int FROM usage_records WHERE app_id = @app_id AND created_at BETWEEN @start_at AND @end_at;

-- name: GetUsageCreditsUsedByTypeBetween :many
SELECT type, SUM(credits_used) FROM usage_records WHERE app_id = @app_id AND created_at BETWEEN @start_at AND @end_at GROUP BY type;

-- name: GetUsageCreditsUsedByDayBetween :many
SELECT 
    d.dt as date, 
    coalesce(u.credits_used, 0) as credits_used 
FROM (
    SELECT dt::date 
    FROM generate_series(@start_at::timestamp, @end_at::timestamp, '1 day'::interval) dt
) d
LEFT JOIN (
    SELECT DATE(created_at) as date, SUM(credits_used) as credits_used 
    FROM usage_records 
    WHERE app_id = @app_id AND created_at BETWEEN @start_at AND @end_at 
    GROUP BY DATE(created_at)
) u ON d.dt = u.date
ORDER BY d.dt;

-- Only enabled apps. A disabled app's usage rows stay for the rest of the
-- month, so without this filter the credit sweep re-disables it on every run:
-- an UPDATE per app per minute, each bumping apps.updated_at, which is the
-- cursor GetDisabledAppIDsUpdatedSince feeds to the gateway manager's poll.
-- name: GetAllUsageCreditsUsedBetween :many
SELECT u.app_id, SUM(u.credits_used) FROM usage_records u
JOIN apps a ON a.id = u.app_id AND a.enabled
WHERE u.created_at BETWEEN @start_at AND @end_at
GROUP BY u.app_id;

-- name: DeleteUsageRecordsBefore :execrows
-- Batched so a large backlog doesn't hold one long transaction.
DELETE FROM usage_records WHERE id IN (
    SELECT expired.id FROM usage_records expired
    WHERE expired.created_at < @before_at
    LIMIT @batch_size
);

-- name: RollupUsageRecordsBefore :execrows
-- Sums every whole day before @before_at that isn't rolled up yet into
-- usage_daily_rollups. Days are rolled up in order, so everything before the
-- latest rolled up day is already done. @before_at must be midnight UTC.
INSERT INTO usage_daily_rollups (date, app_id, type, source_id, executions, credits_used)
SELECT
    DATE(u.created_at),
    u.app_id,
    u.type,
    COALESCE(u.command_id, u.event_listener_id, u.message_id, ''),
    COUNT(*),
    SUM(u.credits_used)
FROM usage_records u
WHERE u.created_at < @before_at::timestamp
    AND u.created_at >= (SELECT COALESCE(MAX(r.date) + 1, '-infinity'::date)::timestamp FROM usage_daily_rollups r)
GROUP BY 1, 2, 3, 4
ON CONFLICT DO NOTHING;

-- name: GetUsageAnalyticsTotalsBetween :many
-- The analytics queries read rolled up days from usage_daily_rollups and the
-- rest from usage_records. Records of rolled up days that the cleanup hasn't
-- deleted yet are skipped so nothing is counted twice.
WITH bounds AS (
    SELECT COALESCE(MAX(date) + 1, '-infinity'::date)::timestamp AS rolled_until FROM usage_daily_rollups
)
SELECT
    s.type::text AS type,
    SUM(s.executions)::bigint AS executions,
    SUM(s.credits_used)::bigint AS credits_used
FROM (
    SELECT r.type, r.executions, r.credits_used
    FROM usage_daily_rollups r
    WHERE r.app_id = @app_id AND r.date::timestamp >= @start_at::timestamp AND r.date::timestamp < @end_at::timestamp
    UNION ALL
    SELECT u.type, 1::bigint, u.credits_used::bigint
    FROM usage_records u, bounds b
    WHERE u.app_id = @app_id AND u.created_at >= GREATEST(@start_at::timestamp, b.rolled_until) AND u.created_at < @end_at::timestamp
) s
GROUP BY s.type;

-- name: GetUsageAnalyticsSeriesBetween :many
-- @bucket is a date_trunc unit: hour, day or month.
WITH bounds AS (
    SELECT COALESCE(MAX(date) + 1, '-infinity'::date)::timestamp AS rolled_until FROM usage_daily_rollups
)
SELECT
    s.bucket::timestamp AS bucket,
    s.type::text AS type,
    SUM(s.executions)::bigint AS executions,
    SUM(s.credits_used)::bigint AS credits_used
FROM (
    SELECT date_trunc(@bucket::text, r.date::timestamp) AS bucket, r.type, r.executions, r.credits_used
    FROM usage_daily_rollups r
    WHERE r.app_id = @app_id AND r.date::timestamp >= @start_at::timestamp AND r.date::timestamp < @end_at::timestamp
    UNION ALL
    SELECT date_trunc(@bucket::text, u.created_at), u.type, 1::bigint, u.credits_used::bigint
    FROM usage_records u, bounds b
    WHERE u.app_id = @app_id AND u.created_at >= GREATEST(@start_at::timestamp, b.rolled_until) AND u.created_at < @end_at::timestamp
) s
GROUP BY 1, 2
ORDER BY 1;

-- name: GetUsageAnalyticsTopSourcesBetween :many
-- Returns up to @per_type sources for each type, most executions first.
WITH bounds AS (
    SELECT COALESCE(MAX(date) + 1, '-infinity'::date)::timestamp AS rolled_until FROM usage_daily_rollups
), totals AS (
    SELECT s.type, s.source_id, SUM(s.executions)::bigint AS executions, SUM(s.credits_used)::bigint AS credits_used
    FROM (
        SELECT r.type, r.source_id, r.executions, r.credits_used
        FROM usage_daily_rollups r
        WHERE r.app_id = @app_id AND r.date::timestamp >= @start_at::timestamp AND r.date::timestamp < @end_at::timestamp
        UNION ALL
        SELECT u.type, COALESCE(u.command_id, u.event_listener_id, u.message_id, ''), 1::bigint, u.credits_used::bigint
        FROM usage_records u, bounds b
        WHERE u.app_id = @app_id AND u.created_at >= GREATEST(@start_at::timestamp, b.rolled_until) AND u.created_at < @end_at::timestamp
    ) s
    WHERE s.source_id <> ''
    GROUP BY s.type, s.source_id
), ranked AS (
    SELECT t.*, ROW_NUMBER() OVER (PARTITION BY t.type ORDER BY t.executions DESC, t.source_id) AS rank
    FROM totals t
)
SELECT
    ranked.type::text AS type,
    ranked.source_id::text AS source_id,
    ranked.executions::bigint AS executions,
    ranked.credits_used::bigint AS credits_used
FROM ranked
WHERE ranked.rank <= @per_type::int
ORDER BY ranked.type, ranked.executions DESC, ranked.source_id;
