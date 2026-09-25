CREATE TABLE IF NOT EXISTS share_codes (
    code TEXT PRIMARY KEY,
    type TEXT NOT NULL,
    data JSONB NOT NULL,
    -- Hashes the normalized JSONB so formatting and key order can't split
    -- duplicates. Only used for dedupe, so md5 is enough.
    data_hash TEXT GENERATED ALWAYS AS (md5(data::text)) STORED,
    creator_user_id TEXT NOT NULL,
    -- No foreign key so codes keep working after the app is deleted.
    app_id TEXT,

    created_at TIMESTAMP NOT NULL,
    last_used_at TIMESTAMP NOT NULL
);

-- Exporting the same data from the same app again returns the existing code.
CREATE UNIQUE INDEX IF NOT EXISTS share_codes_app_id_type_data_hash
    ON share_codes (app_id, type, data_hash) NULLS NOT DISTINCT;

CREATE INDEX IF NOT EXISTS share_codes_last_used_at ON share_codes (last_used_at);
