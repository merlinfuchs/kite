CREATE TABLE IF NOT EXISTS assets (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    content_hash TEXT NOT NULL,
    content_type TEXT NOT NULL,
    content_size INTEGER NOT NULL,

    app_id TEXT NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    module_id TEXT REFERENCES modules(id) ON DELETE SET NULL,
    creator_user_id TEXT NOT NULL,
   
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    expires_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS assets_app_id ON assets (app_id);
CREATE INDEX IF NOT EXISTS assets_module_id ON assets (module_id);
CREATE INDEX IF NOT EXISTS assets_content_hash ON assets (content_hash);
