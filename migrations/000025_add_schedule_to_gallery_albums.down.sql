DROP INDEX IF EXISTS idx_gallery_albums_publish_at;
DROP INDEX IF EXISTS idx_gallery_albums_expire_at;
ALTER TABLE gallery_albums DROP COLUMN IF EXISTS publish_at;
ALTER TABLE gallery_albums DROP COLUMN IF EXISTS expire_at;