-- name: GetWorkspaceForGuild :one
SELECT * FROM workspaces WHERE id = $1 AND guild_id = $2;

-- name: GetWorkspacesForGuild :many
SELECT * FROM workspaces WHERE guild_id = $1 ORDER BY updated_at DESC;

-- name: CreateWorkspace :one
INSERT INTO workspaces (
    id,
    guild_id,
    name,
    description,
    files,
    created_at,
    updated_at
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7
) RETURNING *;

-- name: UpdateWorkspace :one
UPDATE workspaces SET
    name = $3,
    description = $4,
    files = $5,
    updated_at = $6
WHERE 
    id = $1 AND 
    guild_id = $2 
RETURNING *;

-- name: DeleteWorkspace :one
DELETE FROM workspaces WHERE id = $1 AND guild_id = $2 RETURNING *;
