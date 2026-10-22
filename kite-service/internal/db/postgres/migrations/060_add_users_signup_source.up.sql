-- Where the user first came from before signing up, recorded by the website in
-- a cookie. Only set when the user is created.
ALTER TABLE users ADD COLUMN signup_referrer TEXT;
ALTER TABLE users ADD COLUMN signup_utm_source TEXT;
ALTER TABLE users ADD COLUMN signup_utm_medium TEXT;
ALTER TABLE users ADD COLUMN signup_utm_campaign TEXT;
ALTER TABLE users ADD COLUMN signup_landing_page TEXT;
