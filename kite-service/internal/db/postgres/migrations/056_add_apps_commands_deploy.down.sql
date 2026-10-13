DROP INDEX IF EXISTS apps_commands_deploy_pending;
ALTER TABLE apps DROP COLUMN commands_deploy_error_command;
ALTER TABLE apps DROP COLUMN commands_deploy_error;
ALTER TABLE apps DROP COLUMN commands_deployed_hash;
ALTER TABLE apps DROP COLUMN commands_deploy_retry_at;
ALTER TABLE apps DROP COLUMN commands_changed_at;
ALTER TABLE apps DROP COLUMN commands_version;
ALTER TABLE apps DROP COLUMN commands_deploy_pending;
