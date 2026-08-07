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

-- variable_values has no app_id of its own, so every query below joins through
-- variables to scope by app. Without that a variable_id alone is enough to read
-- or write the row, and variable_ids reach the engine straight out of
-- user-authored flow JSON -- which is portable between apps via flow import.

-- name: GetVariableValue :one
SELECT variable_values.* FROM variable_values
JOIN variables ON variables.id = variable_values.variable_id
WHERE variable_values.variable_id = $1
  AND variable_values.scope IS NOT DISTINCT FROM $2
  AND variables.app_id = $3;

-- name: GetVariableValueForUpdate :one
-- FOR UPDATE OF, not a bare FOR UPDATE: the latter would also lock the joined
-- variables row for the duration of the transaction.
SELECT variable_values.* FROM variable_values
JOIN variables ON variables.id = variable_values.variable_id
WHERE variable_values.variable_id = $1
  AND variable_values.scope IS NOT DISTINCT FROM $2
  AND variables.app_id = $3
FOR UPDATE OF variable_values;

-- name: GetVariableValues :many
SELECT variable_values.* FROM variable_values
JOIN variables ON variables.id = variable_values.variable_id
WHERE variable_values.variable_id = $1 AND variables.app_id = $2;

-- name: SetVariableValue :one
-- INSERT ... SELECT rather than VALUES so the app check is part of the write:
-- if the variable does not belong to the app the select is empty, nothing is
-- inserted, and :one surfaces it as ErrNoRows.
INSERT INTO variable_values (
    variable_id,
    scope,
    value,
    created_at,
    updated_at
)
SELECT
    variables.id,
    sqlc.narg(scope)::text,
    @value::jsonb,
    @created_at::timestamp,
    @updated_at::timestamp
FROM variables
WHERE variables.id = @variable_id AND variables.app_id = @app_id
ON CONFLICT (variable_id, scope) DO UPDATE SET
    value = EXCLUDED.value,
    updated_at = EXCLUDED.updated_at
RETURNING *;

-- name: DeleteVariableValue :exec
-- IS NOT DISTINCT FROM so that unscoped values (scope IS NULL) are matched,
-- same as the get queries above. Plain `= NULL` never matches and made
-- deleting an unscoped variable value a silent no-op.
DELETE FROM variable_values
USING variables
WHERE variables.id = variable_values.variable_id
  AND variable_values.variable_id = $1
  AND variable_values.scope IS NOT DISTINCT FROM $2
  AND variables.app_id = $3;

-- name: DeleteAllVariableValues :exec
DELETE FROM variable_values
USING variables
WHERE variables.id = variable_values.variable_id
  AND variable_values.variable_id = $1
  AND variables.app_id = $2;