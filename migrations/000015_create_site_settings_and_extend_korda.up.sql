CREATE TABLE IF NOT EXISTS site_settings (
    id CHAR(26) PRIMARY KEY,
    email VARCHAR(255) NULL,
    youtube_url TEXT NULL,
    instagram_url TEXT NULL,
    facebook_url TEXT NULL,
    address TEXT NULL,
    maps_embed_url TEXT NULL,
    maps_link TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO site_settings (id) VALUES ('01SITESETTINGSDEFAULT001');

ALTER TABLE kordas ADD COLUMN phone VARCHAR(30) NULL;
ALTER TABLE kordas ADD COLUMN admin_name VARCHAR(255) NULL;
ALTER TABLE kordas ADD COLUMN address TEXT NULL;