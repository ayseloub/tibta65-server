CREATE TABLE IF NOT EXISTS members (
    id CHAR(26) PRIMARY KEY,
    full_name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash TEXT NULL,
    google_id VARCHAR(255) UNIQUE NULL,
    avatar_url TEXT NULL,
    korda_id CHAR(26) NOT NULL REFERENCES kordas(id) ON DELETE RESTRICT,
    must_change_password BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_members_korda_id ON members(korda_id);