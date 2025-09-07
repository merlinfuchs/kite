CREATE TABLE IF NOT EXISTS plugin_values (
    id BIGSERIAL PRIMARY KEY,
    plugin_instance_id TEXT NOT NULL REFERENCES plugin_instances(id) ON DELETE CASCADE,

    key TEXT NOT NULL,
    value JSONB NOT NULL,
   
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,

    UNIQUE (plugin_instance_id, key)
);

CREATE INDEX IF NOT EXISTS plugin_values_plugin_instance_id ON plugin_values (plugin_instance_id);
CREATE INDEX IF NOT EXISTS plugin_values_plugin_instance_key ON plugin_values (key);
