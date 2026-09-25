-- The engine's command poll now also reads disabled commands, which the
-- partial commands_enabled_updated_at index can't serve. Dropped in 050.
-- CONCURRENTLY can't run in a transaction, so this file must stay a single statement.
CREATE INDEX CONCURRENTLY IF NOT EXISTS commands_updated_at ON commands (updated_at);
