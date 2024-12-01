CREATE TABLE IF NOT EXISTS message_instances (
    id BIGSERIAL PRIMARY KEY,
    message_id TEXT NOT NULL REFERENCES messages(id) ON DELETE CASCADE,

    hidden BOOLEAN NOT NULL DEFAULT FALSE,
    ephemeral BOOLEAN NOT NULL DEFAULT FALSE,

    discord_guild_id TEXT NOT NULL,
    discord_channel_id TEXT NOT NULL,
    discord_message_id TEXT NOT NULL UNIQUE,
    
    flow_sources JSONB NOT NULL, -- snapshot from the message when sent

    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS message_instances_message_id ON message_instances (message_id);
