CREATE TABLE IF NOT EXISTS sessions (
   token_hash TEXT PRIMARY KEY,
   user_id TEXT NOT NULL,
   guild_ids TEXT[] NOT NULL,
   access_token TEXT NOT NULL,
   
   created_at TIMESTAMP NOT NULL,
   expires_at TIMESTAMP NOT NULL
);
