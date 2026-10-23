-- name: GetModulesByApp :many
SELECT * FROM modules WHERE app_id = $1 ORDER BY created_at DESC, id;

-- name: CreateModule :one
INSERT INTO modules (
    id,
    name,
    description,
    enabled,
    app_id,
    creator_user_id,
    template_id,
    template_version,
    settings,
    resources,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
) RETURNING *;
