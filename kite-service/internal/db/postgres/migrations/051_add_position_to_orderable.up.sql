-- Lets users reorder commands, event listeners and message templates in the
-- dashboard instead of always seeing them in creation order.
ALTER TABLE commands ADD COLUMN position INTEGER NOT NULL DEFAULT 0;
ALTER TABLE event_listeners ADD COLUMN position INTEGER NOT NULL DEFAULT 0;
ALTER TABLE messages ADD COLUMN position INTEGER NOT NULL DEFAULT 0;

-- Backfill existing rows so the newest-first order they already had (created_at DESC) is preserved.
WITH numbered AS (
    SELECT id, ROW_NUMBER() OVER (PARTITION BY app_id ORDER BY created_at DESC, id DESC) - 1 AS rn
    FROM commands
)
UPDATE commands SET position = numbered.rn FROM numbered WHERE commands.id = numbered.id;

WITH numbered AS (
    SELECT id, ROW_NUMBER() OVER (PARTITION BY app_id ORDER BY created_at DESC, id DESC) - 1 AS rn
    FROM event_listeners
)
UPDATE event_listeners SET position = numbered.rn FROM numbered WHERE event_listeners.id = numbered.id;

WITH numbered AS (
    SELECT id, ROW_NUMBER() OVER (PARTITION BY app_id ORDER BY created_at DESC, id DESC) - 1 AS rn
    FROM messages
)
UPDATE messages SET position = numbered.rn FROM numbered WHERE messages.id = numbered.id;

CREATE INDEX IF NOT EXISTS commands_app_id_position ON commands (app_id, position);
CREATE INDEX IF NOT EXISTS event_listeners_app_id_position ON event_listeners (app_id, position);
CREATE INDEX IF NOT EXISTS messages_app_id_position ON messages (app_id, position);
