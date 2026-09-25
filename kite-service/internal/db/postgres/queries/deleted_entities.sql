-- name: GetDeletedEntitiesSince :many
SELECT * FROM deleted_entities WHERE deleted_at > $1;

-- name: DeleteDeletedEntitiesBefore :execrows
-- Batched so a large backlog doesn't hold one long transaction. The table has
-- no key, ctid identifies the rows within the statement. = ANY(ARRAY(...))
-- plans as a TID scan, IN (...) would join against a full scan.
DELETE FROM deleted_entities WHERE ctid = ANY(ARRAY(
    SELECT expired.ctid FROM deleted_entities expired
    WHERE expired.deleted_at < @before_at
    LIMIT @batch_size
));
