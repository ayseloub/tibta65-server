CREATE TABLE IF NOT EXISTS hero_slides (
    id CHAR(26) PRIMARY KEY,
    headline VARCHAR(60) NOT NULL,
    description VARCHAR(200) NOT NULL,
    primary_button_text VARCHAR(50) NOT NULL,
    primary_button_link VARCHAR(255) NOT NULL,
    secondary_button_text VARCHAR(50) NULL,
    secondary_button_link VARCHAR(255) NULL,
    image_url TEXT NOT NULL,
    bg_color VARCHAR(9) NOT NULL DEFAULT '#F8EFDC',
    position INT NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    publish_at TIMESTAMPTZ NULL,
    expire_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_hero_slides_position ON hero_slides(position);
CREATE INDEX idx_hero_slides_status ON hero_slides(status);