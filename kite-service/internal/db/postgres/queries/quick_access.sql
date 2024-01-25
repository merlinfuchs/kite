-- name: GetQuickAccessItems :many
SELECT id, guild_id, name, updated_at, 'DEPLOYMENT' as type FROM deployments WHERE deployments.guild_id = $1
UNION ALL 
SELECT id, guild_id, name, updated_at, 'WORKSPACE' as type FROM workspaces WHERE workspaces.guild_id = $1
ORDER BY updated_at DESC
LIMIT $2;

