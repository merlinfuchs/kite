CREATE TABLE IF NOT EXISTS guild_entitlements (
   id TEXT PRIMARY KEY,
   guild_id TEXT NOT NULL REFERENCES guilds(id) ON DELETE CASCADE,
   user_id TEXT REFERENCES users(id) ON DELETE SET NULL,

   source TEXT NOT NULL, -- "discord" / "manual" / "default"
   source_id TEXT, -- e.g. the entitlement id from discord

   name TEXT,
   description TEXT,

   feature_monthly_cpu_time_limit INTEGER NOT NULL, -- in milliseconds
   feature_monthly_cpu_time_additive BOOLEAN NOT NULL,
   
   created_at TIMESTAMP NOT NULL,
   updated_at TIMESTAMP NOT NULL,
   valid_from TIMESTAMP,
   valid_until TIMESTAMP,

   UNIQUE(guild_id, source, source_id)
);

CREATE VIEW guild_entitlements_resolved_view AS 
SELECT 
    guild_id, 
    (
        MAX(CASE WHEN feature_monthly_cpu_time_additive THEN 0 ELSE feature_monthly_cpu_time_limit END) +
        SUM(CASE WHEN feature_monthly_cpu_time_additive THEN feature_monthly_cpu_time_limit ELSE 0 END)
    ) AS feature_monthly_cpu_time_limit
FROM guild_entitlements 
WHERE (valid_from IS NULL OR valid_from <= NOW()) AND (valid_until IS NULL OR valid_until >= NOW())
GROUP BY guild_id;

