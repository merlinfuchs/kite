DROP INDEX IF EXISTS bundle_installs_app_id;

ALTER INDEX commands_bundle_install_id RENAME TO commands_module_id;
ALTER TABLE commands RENAME CONSTRAINT commands_bundle_install_id_fkey TO commands_module_id_fkey;
ALTER TABLE commands RENAME COLUMN bundle_install_id TO module_id;

ALTER INDEX variables_bundle_install_id RENAME TO variables_module_id;
ALTER TABLE variables RENAME CONSTRAINT variables_bundle_install_id_fkey TO variables_module_id_fkey;
ALTER TABLE variables RENAME COLUMN bundle_install_id TO module_id;

ALTER INDEX messages_bundle_install_id RENAME TO messages_module_id;
ALTER TABLE messages RENAME CONSTRAINT messages_bundle_install_id_fkey TO messages_module_id_fkey;
ALTER TABLE messages RENAME COLUMN bundle_install_id TO module_id;

ALTER INDEX assets_bundle_install_id RENAME TO assets_module_id;
ALTER TABLE assets RENAME CONSTRAINT assets_bundle_install_id_fkey TO assets_module_id_fkey;
ALTER TABLE assets RENAME COLUMN bundle_install_id TO module_id;

ALTER INDEX event_listeners_bundle_install_id RENAME TO event_listeners_module_id;
ALTER TABLE event_listeners RENAME CONSTRAINT event_listeners_bundle_install_id_fkey TO event_listeners_module_id_fkey;
ALTER TABLE event_listeners RENAME COLUMN bundle_install_id TO module_id;

ALTER TABLE bundle_installs DROP COLUMN settings;
ALTER TABLE bundle_installs DROP COLUMN bundle_version;
ALTER TABLE bundle_installs DROP COLUMN bundle_id;

ALTER TABLE bundle_installs RENAME CONSTRAINT bundle_installs_app_id_fkey TO modules_app_id_fkey;
ALTER TABLE bundle_installs RENAME CONSTRAINT bundle_installs_pkey TO modules_pkey;
ALTER TABLE bundle_installs RENAME TO modules;
