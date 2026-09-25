-- Lets the engine's dangling sweep ("SELECT id WHERE enabled") read the IDs
-- from the index instead of scanning the table. Migration 027 measured the
-- sequential scan as faster, but with realistically sized flows inline in the
-- table, the index-only scan is ~4ms against ~40ms at 100k rows.
-- CONCURRENTLY can't run in a transaction, so this file must stay a single statement.
CREATE INDEX CONCURRENTLY IF NOT EXISTS commands_enabled_id ON commands (id) WHERE enabled = TRUE;
