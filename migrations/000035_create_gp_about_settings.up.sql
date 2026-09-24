CREATE TABLE gp_about_settings (
    id CHAR(26) PRIMARY KEY,
    title VARCHAR(255) NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    stat_1 VARCHAR(100) NOT NULL DEFAULT '',
    stat_2 VARCHAR(100) NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO gp_about_settings (id, title, description, stat_1, stat_2)
VALUES ('01HGPABOUTSETTINGSSINGLETON01', '', '', '', '');