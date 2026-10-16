-- Commands are deployed to Discord in the background. commands_deploy_at is
-- NULL when Discord is up to date, and otherwise the earliest time the deployer
-- sends them. The hash is of what Discord has, so changes to what a command
-- does, not how it looks, need no request.
ALTER TABLE apps ADD COLUMN commands_deploy_at TIMESTAMP;
ALTER TABLE apps ADD COLUMN commands_deployed_hash TEXT;
-- {"message": "...", "command": "ban"} when the last deploy failed.
ALTER TABLE apps ADD COLUMN commands_deploy_error JSONB;

CREATE INDEX IF NOT EXISTS apps_commands_deploy_at ON apps (commands_deploy_at) WHERE commands_deploy_at IS NOT NULL;

-- Every change to an app's commands or plugins queues a deploy a few seconds
-- later, so quick successive changes go out together. It's a trigger, so no
-- write can miss it. It doesn't touch updated_at, which the gateway watches for
-- app changes.
CREATE OR REPLACE FUNCTION queue_app_commands_deploy() RETURNS TRIGGER AS $$
BEGIN
    UPDATE apps SET commands_deploy_at = (now() AT TIME ZONE 'UTC') + INTERVAL '3 seconds'
    WHERE id = CASE WHEN TG_OP = 'DELETE' THEN OLD.app_id ELSE NEW.app_id END;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER commands_queue_app_commands_deploy
AFTER INSERT OR UPDATE OR DELETE ON commands
FOR EACH ROW EXECUTE FUNCTION queue_app_commands_deploy();

CREATE TRIGGER plugin_instances_queue_app_commands_deploy
AFTER INSERT OR UPDATE OR DELETE ON plugin_instances
FOR EACH ROW EXECUTE FUNCTION queue_app_commands_deploy();

-- Deploy what was changed in the last 30 days but never deployed. Older
-- changes deploy with the next change of the app.
UPDATE apps SET commands_deploy_at = now() AT TIME ZONE 'UTC'
WHERE id IN (
    SELECT app_id FROM commands
    WHERE (last_deployed_at IS NULL OR last_deployed_at < updated_at)
        AND updated_at > now() AT TIME ZONE 'UTC' - INTERVAL '30 days'
    UNION
    SELECT app_id FROM plugin_instances
    WHERE (last_deployed_at IS NULL OR last_deployed_at < updated_at)
        AND updated_at > now() AT TIME ZONE 'UTC' - INTERVAL '30 days'
);

-- Replaced by the columns above, and only read by the backfill.
ALTER TABLE commands DROP COLUMN last_deployed_at;
ALTER TABLE plugin_instances DROP COLUMN last_deployed_at;
