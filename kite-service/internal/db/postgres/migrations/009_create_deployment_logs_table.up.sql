CREATE TABLE IF NOT EXISTS deployment_logs (
   id BIGSERIAL PRIMARY KEY,
   deployment_id TEXT NOT NULL REFERENCES deployments(id) ON DELETE CASCADE,
   level TEXT NOT NULL,
   message TEXT NOT NULL,
   
   created_at TIMESTAMP NOT NULL
);
