-- modules was never used. It becomes the installs of bundles, so it can sit
-- next to the bundles themselves once they're stored. A bundle is a template
-- for now. resources lists the items an install created, with the hash of
-- their flow at install, so updates can tell what was edited.
ALTER TABLE modules RENAME TO bundle_installs;
ALTER TABLE bundle_installs RENAME CONSTRAINT modules_pkey TO bundle_installs_pkey;
ALTER TABLE bundle_installs RENAME CONSTRAINT modules_app_id_fkey TO bundle_installs_app_id_fkey;

ALTER TABLE bundle_installs ADD COLUMN bundle_id TEXT NOT NULL DEFAULT '';
ALTER TABLE bundle_installs ADD COLUMN bundle_version INT NOT NULL DEFAULT 0;
-- The values the user filled in for the bundle's settings.
ALTER TABLE bundle_installs ADD COLUMN settings JSONB NOT NULL DEFAULT '{}';

CREATE INDEX IF NOT EXISTS bundle_installs_app_id ON bundle_installs (app_id);

ALTER TABLE commands RENAME COLUMN module_id TO bundle_install_id;
ALTER TABLE commands RENAME CONSTRAINT commands_module_id_fkey TO commands_bundle_install_id_fkey;
ALTER INDEX commands_module_id RENAME TO commands_bundle_install_id;

ALTER TABLE variables RENAME COLUMN module_id TO bundle_install_id;
ALTER TABLE variables RENAME CONSTRAINT variables_module_id_fkey TO variables_bundle_install_id_fkey;
ALTER INDEX variables_module_id RENAME TO variables_bundle_install_id;

ALTER TABLE messages RENAME COLUMN module_id TO bundle_install_id;
ALTER TABLE messages RENAME CONSTRAINT messages_module_id_fkey TO messages_bundle_install_id_fkey;
ALTER INDEX messages_module_id RENAME TO messages_bundle_install_id;

ALTER TABLE assets RENAME COLUMN module_id TO bundle_install_id;
ALTER TABLE assets RENAME CONSTRAINT assets_module_id_fkey TO assets_bundle_install_id_fkey;
ALTER INDEX assets_module_id RENAME TO assets_bundle_install_id;

ALTER TABLE event_listeners RENAME COLUMN module_id TO bundle_install_id;
ALTER TABLE event_listeners RENAME CONSTRAINT event_listeners_module_id_fkey TO event_listeners_bundle_install_id_fkey;
ALTER INDEX event_listeners_module_id RENAME TO event_listeners_bundle_install_id;
