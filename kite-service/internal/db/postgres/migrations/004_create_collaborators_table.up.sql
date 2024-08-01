CREATE TABLE IF NOT EXISTS collaborators (
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    app_id TEXT NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    role TEXT NOT NULL,
   
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,

    PRIMARY KEY (user_id, app_id)
);
