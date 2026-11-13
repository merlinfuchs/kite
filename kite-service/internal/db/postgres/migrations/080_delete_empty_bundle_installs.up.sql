-- Deleting the last item of an install now deletes the install. Before, an app
-- that deleted everything a template added kept the install, which blocked
-- adding the template again.
DELETE FROM bundle_installs b
WHERE NOT EXISTS (SELECT 1 FROM commands WHERE bundle_install_id = b.id)
    AND NOT EXISTS (SELECT 1 FROM event_listeners WHERE bundle_install_id = b.id)
    AND NOT EXISTS (SELECT 1 FROM messages WHERE bundle_install_id = b.id)
    AND NOT EXISTS (SELECT 1 FROM variables WHERE bundle_install_id = b.id);
