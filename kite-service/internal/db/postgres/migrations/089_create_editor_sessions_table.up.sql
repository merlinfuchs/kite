-- How long and how often the commands and event listeners of an app were
-- worked on in the editor. A session is one opened editor, and starts with the
-- first input or save in it.
CREATE TABLE IF NOT EXISTS editor_sessions (
    -- Made by the editor, which also sends it with its saves.
    id TEXT PRIMARY KEY,
    app_id TEXT NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    command_id TEXT REFERENCES commands(id) ON DELETE CASCADE,
    event_listener_id TEXT REFERENCES event_listeners(id) ON DELETE CASCADE,

    active_seconds INTEGER NOT NULL DEFAULT 0,
    save_count INTEGER NOT NULL DEFAULT 0,

    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    last_saved_at TIMESTAMP,

    CHECK (num_nonnulls(command_id, event_listener_id) = 1)
);

CREATE INDEX IF NOT EXISTS editor_sessions_app_id_created_at
    ON editor_sessions (app_id, created_at);
CREATE INDEX IF NOT EXISTS editor_sessions_created_at
    ON editor_sessions (created_at);
CREATE INDEX IF NOT EXISTS editor_sessions_command_id
    ON editor_sessions (command_id) WHERE command_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS editor_sessions_event_listener_id
    ON editor_sessions (event_listener_id) WHERE event_listener_id IS NOT NULL;
