ALTER TABLE members ADD COLUMN approved_at TIMESTAMPTZ NULL;

UPDATE members SET approved_at = now();