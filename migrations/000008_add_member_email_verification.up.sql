ALTER TABLE members ADD COLUMN email_verified_at TIMESTAMPTZ NULL;

CREATE TABLE IF NOT EXISTS member_otps (
    id CHAR(26) PRIMARY KEY,
    member_id CHAR(26) NOT NULL REFERENCES members(id) ON DELETE CASCADE,
    otp_code VARCHAR(6) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_member_otps_member_id ON member_otps(member_id);