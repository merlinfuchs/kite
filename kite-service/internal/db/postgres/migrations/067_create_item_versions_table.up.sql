-- Earlier saves of commands, event listeners and messages, so a change can be
-- undone. Only the newest few are kept per item.
CREATE TABLE IF NOT EXISTS item_versions (
    id TEXT PRIMARY KEY,
    app_id TEXT NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    -- No foreign key, so the history outlives a deleted item and it can be
    -- restored.
    target_type TEXT NOT NULL,
    target_id TEXT NOT NULL,

    -- What made the version, like the editor or a template update.
    source TEXT NOT NULL,
    -- Null for the state an item had before it got its first version.
    creator_user_id TEXT,
    -- What the editor edits: the flow of a command or event listener, the
    -- data and flows of a message.
    content JSONB NOT NULL,
    -- The rest of a deleted item, only set on the version made by deleting
    -- it.
    target_data JSONB,

    created_at TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS item_versions_target_id_created_at
    ON item_versions (target_id, created_at DESC);
CREATE INDEX IF NOT EXISTS item_versions_app_id ON item_versions (app_id);
CREATE INDEX IF NOT EXISTS item_versions_deleted_created_at
    ON item_versions (created_at) WHERE source = 'delete';
