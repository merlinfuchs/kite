-- Tombstones for deleted commands, event listeners and plugin instances, so
-- pollers can pick up deletes by deleted_at like they pick up changes by
-- updated_at. Written by triggers to also cover rows deleted by the cascade
-- of an app delete.
CREATE TABLE IF NOT EXISTS deleted_entities (
    id TEXT NOT NULL,
    -- The table the row was deleted from.
    entity_type TEXT NOT NULL,
    app_id TEXT NOT NULL,
    -- UTC like the updated_at columns, which are written by the service.
    deleted_at TIMESTAMP NOT NULL DEFAULT (now() AT TIME ZONE 'UTC')
);

CREATE INDEX IF NOT EXISTS deleted_entities_deleted_at ON deleted_entities (deleted_at);

CREATE OR REPLACE FUNCTION record_deleted_entities() RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO deleted_entities (id, entity_type, app_id)
    SELECT id, TG_TABLE_NAME, app_id FROM deleted_rows;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Statement level, so an app delete cascading to thousands of rows runs the
-- trigger once per table instead of once per row.
CREATE TRIGGER commands_record_deleted
AFTER DELETE ON commands
REFERENCING OLD TABLE AS deleted_rows
FOR EACH STATEMENT EXECUTE FUNCTION record_deleted_entities();

CREATE TRIGGER event_listeners_record_deleted
AFTER DELETE ON event_listeners
REFERENCING OLD TABLE AS deleted_rows
FOR EACH STATEMENT EXECUTE FUNCTION record_deleted_entities();

CREATE TRIGGER plugin_instances_record_deleted
AFTER DELETE ON plugin_instances
REFERENCING OLD TABLE AS deleted_rows
FOR EACH STATEMENT EXECUTE FUNCTION record_deleted_entities();
