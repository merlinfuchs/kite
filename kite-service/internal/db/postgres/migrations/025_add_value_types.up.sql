-- Update variable_values
UPDATE variable_values
SET value = jsonb_build_object(
    't',
    CASE
        WHEN jsonb_typeof(value) = 'string' THEN 'string'
        WHEN jsonb_typeof(value) = 'boolean' THEN 'bool'
        WHEN jsonb_typeof(value) = 'array' THEN 'array'
        WHEN jsonb_typeof(value) = 'object' THEN 'object'
        WHEN jsonb_typeof(value) = 'number' THEN
            CASE
                WHEN (value::text ~ '^-?\d+$') THEN 'int'
                ELSE 'float'
            END
        ELSE 'any'
    END,
    'v',
    value
);

-- Update plugin_values
UPDATE plugin_values
SET value = jsonb_build_object(
    't',
    CASE
        WHEN jsonb_typeof(value) = 'string' THEN 'string'
        WHEN jsonb_typeof(value) = 'boolean' THEN 'bool'
        WHEN jsonb_typeof(value) = 'array' THEN 'array'
        WHEN jsonb_typeof(value) = 'object' THEN 'object'
        WHEN jsonb_typeof(value) = 'number' THEN
            CASE
                WHEN (value::text ~ '^-?\d+$') THEN 'int'
                ELSE 'float'
            END
        ELSE 'any'
    END,
    'v',
    value
);
