-- Superseded by commands_updated_at, nothing else filters commands by updated_at.
-- CONCURRENTLY can't run in a transaction, so this file must stay a single statement.
DROP INDEX CONCURRENTLY IF EXISTS commands_enabled_updated_at;
