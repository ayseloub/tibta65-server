ALTER TABLE beritas ADD COLUMN is_highlight BOOLEAN NOT NULL DEFAULT false;

CREATE UNIQUE INDEX idx_beritas_one_highlight ON beritas ((is_highlight)) WHERE is_highlight = true;