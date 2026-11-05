CREATE TABLE IF NOT EXISTS content_tables (
    id TEXT PRIMARY KEY,
    app_id TEXT NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    columns JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    UNIQUE (app_id, name)
);

CREATE TABLE IF NOT EXISTS content_table_rows (
    id TEXT PRIMARY KEY,
    table_id TEXT NOT NULL REFERENCES content_tables(id) ON DELETE CASCADE,
    data JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS content_tables_app_id ON content_tables (app_id);
CREATE INDEX IF NOT EXISTS content_table_rows_table_id ON content_table_rows (table_id);
