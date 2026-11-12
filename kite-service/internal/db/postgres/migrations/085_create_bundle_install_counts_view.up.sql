-- How many apps added each bundle, and how many of them ran a flow in the last
-- 30 days. Apps of the bundle's owner don't count, so creators can't raise the
-- numbers of their templates themselves.
CREATE VIEW bundle_install_counts AS
SELECT
    bundle_installs.bundle_id,
    count(*) AS total,
    count(*) FILTER (WHERE EXISTS (
        SELECT 1 FROM usage_records
        WHERE usage_records.app_id = bundle_installs.app_id
            AND usage_records.created_at > LOCALTIMESTAMP - interval '30 days'
    )) AS active
FROM bundle_installs
JOIN bundles ON bundles.id = bundle_installs.bundle_id
JOIN apps ON apps.id = bundle_installs.app_id
WHERE apps.owner_user_id IS DISTINCT FROM bundles.owner_user_id
GROUP BY bundle_installs.bundle_id;
