-- name: GetAppSecretsByApp :many
SELECT * FROM app_secrets WHERE app_id = $1 AND name IS NOT NULL ORDER BY name;

-- name: GetAppSecret :one
SELECT * FROM app_secrets WHERE app_id = $1 AND id = $2 AND name IS NOT NULL;

-- name: GetAppSecretsByNames :many
SELECT * FROM app_secrets WHERE app_id = $1 AND name = ANY(sqlc.arg(names)::TEXT[]);

-- name: CountAppSecretsByApp :one
SELECT COUNT(*) FROM app_secrets WHERE app_id = $1 AND name IS NOT NULL;

-- name: CreateAppSecret :one
INSERT INTO app_secrets (
    id,
    app_id,
    name,
    value_encrypted,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: UpdateAppSecret :one
UPDATE app_secrets SET
    name = $3,
    value_encrypted = $4,
    updated_at = $5
WHERE app_id = $1 AND id = $2 AND name IS NOT NULL RETURNING *;

-- name: DeleteAppSecret :execrows
DELETE FROM app_secrets WHERE app_id = $1 AND id = $2 AND name IS NOT NULL;

-- Integration credentials are the secrets with an integration_id instead of
-- a name. They're removed with their row in app_integrations.

-- name: GetAppIntegrationCredentials :many
SELECT * FROM app_secrets WHERE app_id = $1 AND integration_id IS NOT NULL ORDER BY integration_id;

-- name: GetAppIntegrationCredential :one
SELECT * FROM app_secrets WHERE app_id = $1 AND integration_id = $2;

-- name: SetAppIntegrationCredential :one
INSERT INTO app_secrets (
    id,
    app_id,
    integration_id,
    value_encrypted,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6
)
ON CONFLICT (app_id, integration_id) WHERE integration_id IS NOT NULL DO UPDATE SET
    value_encrypted = EXCLUDED.value_encrypted,
    updated_at = EXCLUDED.updated_at
RETURNING *;
