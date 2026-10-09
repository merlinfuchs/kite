-- Prompts sent to the dashboard's AI assistant, like the one that edits flows.
-- They count towards the app's monthly limit, separate from credits, and keep
-- the tokens used to tune the limits.
CREATE TABLE IF NOT EXISTS assistant_prompts (
    id TEXT PRIMARY KEY,
    app_id TEXT NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL,

    model TEXT NOT NULL,
    -- The user's message.
    prompt TEXT NOT NULL,
    -- Model calls made for the prompt, including repairs of its edits.
    rounds INTEGER NOT NULL,
    -- Whether the AI made changes. Only prompts that did count towards the
    -- limit.
    edited BOOLEAN NOT NULL,
    input_tokens INTEGER NOT NULL,
    cached_input_tokens INTEGER NOT NULL,
    output_tokens INTEGER NOT NULL,

    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS assistant_prompts_app_id_created_at ON assistant_prompts (app_id, created_at);
