CREATE TABLE IF NOT EXISTS apps (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,

    owner_user_id TEXT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    creator_user_id TEXT NOT NULL,

    discord_token TEXT NOT NULL,
    discord_id TEXT NOT NULL,
   
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);
