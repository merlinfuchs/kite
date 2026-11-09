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

-- name: GetVariablesByAppWithoutTotals :many
SELECT * FROM variables WHERE app_id = $1 ORDER BY created_at DESC;

-- name: CountVariablesByApp :one
SELECT COUNT(*) FROM variables WHERE app_id = $1;

-- name: CreateVariable :one
INSERT INTO variables (
    id,
    name,
    scoped,
    app_id,
    bundle_install_id,
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

-- name: DeleteVariable :one
DELETE FROM variables WHERE id = $1 RETURNING bundle_install_id;

-- variable_values has no app_id, so these reach the app through variables.

-- name: GetVariableValue :one
SELECT variable_values.* FROM variable_values
JOIN variables ON variables.id = variable_values.variable_id
WHERE variable_values.variable_id = $1
  AND variable_values.scope IS NOT DISTINCT FROM $2
  AND variables.app_id = $3;

-- name: GetVariableValueForUpdate :one
-- FOR UPDATE OF so the joined variables row isn't locked too
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
-- INSERT ... SELECT so a variable from another app inserts nothing (ErrNoRows)
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

-- Locks a value that may not exist yet, so updates that find no value wait for
-- each other instead of each storing a first one.
-- name: LockVariableValue :exec
SELECT pg_advisory_xact_lock(hashtext('variable:' || @variable_id::text || ':' || coalesce(sqlc.narg(scope)::text, '')));

-- name: InsertVariableValue :one
-- Like SetVariableValue, but inserts nothing (ErrNoRows) when the value
-- exists, also when another transaction inserted it and hasn't committed yet,
-- which it waits for.
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
ON CONFLICT (variable_id, scope) DO NOTHING
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

-- name: SearchVariableValues :many
-- position() instead of LIKE so % and _ in the search text match literally
SELECT variable_values.* FROM variable_values
JOIN variables ON variables.id = variable_values.variable_id
WHERE variable_values.variable_id = @variable_id
  AND variables.app_id = @app_id
  AND (
    @search::text = ''
    OR position(lower(@search::text) in lower(coalesce(variable_values.scope, ''))) > 0
  )
ORDER BY variable_values.updated_at DESC, variable_values.id DESC
LIMIT @row_limit OFFSET @row_offset;

-- name: ListVariableValues :many
-- Values sort as numbers when they are numbers or text of a number, like
-- "42". Other values come last when sorting by value. Values over 128 bytes,
-- like saved messages, can't be a number of at most 30 digits, so
-- pg_column_size skips them without reading them from TOAST, which was most of
-- the time on variables with many. starts_with instead of LIKE so % and _ in
-- the prefix match literally.
SELECT variable_values.* FROM variable_values
JOIN variables ON variables.id = variable_values.variable_id
WHERE variable_values.variable_id = @variable_id
  AND variables.app_id = @app_id
  AND (
    @scope_prefix::text = ''
    OR starts_with(coalesce(variable_values.scope, ''), @scope_prefix::text)
  )
ORDER BY
  CASE WHEN @sort::text = 'value_desc' THEN (
    CASE
      WHEN pg_column_size(variable_values.value) > 128 THEN NULL
      WHEN jsonb_typeof(variable_values.value->'v') = 'number' THEN (variable_values.value->>'v')::numeric
      WHEN jsonb_typeof(variable_values.value->'v') = 'string'
        AND variable_values.value->>'v' ~ '^\s*-?[0-9]{1,30}(\.[0-9]{1,30})?\s*$'
        THEN trim(variable_values.value->>'v')::numeric
    END
  ) END DESC NULLS LAST,
  CASE WHEN @sort::text = 'value_asc' THEN (
    CASE
      WHEN pg_column_size(variable_values.value) > 128 THEN NULL
      WHEN jsonb_typeof(variable_values.value->'v') = 'number' THEN (variable_values.value->>'v')::numeric
      WHEN jsonb_typeof(variable_values.value->'v') = 'string'
        AND variable_values.value->>'v' ~ '^\s*-?[0-9]{1,30}(\.[0-9]{1,30})?\s*$'
        THEN trim(variable_values.value->>'v')::numeric
    END
  ) END ASC NULLS LAST,
  CASE WHEN @sort::text = 'updated_desc' THEN variable_values.updated_at END DESC,
  CASE WHEN @sort::text = 'updated_asc' THEN variable_values.updated_at END ASC,
  variable_values.id
LIMIT @row_limit OFFSET @row_offset;

-- name: CountSearchVariableValues :one
SELECT COUNT(*) FROM variable_values
JOIN variables ON variables.id = variable_values.variable_id
WHERE variable_values.variable_id = @variable_id
  AND variables.app_id = @app_id
  AND (
    @search::text = ''
    OR position(lower(@search::text) in lower(coalesce(variable_values.scope, ''))) > 0
  );

-- name: DeleteVariablesByBundleInstall :exec
DELETE FROM variables WHERE bundle_install_id = @bundle_install_id AND NOT (id = ANY(@keep_ids::TEXT[]));
