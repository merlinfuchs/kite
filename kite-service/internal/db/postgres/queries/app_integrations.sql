-- name: GetAppIntegrations :many
SELECT * FROM app_integrations WHERE app_id = $1 ORDER BY integration_id;

-- name: SetAppIntegrationEnabled :one
INSERT INTO app_integrations (
    app_id,
    integration_id,
    enabled,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5
)
ON CONFLICT (app_id, integration_id) DO UPDATE SET
    enabled = EXCLUDED.enabled,
    updated_at = EXCLUDED.updated_at
RETURNING *;

-- name: CreateAppIntegrationIfMissing :exec
INSERT INTO app_integrations (
    app_id,
    integration_id,
    enabled,
    created_at,
    updated_at
) VALUES (
    $1, $2, TRUE, $3, $4
)
ON CONFLICT (app_id, integration_id) DO NOTHING;

-- name: UpdateAppIntegrationEnabled :one
UPDATE app_integrations SET
    enabled = $3,
    updated_at = $4
WHERE app_id = $1 AND integration_id = $2
RETURNING *;

-- name: DeleteAppIntegration :execrows
DELETE FROM app_integrations WHERE app_id = $1 AND integration_id = $2;
