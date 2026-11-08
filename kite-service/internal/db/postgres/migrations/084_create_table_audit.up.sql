CREATE TABLE content_table_audit (
    id TEXT PRIMARY KEY,
    app_id TEXT NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    table_id TEXT NOT NULL,
    actor_user_id TEXT NOT NULL,
    action TEXT NOT NULL,
    row_id TEXT,
    before_data JSONB,
    after_data JSONB,
    created_at TIMESTAMP NOT NULL,
    undone_at TIMESTAMP
);

CREATE INDEX content_table_audit_table_created
    ON content_table_audit (table_id, created_at DESC);
