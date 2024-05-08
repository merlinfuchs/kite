CREATE TABLE IF NOT EXISTS deployments (
   id TEXT PRIMARY KEY,
   -- Unique identifier for the deployment, it is the same as the plugin key if the deployment is associated with a plugin
   key TEXT NOT NULL,
   name TEXT NOT NULL,
   description TEXT NOT NULL,
   app_id TEXT NOT NULL REFERENCES apps(id) ON DELETE CASCADE,

   -- Deployments can be associated with a public plugin but don't have to be
   plugin_version_id TEXT REFERENCES plugin_versions(id) ON DELETE RESTRICT,
   
   -- Copied over from the plugin version if the deployment is associated with a plugin
   wasm_bytes BYTEA NOT NULL,
   manifest JSONB NOT NULL,

   config JSONB NOT NULL,
   
   created_at TIMESTAMP NOT NULL,
   updated_at TIMESTAMP NOT NULL,
   deployed_at TIMESTAMP,

   UNIQUE (key, app_id)
);
