-- Tokens users create to use the API without logging in, e.g. from an MCP
-- client. Only a hash of the token is stored, it's shown once when created.
CREATE TABLE IF NOT EXISTS api_tokens (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    key_hash TEXT NOT NULL UNIQUE,
    -- What the token may do. Every token has the "all" scope for now.
    scopes TEXT[] NOT NULL,
    last_used_at TIMESTAMP,
    -- Null for tokens that don't expire.
    expires_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS api_tokens_user_id ON api_tokens (user_id);
