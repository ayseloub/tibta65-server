ALTER TABLE beritas ADD COLUMN expire_at TIMESTAMPTZ NULL;
CREATE INDEX idx_beritas_expire_at ON beritas(expire_at);