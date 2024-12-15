CREATE TABLE IF NOT EXISTS usage_records (
    id BIGSERIAL PRIMARY KEY,
    type TEXT NOT NULL,

    app_id TEXT NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    command_id TEXT REFERENCES commands(id) ON DELETE SET NULL,
    event_listener_id TEXT REFERENCES event_listeners(id) ON DELETE SET NULL,
    message_id TEXT REFERENCES messages(id) ON DELETE SET NULL,
    
    credits_used INTEGER NOT NULL,
   
    created_at TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS usage_records_app_id ON usage_records (app_id);
CREATE INDEX IF NOT EXISTS usage_records_command_id ON usage_records (command_id);
CREATE INDEX IF NOT EXISTS usage_records_event_listener_id ON usage_records (event_listener_id);
CREATE INDEX IF NOT EXISTS usage_records_message_id ON usage_records (message_id);
