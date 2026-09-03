package kegiatan

import "time"

type Kegiatan struct {
	ID           string    `db:"id" json:"id"`
	Slug         string    `db:"slug" json:"slug"`
	Title        string    `db:"title" json:"title"`
	Date         time.Time `db:"date" json:"date"`
	KordaID      string    `db:"korda_id" json:"korda_id"`
	KordaName    string    `db:"korda_name" json:"korda_name"`
	KategoriID   string    `db:"kategori_id" json:"kategori_id"`
	KategoriName string    `db:"kategori_name" json:"kategori_name"`
	Location     string    `db:"location" json:"location"`
	ImageURL     string    `db:"image_url" json:"image_url"`
	Description  string    `db:"description" json:"description"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}
