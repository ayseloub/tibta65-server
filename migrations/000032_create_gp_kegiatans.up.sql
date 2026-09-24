CREATE TABLE gp_kegiatans (
    id CHAR(26) PRIMARY KEY,
    slug VARCHAR(255) UNIQUE NOT NULL,
    title VARCHAR(255) NOT NULL,
    date DATE NOT NULL,
    korda_id CHAR(26) NOT NULL REFERENCES kordas(id) ON DELETE RESTRICT,
    location VARCHAR(255) NOT NULL,
    image_url TEXT NOT NULL,
    description TEXT NOT NULL,
    publish_at TIMESTAMPTZ NULL,
    expire_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_gp_kegiatans_korda_id ON gp_kegiatans(korda_id);
CREATE INDEX idx_gp_kegiatans_date ON gp_kegiatans(date);