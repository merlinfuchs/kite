DROP TRIGGER IF EXISTS plugin_instances_record_deleted ON plugin_instances;
DROP TRIGGER IF EXISTS event_listeners_record_deleted ON event_listeners;
DROP TRIGGER IF EXISTS commands_record_deleted ON commands;
DROP FUNCTION IF EXISTS record_deleted_entities();
DROP TABLE IF EXISTS deleted_entities;
