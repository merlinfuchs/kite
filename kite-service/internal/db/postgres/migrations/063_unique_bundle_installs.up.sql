-- An app has a template once. The few apps that added one several times keep
-- the latest install, what the others added stays in the app as its own.
DELETE FROM bundle_installs b
USING bundle_installs newer
WHERE newer.app_id = b.app_id
    AND newer.bundle_id = b.bundle_id
    AND (newer.created_at, newer.id) > (b.created_at, b.id);

CREATE UNIQUE INDEX IF NOT EXISTS bundle_installs_app_id_bundle_id ON bundle_installs (app_id, bundle_id);
