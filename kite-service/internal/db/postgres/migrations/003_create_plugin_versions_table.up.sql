CREATE TABLE IF NOT EXISTS plugin_versions (
   id TEXT PRIMARY KEY,
   plugin_id TEXT NOT NULL REFERENCES plugins(id) ON DELETE CASCADE,

   version_major INTEGER NOT NULL,
   version_minor INTEGER NOT NULL,
   version_patch INTEGER NOT NULL,
   
   wasm_bytes BYTEA NOT NULL,
   manifest_default_config JSONB,
   manifest_events TEXT[],
   manifest_commands TEXT[],
   
   created_at TIMESTAMP NOT NULL
);
