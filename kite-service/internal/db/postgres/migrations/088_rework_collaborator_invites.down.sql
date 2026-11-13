DROP TABLE IF EXISTS app_audit_logs;

DROP INDEX IF EXISTS collaborator_invites_invited_by_user_id;
DROP INDEX IF EXISTS collaborator_invites_user_id;
ALTER TABLE collaborator_invites DROP COLUMN IF EXISTS invited_by_user_id;
ALTER TABLE collaborator_invites DROP COLUMN IF EXISTS user_id;
