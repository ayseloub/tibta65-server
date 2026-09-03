CREATE TABLE IF NOT EXISTS kordas (
    id CHAR(26) PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS kategoris (
    id CHAR(26) PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS kegiatans (
    id CHAR(26) PRIMARY KEY,
    slug VARCHAR(255) UNIQUE NOT NULL,
    title VARCHAR(255) NOT NULL,
    date DATE NOT NULL,
    korda_id CHAR(26) NOT NULL REFERENCES kordas(id) ON DELETE RESTRICT,
    kategori_id CHAR(26) NOT NULL REFERENCES kategoris(id) ON DELETE RESTRICT,
    location VARCHAR(255) NOT NULL,
    image_url TEXT NOT NULL,
    description TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_kegiatans_korda_id ON kegiatans(korda_id);
CREATE INDEX IF NOT EXISTS idx_kegiatans_kategori_id ON kegiatans(kategori_id);
CREATE INDEX IF NOT EXISTS idx_kegiatans_date ON kegiatans(date);