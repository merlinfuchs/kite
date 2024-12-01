CREATE TABLE IF NOT EXISTS api_keys (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL,
    name TEXT NOT NULL,

    key TEXT NOT NULL,
    key_hash TEXT NOT NULL,

    app_id TEXT NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    creator_user_id TEXT NOT NULL,
   
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    expires_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS api_keys_app_id ON api_keys (app_id);
CREATE INDEX IF NOT EXISTS api_keys_key_hash ON api_keys (key_hash);
