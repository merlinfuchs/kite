-- Codes of the OAuth authorization code flow, made when a user approves a
-- client like Claude. Used codes stay until they're cleaned up, so a code
-- used twice can revoke the token it was exchanged for.
CREATE TABLE IF NOT EXISTS oauth_codes (
    code_hash TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    -- The URL of the client's metadata document.
    client_id TEXT NOT NULL,
    client_name TEXT NOT NULL,
    redirect_uri TEXT NOT NULL,
    code_challenge TEXT NOT NULL,
    scopes TEXT[] NOT NULL,
    app_ids TEXT[],
    resource TEXT NOT NULL,
    -- The token the code was exchanged for, null until it's used.
    api_token_id TEXT,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS oauth_codes_expires_at ON oauth_codes (expires_at);

-- OAuth clients get API tokens too, with a refresh token. For them expires_at
-- is when the refresh token expires, which every refresh pushes back, and
-- access_expires_at is when the token itself does.
ALTER TABLE api_tokens ADD COLUMN IF NOT EXISTS oauth_client_id TEXT;
ALTER TABLE api_tokens ADD COLUMN IF NOT EXISTS oauth_resource TEXT;
ALTER TABLE api_tokens ADD COLUMN IF NOT EXISTS refresh_key_hash TEXT UNIQUE;
ALTER TABLE api_tokens ADD COLUMN IF NOT EXISTS access_expires_at TIMESTAMP;
