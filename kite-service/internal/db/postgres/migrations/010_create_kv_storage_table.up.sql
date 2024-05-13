CREATE TABLE IF NOT EXISTS kv_storage (
   app_id TEXT NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
   namespace TEXT NOT NULL,
   key TEXT NOT NULL,
   value JSONB NOT NULL,
   
   created_at TIMESTAMP NOT NULL,
   updated_at TIMESTAMP NOT NULL,

   PRIMARY KEY (app_id, namespace, key)
);
