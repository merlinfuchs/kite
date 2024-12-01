CREATE TABLE IF NOT EXISTS logs (
    id BIGSERIAL PRIMARY KEY,
    app_id TEXT NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    
    message TEXT NOT NULL,
    level TEXT NOT NULL,
   
    created_at TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS logs_app_id ON logs (app_id);
