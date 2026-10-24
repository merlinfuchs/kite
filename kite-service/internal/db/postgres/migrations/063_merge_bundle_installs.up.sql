-- An app has a template once. Adding it again adds the items to the install
-- it has, so earlier installs of the same template are merged into the
-- oldest. Its version becomes the lowest of them, which offers the update
-- again where they differed.
CREATE TEMP TABLE bundle_install_merges AS
SELECT
    id,
    FIRST_VALUE(id) OVER w AS keep_id,
    created_at
FROM bundle_installs
WINDOW w AS (PARTITION BY app_id, bundle_id ORDER BY created_at, id);

UPDATE bundle_installs b SET
    -- An item that is in several installs, because it was deleted and added
    -- again, is the one the app still has, or else the latest.
    resources = (
        SELECT COALESCE(jsonb_agg(item ORDER BY created_at, ord), '[]')
        FROM (
            SELECT DISTINCT ON (r.item->>'key') r.item, m.created_at, r.ord
            FROM bundle_install_merges m
            JOIN bundle_installs o ON o.id = m.id
            CROSS JOIN LATERAL jsonb_array_elements(o.resources) WITH ORDINALITY AS r(item, ord)
            WHERE m.keep_id = b.id
            ORDER BY
                r.item->>'key',
                (
                    EXISTS (SELECT 1 FROM commands WHERE id = r.item->>'id' AND bundle_install_id = o.id)
                    OR EXISTS (SELECT 1 FROM event_listeners WHERE id = r.item->>'id' AND bundle_install_id = o.id)
                    OR EXISTS (SELECT 1 FROM messages WHERE id = r.item->>'id' AND bundle_install_id = o.id)
                    OR EXISTS (SELECT 1 FROM variables WHERE id = r.item->>'id' AND bundle_install_id = o.id)
                ) DESC,
                m.created_at DESC,
                r.ord DESC
        ) items
    ),
    -- Values picked later win.
    settings = (
        SELECT COALESCE(jsonb_object_agg(s.key, s.value ORDER BY m.created_at), '{}')
        FROM bundle_install_merges m
        JOIN bundle_installs o ON o.id = m.id
        CROSS JOIN LATERAL jsonb_each(o.settings) AS s(key, value)
        WHERE m.keep_id = b.id
    ),
    bundle_version = (
        SELECT MIN(o.bundle_version)
        FROM bundle_install_merges m
        JOIN bundle_installs o ON o.id = m.id
        WHERE m.keep_id = b.id
    )
WHERE b.id IN (SELECT keep_id FROM bundle_install_merges WHERE id <> keep_id);

UPDATE commands SET bundle_install_id = m.keep_id
FROM bundle_install_merges m
WHERE commands.bundle_install_id = m.id AND m.id <> m.keep_id;

UPDATE event_listeners SET bundle_install_id = m.keep_id
FROM bundle_install_merges m
WHERE event_listeners.bundle_install_id = m.id AND m.id <> m.keep_id;

UPDATE messages SET bundle_install_id = m.keep_id
FROM bundle_install_merges m
WHERE messages.bundle_install_id = m.id AND m.id <> m.keep_id;

UPDATE variables SET bundle_install_id = m.keep_id
FROM bundle_install_merges m
WHERE variables.bundle_install_id = m.id AND m.id <> m.keep_id;

DELETE FROM bundle_installs
WHERE id IN (SELECT id FROM bundle_install_merges WHERE id <> keep_id);

DROP TABLE bundle_install_merges;

CREATE UNIQUE INDEX IF NOT EXISTS bundle_installs_app_id_bundle_id ON bundle_installs (app_id, bundle_id);
