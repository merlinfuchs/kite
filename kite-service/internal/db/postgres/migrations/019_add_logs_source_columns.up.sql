ALTER TABLE logs ADD COLUMN IF NOT EXISTS command_id TEXT REFERENCES commands(id) ON DELETE SET NULL;
ALTER TABLE logs ADD COLUMN IF NOT EXISTS event_listener_id TEXT REFERENCES event_listeners(id) ON DELETE SET NULL;
ALTER TABLE logs ADD COLUMN IF NOT EXISTS message_id TEXT REFERENCES messages(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS logs_command_id ON logs (command_id);
CREATE INDEX IF NOT EXISTS logs_event_listener_id ON logs (event_listener_id);
CREATE INDEX IF NOT EXISTS logs_message_id ON logs (message_id);
