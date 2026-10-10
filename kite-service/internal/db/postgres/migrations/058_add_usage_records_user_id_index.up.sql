-- Like usage_records_app_id_guild_id_created_at, for per-user credit limits.
--
-- CONCURRENTLY can't run in a transaction, so this file must stay a single
-- statement.
CREATE INDEX CONCURRENTLY IF NOT EXISTS usage_records_app_id_user_id_created_at
    ON usage_records (app_id, user_id, created_at) INCLUDE (credits_used)
    WHERE user_id IS NOT NULL;
