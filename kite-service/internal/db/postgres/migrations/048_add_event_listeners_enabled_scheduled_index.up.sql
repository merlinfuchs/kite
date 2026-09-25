-- The engine checks for deleted scheduled listeners on every poll, which
-- scanned the whole table since nothing indexes source.
-- CONCURRENTLY can't run in a transaction, so this file must stay a single statement.
CREATE INDEX CONCURRENTLY IF NOT EXISTS event_listeners_enabled_scheduled ON event_listeners (id) WHERE enabled = TRUE AND source = 'schedule';
