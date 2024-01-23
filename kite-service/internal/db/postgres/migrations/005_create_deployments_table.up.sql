CREATE TABLE IF NOT EXISTS deployments (
   id TEXT PRIMARY KEY,
   -- Unique identifier for the deployment, it is the same as the plugin key if the deployment is associated with a plugin
   key TEXT NOT NULL,
   name TEXT NOT NULL,
   description TEXT NOT NULL,
   guild_id TEXT NOT NULL REFERENCES guilds(id) ON DELETE CASCADE,

   -- Deployments can be associated with a public plugin but don't have to be
   plugin_version_id TEXT REFERENCES plugin_versions(id) ON DELETE RESTRICT,
   -- Deployments can be associated with a workspace but don't have to be
   workspace_id TEXT REFERENCES workspaces(id) ON DELETE SET NULL,
   
   -- Copied over from the plugin version if the deployment is associated with a plugin
   wasm_bytes BYTEA NOT NULL,
   manifest_default_config JSONB,
   manifest_events TEXT[],
   manifest_commands TEXT[],

   config JSONB,
   
   created_at TIMESTAMP NOT NULL,
   updated_at TIMESTAMP NOT NULL,

   UNIQUE (key, guild_id)
);
