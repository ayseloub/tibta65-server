ALTER TABLE kegiatans ADD COLUMN publish_at TIMESTAMPTZ NULL;
ALTER TABLE kegiatans ADD COLUMN expire_at TIMESTAMPTZ NULL;
CREATE INDEX idx_kegiatans_publish_at ON kegiatans(publish_at);
CREATE INDEX idx_kegiatans_expire_at ON kegiatans(expire_at);