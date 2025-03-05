CREATE TABLE IF NOT EXISTS plugin_instances (
    plugin_id TEXT NOT NULL,
    app_id TEXT NOT NULL REFERENCES apps(id) ON DELETE CASCADE,

    enabled BOOLEAN NOT NULL DEFAULT FALSE,
    config JSONB NOT NULL,

    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    commands_deployed_at TIMESTAMP,

    PRIMARY KEY (app_id, plugin_id)
);

INSERT INTO plugin_instances 
SELECT 
    'builder' as plugin_id, 
    id as app_id, 
    TRUE as enabled, 
    '{}' as config, 
    NOW() as created_at, 
    NOW() as updated_at, 
    NULL as commands_deployed_at 
FROM apps 
WHERE id NOT IN (
    SELECT DISTINCT app_id FROM plugin_instances WHERE plugin_id = 'builder'
);

CREATE INDEX IF NOT EXISTS plugin_instances_app_id ON plugin_instances (app_id);
