-- name: GetAllApps :many
SELECT * FROM apps;

-- name: GetApp :one
SELECT * FROM apps WHERE id = $1;

-- name: GetAppCredentials :one
SELECT discord_id, discord_token FROM apps WHERE id = $1;

-- name: GetAppsByOwner :many
SELECT * FROM apps WHERE owner_user_id = $1 ORDER BY created_at DESC;

-- name: GetAppsByCollaborator :many
SELECT a.* FROM apps a
LEFT JOIN collaborators c ON a.id = c.app_id
WHERE a.owner_user_id = @user_id OR c.user_id = @user_id
ORDER BY a.created_at DESC;

-- name: CountAppsByOwner :one
SELECT COUNT(*) FROM apps WHERE owner_user_id = $1;

-- name: CreateApp :one
INSERT INTO apps (
    id,
    name,
    description,
    owner_user_id,
    creator_user_id,
    discord_token,
    discord_id,
    discord_verify_key,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
) RETURNING *;

-- name: UpdateApp :one
UPDATE apps SET
    name = $2,
    description = $3,
    discord_token = $4,
    discord_status = $5,
    enabled = $6,
    disabled_reason = $7,
    updated_at = $8
WHERE id = $1 RETURNING *;

-- Idempotent: disabling an already-disabled app is a no-op rather than a write
-- that bumps updated_at, which several polls use as their cursor.
-- name: DisableApp :exec
UPDATE apps SET
    enabled = FALSE,
    disabled_reason = $2,
    updated_at = $3
WHERE id = $1 AND enabled;

-- name: DeleteApp :exec
DELETE FROM apps WHERE id = $1;

-- name: GetEnabledAppIDs :many
SELECT id FROM apps WHERE enabled = TRUE;

-- name: GetEnabledAppsUpdatedSince :many
SELECT * FROM apps WHERE enabled = TRUE AND updated_at > $1;

-- name: GetDisabledAppIDsUpdatedSince :many
SELECT id FROM apps WHERE enabled = FALSE AND updated_at > $1;

-- Apps whose gateway intent requirements may have changed.
--
-- Deletes are deliberately not covered: a deleted listener or plugin instance
-- leaves no row for updated_at to match. That only ever narrows the required
-- intents, so missing it means running with more intents than necessary until
-- the next reconnect -- wasteful, but never dropping events. Additions and
-- updates, which widen the requirements, are caught reliably.
-- name: GetAppIDsWithGatewayRequirementsChangedSince :many
SELECT el.app_id FROM event_listeners el WHERE el.updated_at > $1 AND el.source = 'discord'
UNION
SELECT pi.app_id FROM plugin_instances pi WHERE pi.updated_at > $1;

-- Everything the gateway needs to decide which intents to identify with.
-- Plugin resources come back as "plugin_id:resource_id" so this stays a single
-- round trip; the caller resolves them to event types via the plugin registry.
--
-- has_message_instances keeps GUILD_MESSAGES on for apps that rely on
-- MESSAGE_DELETE to clean up message_instances rows. It can be dropped once
-- that cleanup no longer depends on the gateway.
-- name: GetAppGatewayRequirements :one
SELECT
    coalesce(
        (SELECT array_agg(DISTINCT el.type)
         FROM event_listeners el
         WHERE el.app_id = $1 AND el.enabled = TRUE AND el.source = 'discord'),
        '{}'
    )::text[] AS event_listener_types,
    coalesce(
        (SELECT array_agg(DISTINCT pi.plugin_id || ':' || rid)
         FROM plugin_instances pi, unnest(pi.enabled_resource_ids) AS rid
         WHERE pi.app_id = $1 AND pi.enabled = TRUE),
        '{}'
    )::text[] AS plugin_resources,
    EXISTS (
        SELECT 1 FROM messages m
        JOIN message_instances mi ON mi.message_id = m.id
        WHERE m.app_id = $1
    ) AS has_message_instances;

-- name: GetAppEntities :many
SELECT 
    id,
    'command' AS type,
    name
FROM commands
WHERE commands.app_id = $1
UNION ALL
SELECT 
    id,
    'event_listener' AS type,
    type as name
FROM event_listeners
WHERE event_listeners.app_id = $1
UNION ALL
SELECT 
    id,
    'message' AS type,
    name
FROM messages
WHERE messages.app_id = $1
UNION ALL
SELECT 
    id,
    'variable' AS type,
    name
FROM variables
WHERE variables.app_id = $1;
-- Changes to commands and plugins queue a deploy by themselves, see migration
-- 056. This is for deploying again without a change, like after a new token.
-- Forcing clears the hash, so the commands are sent even if they look unchanged.
-- name: QueueAppCommandsDeploy :exec
UPDATE apps SET
    commands_deploy_at = (now() AT TIME ZONE 'UTC') + INTERVAL '3 seconds',
    commands_deployed_hash = CASE WHEN sqlc.arg(force)::BOOLEAN THEN NULL ELSE commands_deployed_hash END
WHERE id = sqlc.arg(id);

-- Newest changes first, so a user waiting for one isn't stuck behind a backlog.
-- name: GetAppIDsWithDueCommandsDeploy :many
SELECT id FROM apps
WHERE enabled AND commands_deploy_at <= now() AT TIME ZONE 'UTC'
ORDER BY commands_deploy_at DESC
LIMIT $1;

-- deploy_at is the commands_deploy_at the deploy started from. Every change
-- sets it to a later time, so if it changed, the commands changed meanwhile and
-- stay queued, keeping the hash, which may have been cleared to force the next
-- deploy.
-- name: FinishAppCommandsDeploy :exec
UPDATE apps SET
    commands_deploy_at = CASE WHEN commands_deploy_at = sqlc.arg(deploy_at) THEN NULL ELSE commands_deploy_at END,
    commands_deployed_hash = CASE
        WHEN commands_deploy_at = sqlc.arg(deploy_at) THEN COALESCE(sqlc.narg(deployed_hash), commands_deployed_hash)
        ELSE commands_deployed_hash
    END,
    commands_deploy_error = sqlc.narg(error)
WHERE id = sqlc.arg(id);

-- Tries again in a few minutes, when Discord or the network failed.
-- name: DelayAppCommandsDeploy :exec
UPDATE apps SET
    commands_deploy_at = GREATEST(commands_deploy_at, (now() AT TIME ZONE 'UTC') + INTERVAL '5 minutes'),
    commands_deploy_error = sqlc.arg(error)
WHERE id = sqlc.arg(id);

-- Doesn't touch updated_at, which the gateway watches for app changes.
-- name: UpdateAppDiscordCounts :exec
UPDATE apps SET
    discord_guild_count = $2,
    discord_user_install_count = $3,
    discord_counts_updated_at = now() AT TIME ZONE 'UTC'
WHERE id = $1;

-- Doesn't touch updated_at either.
-- name: UpdateAppDiscordVerifyKey :exec
UPDATE apps SET discord_verify_key = $2 WHERE id = $1;
