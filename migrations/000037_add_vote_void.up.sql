ALTER TABLE votes ADD COLUMN voided_at TIMESTAMPTZ NULL;

CREATE TABLE vote_void_logs (
    id CHAR(26) PRIMARY KEY,
    member_id CHAR(26) NOT NULL,
    member_name VARCHAR(255) NOT NULL,
    member_number VARCHAR(30) NOT NULL,
    action VARCHAR(10) NOT NULL,
    reason TEXT NOT NULL,
    admin_id CHAR(26) NOT NULL REFERENCES admins(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_vote_void_logs_member ON vote_void_logs(member_id, created_at);

ALTER TABLE votes DROP CONSTRAINT votes_member_id_fkey;
ALTER TABLE votes ADD CONSTRAINT votes_member_id_fkey
    FOREIGN KEY (member_id) REFERENCES members(id) ON DELETE RESTRICT;