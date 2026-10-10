-- Credit limits sum an app's usage in one server since the start of the day or
-- month. Partial, so rows without a server cost nothing to index.
--
-- CONCURRENTLY can't run in a transaction, so this file must stay a single
-- statement.
CREATE INDEX CONCURRENTLY IF NOT EXISTS usage_records_app_id_guild_id_created_at
    ON usage_records (app_id, guild_id, created_at) INCLUDE (credits_used)
    WHERE guild_id IS NOT NULL;
