CREATE TABLE IF NOT EXISTS guilds (
   id TEXT PRIMARY KEY,
   name TEXT NOT NULL,
   icon TEXT,
   description TEXT,
   
   created_at TIMESTAMP NOT NULL,
   updated_at TIMESTAMP NOT NULL
);
