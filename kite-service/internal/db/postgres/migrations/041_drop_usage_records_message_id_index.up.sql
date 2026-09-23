-- CONCURRENTLY can't run in a transaction, so this file must stay a single statement.
DROP INDEX CONCURRENTLY IF EXISTS usage_records_message_id;
