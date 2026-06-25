-- Add kind, name and created_by to rooms table
ALTER TABLE rooms ADD COLUMN kind TEXT NOT NULL DEFAULT 'direct';
ALTER TABLE rooms ADD COLUMN name TEXT;
ALTER TABLE rooms ADD COLUMN created_by BIGINT REFERENCES users (id) ON DELETE SET NULL;

-- Index for kind to speed up filtering if needed
CREATE INDEX IF NOT EXISTS idx_rooms_kind ON rooms (kind);
