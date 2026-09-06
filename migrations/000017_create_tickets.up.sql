CREATE TABLE IF NOT EXISTS tickets (
    id CHAR(26) PRIMARY KEY,
    member_id CHAR(26) NOT NULL REFERENCES members(id) ON DELETE CASCADE,
    subject VARCHAR(150) NOT NULL,
    message TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'open',
    admin_reply TEXT NULL,
    replied_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_tickets_member_id ON tickets(member_id);
CREATE INDEX idx_tickets_status ON tickets(status);
CREATE INDEX idx_tickets_member_created ON tickets(member_id, created_at);