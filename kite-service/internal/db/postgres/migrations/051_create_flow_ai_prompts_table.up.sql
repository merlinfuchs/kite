-- Prompts sent to the flow AI. They count towards the app's monthly limit,
-- separate from credits, and keep the tokens used to tune the limits.
CREATE TABLE IF NOT EXISTS flow_ai_prompts (
    id TEXT PRIMARY KEY,
    app_id TEXT NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL,

    model TEXT NOT NULL,
    -- Model calls made for the prompt, including repairs of its edits.
    rounds INTEGER NOT NULL,
    input_tokens INTEGER NOT NULL,
    cached_input_tokens INTEGER NOT NULL,
    output_tokens INTEGER NOT NULL,

    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS flow_ai_prompts_app_id_created_at ON flow_ai_prompts (app_id, created_at);
