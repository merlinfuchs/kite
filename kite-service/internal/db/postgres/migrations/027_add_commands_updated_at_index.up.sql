-- The engine polls for changed commands every few seconds:
--   SELECT * FROM commands WHERE enabled = TRUE AND updated_at > $1
-- and commands was the one of the three polled tables with no index on
-- updated_at, so every poll was a sequential scan -- over the widest of them,
-- since commands carries a flow_source JSONB column.
--
-- Partial on enabled because both queries that use this column filter on it,
-- and nothing reads commands.updated_at as a range without it. Migration 026
-- already covers event_listeners and plugin_instances with an unconditional
-- index, which their gateway requirements query needs.
--
-- Deliberately no INCLUDE (id) for the dangling sweep's "SELECT id WHERE
-- enabled": that returns almost every row, and a sequential scan measured
-- faster than the index-only scan it would enable (7.4ms vs 12.3ms at 100k
-- rows), so the planner correctly ignores it.
CREATE INDEX IF NOT EXISTS commands_enabled_updated_at
    ON commands (updated_at) WHERE enabled;
