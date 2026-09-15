ALTER TABLE beritas ADD COLUMN publish_at TIMESTAMPTZ NULL;
CREATE INDEX idx_beritas_publish_at ON beritas(publish_at);