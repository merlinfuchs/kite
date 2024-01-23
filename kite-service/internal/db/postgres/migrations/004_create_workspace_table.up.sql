CREATE TABLE IF NOT EXISTS workspaces (
   id TEXT PRIMARY KEY,
   guild_id TEXT NOT NULL REFERENCES guilds(id) ON DELETE CASCADE,
   name TEXT NOT NULL,
   description TEXT NOT NULL,
   files JSONB NOT NULL,
   
   created_at TIMESTAMP NOT NULL,
   updated_at TIMESTAMP NOT NULL
);
