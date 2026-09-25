-- CONCURRENTLY can't run in a transaction, so this file must stay a single statement.
CREATE INDEX CONCURRENTLY IF NOT EXISTS commands_enabled_updated_at ON commands (updated_at) WHERE enabled;
