-- Nothing reads these columns. The foreign keys only forced an index on each
-- (dropped in 039-041) so deleting a command, event listener or message didn't
-- scan the whole table for its SET NULL.
ALTER TABLE usage_records DROP CONSTRAINT IF EXISTS usage_records_command_id_fkey;
ALTER TABLE usage_records DROP CONSTRAINT IF EXISTS usage_records_event_listener_id_fkey;
ALTER TABLE usage_records DROP CONSTRAINT IF EXISTS usage_records_message_id_fkey;

-- The default 20% waited ~10M rows between runs. The covering
-- usage_records_app_id_created_at index only avoids heap reads on pages vacuum
-- has marked all-visible, and the month-to-date pages it serves are the newest.
ALTER TABLE usage_records SET (
    autovacuum_vacuum_scale_factor = 0.01,
    autovacuum_vacuum_insert_scale_factor = 0.01
);
