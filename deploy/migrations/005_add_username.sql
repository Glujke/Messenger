-- Add username column to users table
ALTER TABLE users ADD COLUMN username TEXT;

-- Set NOT NULL and UNIQUE constraints
-- We assume the DB is cleared as per user's confirmation
ALTER TABLE users ALTER COLUMN username SET NOT NULL;
ALTER TABLE users ADD CONSTRAINT users_username_unique UNIQUE (username);

-- Index for faster lookups by username
CREATE INDEX IF NOT EXISTS idx_users_username ON users (username);
