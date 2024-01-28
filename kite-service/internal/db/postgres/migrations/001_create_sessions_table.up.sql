CREATE TABLE IF NOT EXISTS sessions (
   token_hash TEXT PRIMARY KEY,
   type TEXT NOT NULL,
   user_id TEXT NOT NULL,
   guild_ids TEXT[] NOT NULL,
   access_token TEXT NOT NULL,
   revoked BOOLEAN NOT NULL DEFAULT false,
   
   created_at TIMESTAMP NOT NULL,
   expires_at TIMESTAMP NOT NULL
);
