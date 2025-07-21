UPDATE variable_values SET value = value::jsonb->'v';
UPDATE plugin_values SET value = value::jsonb->'v';
