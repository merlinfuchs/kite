CREATE TABLE IF NOT EXISTS variables (
    id TEXT PRIMARY KEY,
    scope TEXT NOT NULL, -- global, guild, user, member, channel, custom
    name TEXT NOT NULL,
    type TEXT NOT NULL, -- string, number, boolean, array, object, ...

    app_id TEXT NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    module_id TEXT REFERENCES modules(id) ON DELETE SET NULL,
   
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,

    UNIQUE (app_id, name)
);
