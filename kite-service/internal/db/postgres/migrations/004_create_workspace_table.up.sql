CREATE TABLE IF NOT EXISTS workspaces (
   id TEXT PRIMARY KEY,
   name TEXT NOT NULL,
   description TEXT NOT NULL,
   files JSONB NOT NULL,
   
   created_at TIMESTAMP NOT NULL,
   updated_at TIMESTAMP NOT NULL
);
