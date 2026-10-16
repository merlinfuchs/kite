-- name: GetFunction :one
SELECT * FROM functions WHERE id = $1 AND app_id = $2;

-- name: GetFunctionsByApp :many
SELECT * FROM functions WHERE app_id = $1 ORDER BY created_at DESC;

-- name: CountFunctionsByApp :one
SELECT COUNT(*) FROM functions WHERE app_id = $1;

-- name: CreateFunction :one
INSERT INTO functions (
    id,
    name,
    description,
    app_id,
    module_id,
    creator_user_id,
    flow_source,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING *;

-- name: UpdateFunction :one
UPDATE functions SET
    name = $3,
    description = $4,
    flow_source = $5,
    updated_at = $6
WHERE id = $1 AND app_id = $2 RETURNING *;

-- name: DeleteFunction :exec
DELETE FROM functions WHERE id = $1 AND app_id = $2;
