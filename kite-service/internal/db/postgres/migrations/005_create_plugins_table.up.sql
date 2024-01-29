CREATE TABLE IF NOT EXISTS plugins (
   id TEXT PRIMARY KEY,
   key TEXT NOT NULL UNIQUE,
   name TEXT NOT NULL,
   description TEXT NOT NULL,
   
   created_at TIMESTAMP NOT NULL,
   updated_at TIMESTAMP NOT NULL
);
