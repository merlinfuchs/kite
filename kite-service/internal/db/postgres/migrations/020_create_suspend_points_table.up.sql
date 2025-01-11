CREATE TABLE IF NOT EXISTS suspend_points (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL,

    app_id TEXT NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    command_id TEXT REFERENCES commands(id) ON DELETE SET NULL,
    event_listener_id TEXT REFERENCES event_listeners(id) ON DELETE SET NULL,
    message_id TEXT REFERENCES messages(id) ON DELETE SET NULL,

    flow_source_id TEXT, -- Message templates have multiple flows
    flow_node_id TEXT NOT NULL,
    flow_state JSONB NOT NULL,
   
    created_at TIMESTAMP NOT NULL,
    expires_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS suspend_points_app_id ON suspend_points (app_id);
CREATE INDEX IF NOT EXISTS suspend_points_command_id ON suspend_points (command_id);
CREATE INDEX IF NOT EXISTS suspend_points_event_listener_id ON suspend_points (event_listener_id);
CREATE INDEX IF NOT EXISTS suspend_points_message_id ON suspend_points (message_id);
