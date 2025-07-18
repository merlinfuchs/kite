CREATE TABLE IF NOT EXISTS plugin_instances (
    id TEXT PRIMARY KEY,
    plugin_id TEXT NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT FALSE,

    app_id TEXT NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    creator_user_id TEXT NOT NULL,

    config JSONB NOT NULL,
    enabled_resource_ids TEXT[] NOT NULL,
   
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    last_deployed_at TIMESTAMP,

    UNIQUE (plugin_id, app_id)
);

CREATE INDEX IF NOT EXISTS plugin_instances_app_id ON plugin_instances (app_id);
CREATE INDEX IF NOT EXISTS plugin_instances_plugin_id ON plugin_instances (plugin_id);
