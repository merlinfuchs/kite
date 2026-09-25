-- name: GetDeletedEntitiesSince :many
SELECT * FROM deleted_entities WHERE deleted_at > $1;

-- name: DeleteDeletedEntitiesBefore :execrows
-- Batched so a large backlog doesn't hold one long transaction. The table has
-- no key, ctid identifies the rows within the statement.
DELETE FROM deleted_entities WHERE ctid IN (
    SELECT expired.ctid FROM deleted_entities expired
    WHERE expired.deleted_at < @before_at
    LIMIT @batch_size
);
