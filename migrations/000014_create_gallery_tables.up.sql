CREATE TABLE IF NOT EXISTS gallery_albums (
    id CHAR(26) PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    event_date DATE NOT NULL,
    korda_id CHAR(26) NULL REFERENCES kordas(id) ON DELETE SET NULL,
    kategori_id CHAR(26) NULL REFERENCES kategoris(id) ON DELETE SET NULL,
    is_highlight BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS gallery_photos (
    id CHAR(26) PRIMARY KEY,
    album_id CHAR(26) NOT NULL REFERENCES gallery_albums(id) ON DELETE CASCADE,
    image_url TEXT NOT NULL,
    caption TEXT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_gallery_photos_album_id ON gallery_photos(album_id);

CREATE UNIQUE INDEX IF NOT EXISTS idx_gallery_albums_one_highlight
    ON gallery_albums(is_highlight) WHERE is_highlight = true;