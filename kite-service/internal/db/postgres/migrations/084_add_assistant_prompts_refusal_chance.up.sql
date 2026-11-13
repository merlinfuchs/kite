-- The chance of harm the decision model gave a prompt the flow AI refused,
-- so users who keep asking for bots that attack servers can be found.
ALTER TABLE assistant_prompts ADD COLUMN IF NOT EXISTS refusal_chance REAL;
