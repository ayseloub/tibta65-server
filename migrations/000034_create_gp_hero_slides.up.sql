CREATE TABLE gp_hero_slides (
    id CHAR(26) PRIMARY KEY,
    headline VARCHAR(255) NOT NULL,
    description TEXT NULL,
    primary_button_text VARCHAR(100) NULL,
    primary_button_link VARCHAR(255) NULL,
    secondary_button_text VARCHAR(100) NULL,
    secondary_button_link VARCHAR(255) NULL,
    image_url TEXT NOT NULL,
    position INT NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_gp_hero_slides_position ON gp_hero_slides(position);