-- What users see when they run into a limit. A limit without a message uses
-- the default message of its app from credit_limit_settings, and without one
-- of those Kite's own message.
ALTER TABLE credit_limits ADD COLUMN IF NOT EXISTS message TEXT;

CREATE TABLE IF NOT EXISTS credit_limit_settings (
    app_id TEXT PRIMARY KEY REFERENCES apps(id) ON DELETE CASCADE,
    message TEXT,

    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);
