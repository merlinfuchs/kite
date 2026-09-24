-- The usage dashboard queries filter usage_records on an app and a created_at
-- range, but only app_id was indexed, so they fetched every row the app has in
-- retention and filtered the dates out afterwards. For the busiest app that was
-- ~6M rows to keep ~1M, taking seconds per query.
--
-- The INCLUDE columns make the aggregates index-only scans. This replaces
-- usage_records_app_id, which migration 033 drops; the leading app_id still
-- serves the ON DELETE CASCADE from apps.
--
-- CONCURRENTLY can't run in a transaction, so this file must stay a single
-- statement.
CREATE INDEX CONCURRENTLY IF NOT EXISTS usage_records_app_id_created_at
    ON usage_records (app_id, created_at) INCLUDE (type, credits_used);
