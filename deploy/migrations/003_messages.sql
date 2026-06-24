CREATE TABLE IF NOT EXISTS messages (
    id            BIGSERIAL PRIMARY KEY,
    room_id       BIGINT NOT NULL REFERENCES rooms (id) ON DELETE CASCADE,
    sender_id     BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    type          TEXT NOT NULL,
    body          TEXT NOT NULL,
    attachment_id BIGINT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_messages_room_id_created_at ON messages (room_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_messages_room_id_id ON messages (room_id, id DESC);
