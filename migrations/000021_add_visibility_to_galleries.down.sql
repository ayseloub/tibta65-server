DROP INDEX IF EXISTS idx_gallery_albums_visibility;
ALTER TABLE gallery_albums DROP COLUMN IF EXISTS visibility;