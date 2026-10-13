-- Commands are deployed to Discord in the background. An app is pending when
-- its commands or plugins changed after the last deploy attempt, and isn't
-- attempted again before commands_deploy_retry_at. Every change increments
-- commands_version, so a deploy can tell if the commands changed meanwhile.
ALTER TABLE apps ADD COLUMN commands_deploy_pending BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE apps ADD COLUMN commands_version BIGINT NOT NULL DEFAULT 0;
ALTER TABLE apps ADD COLUMN commands_changed_at TIMESTAMP;
ALTER TABLE apps ADD COLUMN commands_deploy_retry_at TIMESTAMP;
-- Hash of what was last sent to Discord, so changes that don't affect how
-- commands look in Discord don't need a request.
ALTER TABLE apps ADD COLUMN commands_deployed_hash TEXT;
-- Why Discord rejected the last deploy, and the command it rejected if known.
ALTER TABLE apps ADD COLUMN commands_deploy_error TEXT;
ALTER TABLE apps ADD COLUMN commands_deploy_error_command TEXT;

CREATE INDEX IF NOT EXISTS apps_commands_deploy_pending ON apps (commands_changed_at) WHERE commands_deploy_pending;

-- Deploy what was changed in the last 30 days but never deployed. Older
-- changes deploy with the next change of the app.
UPDATE apps SET
    commands_deploy_pending = TRUE,
    commands_changed_at = now() AT TIME ZONE 'UTC'
WHERE id IN (
    SELECT app_id FROM commands
    WHERE (last_deployed_at IS NULL OR last_deployed_at < updated_at)
        AND updated_at > now() AT TIME ZONE 'UTC' - INTERVAL '30 days'
    UNION
    SELECT app_id FROM plugin_instances
    WHERE (last_deployed_at IS NULL OR last_deployed_at < updated_at)
        AND updated_at > now() AT TIME ZONE 'UTC' - INTERVAL '30 days'
);
