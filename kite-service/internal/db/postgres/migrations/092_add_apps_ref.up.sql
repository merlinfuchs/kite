-- Where the app came from, like ref=xenon on a link that adds the Kite bot.
-- It's set when the app is created and never changes.
ALTER TABLE apps ADD COLUMN IF NOT EXISTS ref TEXT;
