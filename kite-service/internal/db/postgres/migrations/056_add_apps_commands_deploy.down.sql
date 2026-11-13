ALTER TABLE plugin_instances ADD COLUMN last_deployed_at TIMESTAMP;
ALTER TABLE commands ADD COLUMN last_deployed_at TIMESTAMP;
DROP TRIGGER IF EXISTS plugin_instances_queue_app_commands_deploy ON plugin_instances;
DROP TRIGGER IF EXISTS commands_queue_app_commands_deploy ON commands;
DROP FUNCTION IF EXISTS queue_app_commands_deploy();
DROP INDEX IF EXISTS apps_commands_deploy_at;
ALTER TABLE apps DROP COLUMN commands_deploy_error;
ALTER TABLE apps DROP COLUMN commands_deployed_hash;
ALTER TABLE apps DROP COLUMN commands_deploy_at;
