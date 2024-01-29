CREATE TABLE IF NOT EXISTS users (
   id TEXT PRIMARY KEY,
   username TEXT NOT NULL,
   discriminator TEXT,
   global_name TEXT,
   avatar TEXT,
   public_flags INTEGER NOT NULL,
   
   created_at TIMESTAMP NOT NULL,
   updated_at TIMESTAMP NOT NULL
);
