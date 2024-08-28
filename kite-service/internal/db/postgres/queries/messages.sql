-- name: GetMessage :one
SELECT * FROM messages WHERE id = $1;

-- name: GetMessagesByApp :many
SELECT * FROM messages WHERE app_id = $1 ORDER BY created_at DESC;

-- name: CountMessagesByApp :one
SELECT COUNT(*) FROM messages WHERE app_id = $1;

/*
 id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,

    data JSONB NOT NULL, -- message data
    flow_sources JSONB NOT NULL, -- map of flow source ids to flow source objects

    app_id TEXT NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    module_id TEXT REFERENCES modules(id) ON DELETE SET NULL,

    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
    */

-- name: CreateMessage :one
INSERT INTO messages (
    id,
    name,
    description,
    app_id,
    module_id,
    creator_user_id,
    data,
    flow_sources,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
) RETURNING *;

-- name: UpdateMessage :one
UPDATE messages SET
    name = $2,
    description = $3,
    data = $4,
    flow_sources = $5,
    updated_at = $6
WHERE id = $1 RETURNING *;

-- name: DeleteMessage :exec
DELETE FROM messages WHERE id = $1;