ALTER TABLE usage_records RESET (
    autovacuum_vacuum_scale_factor,
    autovacuum_vacuum_insert_scale_factor
);

ALTER TABLE usage_records ADD CONSTRAINT usage_records_command_id_fkey
    FOREIGN KEY (command_id) REFERENCES commands(id) ON DELETE SET NULL NOT VALID;
ALTER TABLE usage_records ADD CONSTRAINT usage_records_event_listener_id_fkey
    FOREIGN KEY (event_listener_id) REFERENCES event_listeners(id) ON DELETE SET NULL NOT VALID;
ALTER TABLE usage_records ADD CONSTRAINT usage_records_message_id_fkey
    FOREIGN KEY (message_id) REFERENCES messages(id) ON DELETE SET NULL NOT VALID;
