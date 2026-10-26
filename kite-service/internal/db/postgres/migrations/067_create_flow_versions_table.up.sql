-- Earlier saves of the flows of commands and event listeners, so a change can
-- be undone. Only the newest few are kept per flow.
CREATE TABLE IF NOT EXISTS flow_versions (
    id TEXT PRIMARY KEY,
    app_id TEXT NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    -- No foreign key, so the history outlives a deleted command or event
    -- listener and it can be restored.
    target_type TEXT NOT NULL,
    target_id TEXT NOT NULL,

    -- What made the version, like the editor, auto-save or a template update.
    source TEXT NOT NULL,
    -- Null for the state a flow had before it got its first version.
    creator_user_id TEXT,
    flow_source JSONB NOT NULL,
    -- The rest of a deleted command or event listener, only set on the
    -- version made by deleting it.
    target_data JSONB,

    created_at TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS flow_versions_target_id_created_at
    ON flow_versions (target_id, created_at DESC);
CREATE INDEX IF NOT EXISTS flow_versions_app_id ON flow_versions (app_id);
CREATE INDEX IF NOT EXISTS flow_versions_deleted_created_at
    ON flow_versions (created_at) WHERE source = 'delete';
