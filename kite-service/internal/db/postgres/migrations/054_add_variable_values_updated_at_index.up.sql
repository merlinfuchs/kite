-- Serves the dashboard's list of a variable's values, newest first.
-- CONCURRENTLY can't run in a transaction, so this file must stay a single statement.
CREATE INDEX CONCURRENTLY IF NOT EXISTS variable_values_variable_id_updated_at ON variable_values (variable_id, updated_at DESC);
