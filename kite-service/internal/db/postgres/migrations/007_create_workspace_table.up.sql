CREATE TABLE IF NOT EXISTS workspaces (
   id TEXT PRIMARY KEY,
   app_id TEXT NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
   type TEXT NOT NULL, -- JS / FLOW
   name TEXT NOT NULL,
   description TEXT NOT NULL,
   files JSONB NOT NULL,
   
   created_at TIMESTAMP NOT NULL,
   updated_at TIMESTAMP NOT NULL
);
