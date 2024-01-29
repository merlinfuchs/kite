CREATE TABLE IF NOT EXISTS sessions_pending (
   code TEXT PRIMARY KEY,
   token TEXT,
   
   created_at TIMESTAMP NOT NULL,
   expires_at TIMESTAMP NOT NULL
);
