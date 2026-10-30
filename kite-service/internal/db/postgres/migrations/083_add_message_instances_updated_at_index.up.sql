-- The gateway manager polls message_instances by updated_at every few seconds
-- on every cluster, to notice messages whose flows play audio:
--   SELECT ... FROM message_instances mi ... WHERE mi.updated_at > $1
-- Without an index each poll is a sequential scan.
--
-- CONCURRENTLY can't run in a transaction, so this file must stay a single
-- statement.
CREATE INDEX CONCURRENTLY IF NOT EXISTS message_instances_updated_at ON message_instances (updated_at);
