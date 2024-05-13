-- name: GetQuickAccessItems :many
SELECT id, app_id, name, updated_at, 'DEPLOYMENT' as type FROM deployments WHERE deployments.app_id = $1
UNION ALL 
SELECT id, app_id, name, updated_at, 'WORKSPACE' as type FROM workspaces WHERE workspaces.app_id = $1
ORDER BY updated_at DESC
LIMIT $2;

