-- Serves the engine's two ID queries as index-only scans: the per-poll check
-- for deleted scheduled listeners (source = 'schedule') and the dangling sweep
-- (all enabled IDs). Both scanned the whole table, which is mostly flow_source.
-- CONCURRENTLY can't run in a transaction, so this file must stay a single statement.
CREATE INDEX CONCURRENTLY IF NOT EXISTS event_listeners_enabled_source_id ON event_listeners (source, id) WHERE enabled = TRUE;
