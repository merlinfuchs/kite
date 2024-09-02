CREATE TABLE IF NOT EXISTS message_instances (
    id BIGSERIAL PRIMARY KEY,
    message_id TEXT NOT NULL REFERENCES messages(id) ON DELETE CASCADE,

    discord_guild_id TEXT NOT NULL,
    discord_channel_id TEXT NOT NULL,
    discord_message_id TEXT NOT NULL UNIQUE,
    
    flow_sources JSONB NOT NULL, -- map of flow source ids to flow source objects

    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);