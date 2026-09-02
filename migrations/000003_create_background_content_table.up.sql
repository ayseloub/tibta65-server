CREATE TABLE IF NOT EXISTS background_content (
    id CHAR(26) PRIMARY KEY,
    section VARCHAR(20) UNIQUE NOT NULL CHECK (section IN ('about', 'sejarah')),
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO background_content (id, section, title, description) VALUES
(
    '01JC0000000000000000000002',
    'about',
    'Tentang TIBTA 65',
    'Tibta 65 adalah perhimpunan keluarga besar alumni yang berdiri atas dasar kebersamaan, kekeluargaan, dan semangat saling menopang. Anggota Tibta 65 tersebar di berbagai daerah dan terhimpun dalam koordinator daerah (korda) yang aktif menyelenggarakan kegiatan sosial, keagamaan, olahraga, serta pertemuan rutin.'
),
(
    '01JC0000000000000000000003',
    'sejarah',
    'Sejarah Pembentukan',
    'Berawal dari pertemuan kecil antar alumni yang ingin menjaga tali silaturahmi, Tibta 65 tumbuh menjadi organisasi yang terstruktur. Pertemuan informal berkembang menjadi agenda tahunan, lalu melahirkan kepengurusan, korda, dan berbagai program kerja yang berjalan hingga hari ini.'
);