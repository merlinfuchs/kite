UPDATE commands SET flow_source = REGEXP_REPLACE(flow_source #>> '{}', '{{node\(([0-9]+)\)', '{{nodes.\1', 'g')::jsonb;
UPDATE event_listeners SET flow_source = REGEXP_REPLACE(flow_source #>> '{}', '{{node\(([0-9]+)\)', '{{nodes.\1', 'g')::jsonb;
UPDATE messages SET flow_sources = REGEXP_REPLACE(flow_sources #>> '{}', '{{node\(([0-9]+)\)', '{{nodes.\1', 'g')::jsonb;
UPDATE message_instances SET flow_sources = REGEXP_REPLACE(flow_sources #>> '{}', '{{node\(([0-9]+)\)', '{{nodes.\1', 'g')::jsonb;
