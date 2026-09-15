ALTER TABLE kegiatans ADD COLUMN visibility VARCHAR(20) NOT NULL DEFAULT 'public';
CREATE INDEX idx_kegiatans_visibility ON kegiatans(visibility);