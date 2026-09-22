ALTER TABLE members RENAME COLUMN member_number TO legacy_member_number;
ALTER TABLE members ALTER COLUMN legacy_member_number DROP NOT NULL;
ALTER TABLE members RENAME CONSTRAINT members_member_number_unique TO members_legacy_member_number_unique;

ALTER TABLE members ADD COLUMN member_number VARCHAR(20) NULL;

ALTER TABLE members ADD COLUMN username VARCHAR(50) UNIQUE NULL;
ALTER TABLE members ADD COLUMN generation INT NOT NULL DEFAULT 1;
ALTER TABLE members ADD COLUMN parent_member_id CHAR(26) NULL REFERENCES members(id);
ALTER TABLE members ADD COLUMN status VARCHAR(20) NOT NULL DEFAULT 'active';
ALTER TABLE members ADD COLUMN nama_suci VARCHAR(100) NULL;
ALTER TABLE members ADD COLUMN agama VARCHAR(50) NULL;
ALTER TABLE members ADD COLUMN nrp VARCHAR(50) NULL;
ALTER TABLE members ADD COLUMN no_ak VARCHAR(50) NULL;
ALTER TABLE members ADD COLUMN pangkat_terakhir VARCHAR(100) NULL;
ALTER TABLE members ADD COLUMN legacy_identifier_raw TEXT NULL;
ALTER TABLE members ADD COLUMN profile_completed BOOLEAN NOT NULL DEFAULT true;

WITH numbered AS (
	SELECT id, ROW_NUMBER() OVER (ORDER BY created_at) AS rn
	FROM members
)
UPDATE members m
SET member_number = '65-G01-' || LPAD(numbered.rn::text, 4, '0')
FROM numbered
WHERE m.id = numbered.id;

ALTER TABLE members ALTER COLUMN member_number SET NOT NULL;
ALTER TABLE members ADD CONSTRAINT members_member_number_key UNIQUE (member_number);

CREATE INDEX idx_members_parent_id ON members(parent_member_id);
CREATE INDEX idx_members_generation ON members(generation);
CREATE INDEX idx_members_status ON members(status);

CREATE TABLE member_activation_tokens (
    id CHAR(26) PRIMARY KEY,
    member_id CHAR(26) NOT NULL REFERENCES members(id),
    token_hash TEXT NOT NULL,
    purpose VARCHAR(20) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_by_admin_id CHAR(26) NOT NULL REFERENCES admins(id),
    used_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_activation_tokens_member ON member_activation_tokens(member_id);

CREATE TABLE member_pending_changes (
    id CHAR(26) PRIMARY KEY,
    member_id CHAR(26) NOT NULL REFERENCES members(id),
    changes JSONB NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    reviewed_by_admin_id CHAR(26) NULL REFERENCES admins(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    reviewed_at TIMESTAMPTZ NULL
);
CREATE INDEX idx_pending_changes_member ON member_pending_changes(member_id);

ALTER TABLE tickets ADD COLUMN reported_member_id CHAR(26) NULL REFERENCES members(id);

ALTER TABLE pemilu_settings ADD COLUMN eligible_generations INTEGER[] NOT NULL DEFAULT '{}';

ALTER TABLE kegiatans ADD COLUMN target_generations INTEGER[] NULL;
ALTER TABLE beritas ADD COLUMN target_generations INTEGER[] NULL;
ALTER TABLE gallery_albums ADD COLUMN target_generations INTEGER[] NULL;