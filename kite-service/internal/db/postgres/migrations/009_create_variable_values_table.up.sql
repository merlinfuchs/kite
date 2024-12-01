CREATE TABLE IF NOT EXISTS variable_values (
    id BIGSERIAL PRIMARY KEY,
    variable_id TEXT NOT NULL REFERENCES variables(id) ON DELETE CASCADE,
    scope TEXT, -- resolved guild, user, member, channel id, or custom
    value JSONB NOT NULL,
   
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,

    UNIQUE NULLS NOT DISTINCT (variable_id, scope)
);

CREATE INDEX IF NOT EXISTS variable_values_variable_id ON variable_values (variable_id);
CREATE INDEX IF NOT EXISTS variable_values_scope ON variable_values (scope);
