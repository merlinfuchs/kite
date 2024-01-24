-- name: CreateDeploymentLogEntry :exec
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
);

-- name: GetDeploymentLogEntries :many
SELECT * FROM deployment_logs WHERE deployment_id = $1 ORDER BY created_at ASC;

-- name: GetDeploymentLogSummary :one
SELECT
    deployment_id,
    COUNT(*) AS total_count,
    SUM(CASE WHEN level = 'ERROR' THEN 1 ELSE 0 END) AS error_count,
    SUM(CASE WHEN level = 'WARN' THEN 1 ELSE 0 END) AS warn_count,
    SUM(CASE WHEN level = 'INFO' THEN 1 ELSE 0 END) AS info_count,
    SUM(CASE WHEN level = 'DEBUG' THEN 1 ELSE 0 END) AS debug_count
FROM deployment_logs
WHERE deployment_id = $1 AND created_at >= $2
GROUP BY deployment_id;
