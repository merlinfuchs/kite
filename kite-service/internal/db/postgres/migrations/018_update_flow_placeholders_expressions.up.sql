UPDATE commands SET flow_source = REGEXP_REPLACE(flow_source #>> '{}', '{{nodes.([0-9]+)', '{{node(\1)', 'g')::jsonb;
UPDATE event_listeners SET flow_source = REGEXP_REPLACE(flow_source #>> '{}', '{{nodes.([0-9]+)', '{{node(\1)', 'g')::jsonb;
UPDATE messages SET flow_sources = REGEXP_REPLACE(flow_sources #>> '{}', '{{nodes.([0-9]+)', '{{node(\1)', 'g')::jsonb;
UPDATE message_instances SET flow_sources = REGEXP_REPLACE(flow_sources #>> '{}', '{{nodes.([0-9]+)', '{{node(\1)', 'g')::jsonb;

