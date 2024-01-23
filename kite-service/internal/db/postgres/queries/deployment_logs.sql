-- name: CreateDeploymentLogEntry :one
INSERT INTO deployment_logs (
    id,
    deployment_id,
    level,
    message,
    created_at
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
) RETURNING *;

-- name: GetDeploymentLogEntries :many
SELECT * FROM deployment_logs WHERE deployment_id = $1 ORDER BY created_at DESC;