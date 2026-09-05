CREATE TABLE IF NOT EXISTS pemilu_settings (
    id CHAR(26) PRIMARY KEY,
    start_at TIMESTAMPTZ NOT NULL,
    end_at TIMESTAMPTZ NOT NULL,
    closed_early_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS kandidats (
    id CHAR(26) PRIMARY KEY,
    full_name VARCHAR(255) NOT NULL,
    visi TEXT NOT NULL,
    misi TEXT NOT NULL,
    pangkat VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS votes (
    id CHAR(26) PRIMARY KEY,
    member_id CHAR(26) NOT NULL UNIQUE REFERENCES members(id) ON DELETE CASCADE,
    kandidat_id CHAR(26) NOT NULL REFERENCES kandidats(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_votes_kandidat_id ON votes(kandidat_id);

-- Seed 1 baris settings default, biar selalu ada row buat diedit (pola sama kayak background_content dulu)
INSERT INTO pemilu_settings (id, start_at, end_at)
VALUES ('01PEMILUSETTINGSDEFAULT01', now(), now() + interval '30 days');