ALTER TABLE gallery_albums ADD COLUMN publish_at TIMESTAMPTZ NULL;
ALTER TABLE gallery_albums ADD COLUMN expire_at TIMESTAMPTZ NULL;
CREATE INDEX idx_gallery_albums_publish_at ON gallery_albums(publish_at);
CREATE INDEX idx_gallery_albums_expire_at ON gallery_albums(expire_at);