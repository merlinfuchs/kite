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
    scoped,
    app_id,
    module_id,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: UpdateVariable :one
UPDATE variables SET
    name = $2,
    scoped = $3,
    updated_at = $4
WHERE id = $1 RETURNING *;

-- name: DeleteVariable :exec
DELETE FROM variables WHERE id = $1;

-- name: GetVariableValue :one
SELECT * FROM variable_values WHERE variable_id = $1 AND scope IS NOT DISTINCT FROM $2;

-- name: GetVariableValues :many
SELECT * FROM variable_values WHERE variable_id = $1;

-- name: SetVariableValue :one
INSERT INTO variable_values (
    variable_id,
    scope,
    value,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5
) ON CONFLICT (variable_id, scope) DO UPDATE SET
    value = EXCLUDED.value,
    updated_at = EXCLUDED.updated_at
RETURNING *;

-- name: DeleteVariableValue :exec
DELETE FROM variable_values WHERE variable_id = $1 AND scope = $2;

-- name: DeleteAllVariableValues :exec
DELETE FROM variable_values WHERE variable_id = $1;