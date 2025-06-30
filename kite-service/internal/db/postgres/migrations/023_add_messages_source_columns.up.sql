ALTER TABLE messages ADD COLUMN IF NOT EXISTS command_id TEXT REFERENCES commands(id) ON DELETE CASCADE;
ALTER TABLE messages ADD COLUMN IF NOT EXISTS event_listener_id TEXT REFERENCES event_listeners(id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS messages_command_id ON messages (command_id);
CREATE INDEX IF NOT EXISTS messages_event_listener_id ON messages (event_listener_id);
