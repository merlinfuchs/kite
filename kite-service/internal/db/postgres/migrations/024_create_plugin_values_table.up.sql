CREATE TABLE IF NOT EXISTS plugin_values (
    plugin_id TEXT NOT NULL,
    key TEXT NOT NULL,
    app_id TEXT NOT NULL REFERENCES apps(id) ON DELETE CASCADE,

    value JSONB NOT NULL,

    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,

    PRIMARY KEY (app_id, plugin_id, key)
);

CREATE INDEX IF NOT EXISTS plugin_values_app_id ON plugin_values (app_id);
CREATE INDEX IF NOT EXISTS plugin_values_plugin_id ON plugin_values (plugin_id);
