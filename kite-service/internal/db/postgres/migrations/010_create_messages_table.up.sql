CREATE TABLE IF NOT EXISTS messages (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,

    data JSONB NOT NULL, -- message data
    flow_sources JSONB NOT NULL, -- map of flow source ids to flow source objects

    app_id TEXT NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    module_id TEXT REFERENCES modules(id) ON DELETE SET NULL,
    creator_user_id TEXT NOT NULL,

    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);
