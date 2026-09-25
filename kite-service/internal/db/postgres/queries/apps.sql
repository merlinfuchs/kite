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
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
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

-- Apps whose gateway intent requirements may have changed. Deleted scheduled
-- listeners can't be told apart from Discord ones, those only cause a refresh
-- that finds the intents unchanged.
-- name: GetAppIDsWithGatewayRequirementsChangedSince :many
SELECT el.app_id FROM event_listeners el WHERE el.updated_at > $1 AND el.source = 'discord'
UNION
SELECT pi.app_id FROM plugin_instances pi WHERE pi.updated_at > $1
UNION
SELECT de.app_id FROM deleted_entities de
WHERE de.deleted_at > $1 AND de.entity_type IN ('event_listeners', 'plugin_instances');

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