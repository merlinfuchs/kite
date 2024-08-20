-- name: GetVariable :one
SELECT sqlc.embed(variables), COUNT(variable_values.*) as total_values FROM variables 
LEFT JOIN variable_values ON variables.id = variable_values.variable_id
WHERE variables.id = $1
GROUP BY variables.id;

-- name: GetVariableByName :one
SELECT sqlc.embed(variables), COUNT(variable_values.*) as total_values FROM variables 
LEFT JOIN variable_values ON variables.id = variable_values.variable_id
WHERE app_id = $1 AND name = $2
GROUP BY variables.id;

-- name: GetVariablesByApp :many
SELECT sqlc.embed(variables), COUNT(variable_values.*) as total_values FROM variables 
LEFT JOIN variable_values ON variables.id = variable_values.variable_id
WHERE variables.app_id = $1 
GROUP BY variables.id
ORDER BY variables.created_at DESC;

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