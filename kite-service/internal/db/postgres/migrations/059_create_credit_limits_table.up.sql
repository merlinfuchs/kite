-- Limits app owners set on how many credits a single server or user can use
-- per day or month. A NULL target_id is the default for every server or user
-- of the scope that has no limit of its own for that period, and a NULL
-- credits is no limit, so a specific server or user can be exempted from the
-- default.
CREATE TABLE IF NOT EXISTS credit_limits (
    id TEXT PRIMARY KEY,
    app_id TEXT NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    scope TEXT NOT NULL CHECK (scope IN ('guild', 'user')),
    target_id TEXT,
    period TEXT NOT NULL CHECK (period IN ('day', 'month')),
    credits INTEGER CHECK (credits >= 0),

    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,

    CHECK (credits IS NOT NULL OR target_id IS NOT NULL)
);

CREATE UNIQUE INDEX IF NOT EXISTS credit_limits_app_id_scope_target_id_period
    ON credit_limits (app_id, scope, COALESCE(target_id, ''), period);
