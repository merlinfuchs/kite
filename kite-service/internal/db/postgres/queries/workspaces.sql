-- name: GetWorkspaceForApp :one
SELECT * FROM workspaces WHERE id = $1 AND app_id = $2;

-- name: GetWorkspacesForApp :many
SELECT * FROM workspaces WHERE app_id = $1 ORDER BY updated_at DESC;

-- name: CreateWorkspace :one
INSERT INTO workspaces (
    id,
    app_id,
    type,
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
    $7,
    $8
) RETURNING *;

-- name: UpdateWorkspace :one
UPDATE workspaces SET
    name = $3,
    description = $4,
    files = $5,
    updated_at = $6
WHERE 
    id = $1 AND 
    app_id = $2 
RETURNING *;

-- name: DeleteWorkspace :one
DELETE FROM workspaces WHERE id = $1 AND app_id = $2 RETURNING *;
