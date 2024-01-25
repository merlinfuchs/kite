CREATE TABLE IF NOT EXISTS kv_storage (
   guild_id TEXT NOT NULL REFERENCES guilds(id) ON DELETE CASCADE,
   namespace TEXT NOT NULL,
   key TEXT NOT NULL,
   value JSONB NOT NULL,
   
   created_at TIMESTAMP NOT NULL,
   updated_at TIMESTAMP NOT NULL,

   PRIMARY KEY (guild_id, namespace, key)
);
