DROP INDEX IF EXISTS idx_beritas_expire_at;
ALTER TABLE beritas DROP COLUMN IF EXISTS expire_at;