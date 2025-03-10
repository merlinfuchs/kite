CREATE TABLE IF NOT EXISTS variables (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    scoped BOOLEAN NOT NULL DEFAULT FALSE,

    app_id TEXT NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    module_id TEXT REFERENCES modules(id) ON DELETE SET NULL,
   
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,

    UNIQUE (app_id, name)
);

CREATE INDEX IF NOT EXISTS variables_app_id ON variables (app_id);
CREATE INDEX IF NOT EXISTS variables_module_id ON variables (module_id);
