ALTER TABLE gallery_albums ADD COLUMN visibility VARCHAR(20) NOT NULL DEFAULT 'public';
CREATE INDEX idx_gallery_albums_visibility ON gallery_albums(visibility);