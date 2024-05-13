CREATE TABLE IF NOT EXISTS apps (
   id TEXT PRIMARY KEY,
   owner_user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,

   token TEXT NOT NULL,
   token_invalid BOOLEAN NOT NULL DEFAULT false,
   public_key TEXT NOT NULL,

   user_id TEXT NOT NULL,
   user_name TEXT NOT NULL,
   user_discriminator TEXT NOT NULL,
   user_avatar TEXT,
   user_banner TEXT,
   user_bio TEXT,

   status_type TEXT NOT NULL DEFAULT 'online',
   status_activity_type INT,
   status_activity_name TEXT,
   status_activity_state TEXT,
   status_activity_url TEXT,
   
   created_at TIMESTAMP NOT NULL,
   updated_at TIMESTAMP NOT NULL
);
