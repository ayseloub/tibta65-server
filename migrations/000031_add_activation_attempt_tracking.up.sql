CREATE TABLE activation_attempts (
    id CHAR(26) PRIMARY KEY,
    member_number VARCHAR(20) NOT NULL,
    purpose VARCHAR(20) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_activation_attempts_lookup ON activation_attempts(member_number, purpose, created_at);