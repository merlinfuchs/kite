CREATE TABLE editor_sessions (
    id TEXT PRIMARY KEY,
    app_id TEXT NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    target_type TEXT NOT NULL CHECK (target_type IN ('command', 'event_listener')),
    target_id TEXT NOT NULL,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    active_seconds INTEGER NOT NULL DEFAULT 0 CHECK (active_seconds >= 0),
    save_count INTEGER NOT NULL DEFAULT 0 CHECK (save_count >= 0),
    created_at TIMESTAMP NOT NULL,
    last_edited_at TIMESTAMP,
    updated_at TIMESTAMP NOT NULL
);

CREATE INDEX editor_sessions_app_created_at
    ON editor_sessions (app_id, created_at);
