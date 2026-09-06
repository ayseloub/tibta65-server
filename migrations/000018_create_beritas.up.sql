CREATE TABLE IF NOT EXISTS beritas (
    id CHAR(26) PRIMARY KEY,
    title VARCHAR(60) NOT NULL,
    slug VARCHAR(120) NOT NULL UNIQUE,
    description VARCHAR(1000) NOT NULL,
    image_url TEXT NOT NULL,
    visibility VARCHAR(20) NOT NULL DEFAULT 'public',
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    author_id CHAR(26) NULL REFERENCES admins(id) ON DELETE SET NULL,
    author_name VARCHAR(150) NOT NULL,
    event_date DATE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_beritas_status ON beritas(status);
CREATE INDEX idx_beritas_visibility ON beritas(visibility);