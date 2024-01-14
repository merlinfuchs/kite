CREATE TABLE IF NOT EXISTS deployments (
   id TEXT PRIMARY KEY,
   name TEXT NOT NULL,
   description TEXT NOT NULL,
   guild_id TEXT NOT NULL,

   -- Deployments can be associated with a public plugin but don't have to be
   plugin_version_id TEXT REFERENCES plugin_versions(id) ON DELETE RESTRICT,
   
   -- Copied over from the plugin version if the deployment is associated with a plugin
   wasm_bytes BYTEA NOT NULL,
   manifest_default_config JSONB,
   manifest_events TEXT[],
   manifest_commands TEXT[],

   config JSONB,
   
   created_at TIMESTAMP NOT NULL,
   updated_at TIMESTAMP NOT NULL
);
