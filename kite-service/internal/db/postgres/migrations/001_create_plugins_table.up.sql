CREATE TABLE IF NOT EXISTS plugins (
   id TEXT PRIMARY KEY,
   name TEXT NOT NULL,
   description TEXT NOT NULL,
   
   created_at TIMESTAMP NOT NULL,
   updated_at TIMESTAMP NOT NULL
);
