-- Flows stored unscoped variable values under the literal scope '<nil>'
-- because the empty scope template was stringified with fmt. Unscoped
-- values now use NULL. Where both exist, the '<nil>' row is the one flows
-- have been reading and writing, so it wins.
DELETE FROM variable_values AS n
USING variable_values AS v
WHERE n.variable_id = v.variable_id
  AND n.scope IS NULL
  AND v.scope = '<nil>';

UPDATE variable_values SET scope = NULL WHERE scope = '<nil>';
