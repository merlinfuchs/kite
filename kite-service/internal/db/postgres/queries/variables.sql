-- name: GetVariable :one
SELECT * FROM variables WHERE id = $1;

-- name: GetVariableByName :one
SELECT * FROM variables WHERE app_id = $1 AND name = $2;

-- name: GetVariablesByApp :many
SELECT * FROM variables WHERE app_id = $1 ORDER BY created_at DESC;

-- name: CountVariablesByApp :one
SELECT COUNT(*) FROM variables WHERE app_id = $1;

-- name: CreateVariable :one
INSERT INTO variables (
    id,
    name,
    scope,
    type,
    app_id,
    module_id,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: UpdateVariable :one
UPDATE variables SET
    name = $2,
    scope = $3,
    type = $4,
    updated_at = $5
WHERE id = $1 RETURNING *;

-- name: DeleteVariable :exec
DELETE FROM variables WHERE id = $1;