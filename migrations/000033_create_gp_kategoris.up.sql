CREATE TABLE gp_kategoris (
    id CHAR(26) PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE gp_kegiatans ADD COLUMN kategori_id CHAR(26) NULL REFERENCES gp_kategoris(id) ON DELETE RESTRICT;
CREATE INDEX idx_gp_kegiatans_kategori_id ON gp_kegiatans(kategori_id);