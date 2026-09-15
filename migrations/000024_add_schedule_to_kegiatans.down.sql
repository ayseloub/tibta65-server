DROP INDEX IF EXISTS idx_kegiatans_publish_at;
DROP INDEX IF EXISTS idx_kegiatans_expire_at;
ALTER TABLE kegiatans DROP COLUMN IF EXISTS publish_at;
ALTER TABLE kegiatans DROP COLUMN IF EXISTS expire_at;