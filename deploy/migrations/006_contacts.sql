-- Create contact_requests table to track invitations
CREATE TABLE IF NOT EXISTS contact_requests (
    id           BIGSERIAL PRIMARY KEY,
    from_user_id BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    to_user_id   BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    status       TEXT NOT NULL DEFAULT 'pending', -- pending, accepted, rejected
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    responded_at TIMESTAMPTZ,
    UNIQUE (from_user_id, to_user_id)
);

-- Index for faster lookup of requests for a specific user
CREATE INDEX IF NOT EXISTS idx_contact_requests_to_user_id ON contact_requests (to_user_id);
CREATE INDEX IF NOT EXISTS idx_contact_requests_from_user_id ON contact_requests (from_user_id);

-- Create contacts table for confirmed friendships
-- We store two rows for each friendship for easier querying (A -> B and B -> A)
CREATE TABLE IF NOT EXISTS contacts (
    user_id    BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    contact_id BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, contact_id)
);

-- Index for listing contacts of a user
CREATE INDEX IF NOT EXISTS idx_contacts_user_id ON contacts (user_id);
