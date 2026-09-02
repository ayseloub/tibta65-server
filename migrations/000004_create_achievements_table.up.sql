CREATE TABLE IF NOT EXISTS achievements (
    id CHAR(26) PRIMARY KEY,
    year VARCHAR(4) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    image_url TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_achievements_year ON achievements(year);