-- Flows add a hidden instance every time they send a template, so listing the
-- newest ones must not read and sort all of a template's instances.
--
-- CONCURRENTLY keeps template sends from blocking while the index builds. It
-- can't run in a transaction, so this file must stay a single statement.
CREATE INDEX CONCURRENTLY IF NOT EXISTS message_instances_flow_message_id_created_at
    ON message_instances (message_id, created_at DESC)
    WHERE hidden AND NOT ephemeral;
