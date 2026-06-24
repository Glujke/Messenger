CREATE TABLE IF NOT EXISTS attachments (
    id           BIGSERIAL PRIMARY KEY,
    room_id      BIGINT NOT NULL REFERENCES rooms (id) ON DELETE CASCADE,
    uploader_id  BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    filename     TEXT NOT NULL,
    content_type TEXT NOT NULL,
    size_bytes   BIGINT NOT NULL,
    storage_key  TEXT NOT NULL UNIQUE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_attachments_room_id ON attachments (room_id);

ALTER TABLE messages
    DROP CONSTRAINT IF EXISTS fk_messages_attachment;

ALTER TABLE messages
    ADD CONSTRAINT fk_messages_attachment
    FOREIGN KEY (attachment_id) REFERENCES attachments (id);
