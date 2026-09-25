CREATE TABLE IF NOT EXISTS share_codes (
    code TEXT PRIMARY KEY,
    data TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL
);
