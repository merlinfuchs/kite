DROP INDEX IF EXISTS commands_app_id_position;
DROP INDEX IF EXISTS event_listeners_app_id_position;
DROP INDEX IF EXISTS messages_app_id_position;

ALTER TABLE commands DROP COLUMN position;
ALTER TABLE event_listeners DROP COLUMN position;
ALTER TABLE messages DROP COLUMN position;
