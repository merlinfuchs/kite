CREATE TABLE IF NOT EXISTS plugin_instances (
    plugin_id TEXT NOT NULL,
    app_id TEXT NOT NULL REFERENCES apps(id) ON DELETE CASCADE,

    enabled BOOLEAN NOT NULL DEFAULT FALSE,
    config JSONB NOT NULL,

    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,

    PRIMARY KEY (app_id, plugin_id)
);

CREATE INDEX IF NOT EXISTS plugin_instances_app_id ON plugin_instances (app_id);
