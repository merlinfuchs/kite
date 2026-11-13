-- For counting the installs of a bundle, the unique index leads with app_id.
CREATE INDEX IF NOT EXISTS bundle_installs_bundle_id ON bundle_installs (bundle_id);
