-- name: GetDeletedEntitiesSince :many
SELECT * FROM deleted_entities WHERE deleted_at > $1;

-- name: DeleteDeletedEntitiesBefore :exec
DELETE FROM deleted_entities WHERE deleted_at < $1;
