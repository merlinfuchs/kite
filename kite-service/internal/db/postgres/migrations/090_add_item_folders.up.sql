-- The folder an item is listed under in the dashboard. NULL lists it on its
-- own, or under the template it was added with.
ALTER TABLE commands ADD COLUMN IF NOT EXISTS folder TEXT;
ALTER TABLE event_listeners ADD COLUMN IF NOT EXISTS folder TEXT;
ALTER TABLE messages ADD COLUMN IF NOT EXISTS folder TEXT;
