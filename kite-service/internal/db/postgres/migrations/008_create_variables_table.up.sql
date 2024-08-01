CREATE TABLE IF NOT EXISTS variables (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    value JSONB NOT NULL,
    scope TEXT NOT NULL, -- global, guild, user, channel, ... other variable?

    command_id TEXT REFERENCES commands(id) ON DELETE CASCADE,
    -- event_id TEXT REFERENCES events(id) ON DELETE CASCADE,
    -- message_id TEXT REFERENCES messages(id) ON DELETE CASCADE,
   
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);
