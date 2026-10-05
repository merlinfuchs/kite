-- Earlier saves of a command's or event listener's flow, so a save can be
-- undone from the editor. Only the newest few are kept per flow.
CREATE TABLE IF NOT EXISTS flow_versions (
    id TEXT PRIMARY KEY,
    app_id TEXT NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    command_id TEXT REFERENCES commands(id) ON DELETE CASCADE,
    event_listener_id TEXT REFERENCES event_listeners(id) ON DELETE CASCADE,

    flow_source JSONB NOT NULL,
    -- Saved by the editor's auto-save rather than by the user.
    auto_saved BOOLEAN NOT NULL DEFAULT FALSE,
    -- Null for the state a flow had before it got its first version.
    creator_user_id TEXT,

    created_at TIMESTAMP NOT NULL,

    CHECK ((command_id IS NULL) <> (event_listener_id IS NULL))
);

CREATE INDEX IF NOT EXISTS flow_versions_command_id_created_at
    ON flow_versions (command_id, created_at DESC) WHERE command_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS flow_versions_event_listener_id_created_at
    ON flow_versions (event_listener_id, created_at DESC) WHERE event_listener_id IS NOT NULL;
