CREATE TABLE IF NOT EXISTS event_listeners (
    id TEXT PRIMARY KEY,
    source TEXT NOT NULL,
    type TEXT NOT NULL,
    description TEXT NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT FALSE,

    app_id TEXT NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    module_id TEXT REFERENCES modules(id) ON DELETE SET NULL,
    creator_user_id TEXT NOT NULL,

    filter JSONB,
    flow_source JSONB NOT NULL,
   
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS event_listeners_app_id ON event_listeners (app_id);
CREATE INDEX IF NOT EXISTS event_listeners_module_id ON event_listeners (module_id);
