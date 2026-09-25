CREATE TABLE IF NOT EXISTS share_codes (
    code TEXT PRIMARY KEY,
    type TEXT NOT NULL,
    data JSONB NOT NULL,
    creator_user_id TEXT NOT NULL,
    -- No foreign key so codes keep working after the app is deleted.
    app_id TEXT,

    created_at TIMESTAMP NOT NULL,
    last_used_at TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS share_codes_last_used_at ON share_codes (last_used_at);
