package gallery

import (
	"time"

	"github.com/lib/pq"
)

type Album struct {
	ID                string        `db:"id" json:"id"`
	Title             string        `db:"title" json:"title"`
	Description       string        `db:"description" json:"description"`
	EventDate         time.Time     `db:"event_date" json:"event_date"`
	KordaID           *string       `db:"korda_id" json:"korda_id"`
	KordaName         *string       `db:"korda_name" json:"korda_name"`
	KategoriID        *string       `db:"kategori_id" json:"kategori_id"`
	KategoriName      *string       `db:"kategori_name" json:"kategori_name"`
	IsHighlight       bool          `db:"is_highlight" json:"is_highlight"`
	Visibility        string        `db:"visibility" json:"visibility"`
	TargetGenerations pq.Int64Array `db:"target_generations" json:"target_generations"`
	PhotoCount        int           `db:"photo_count" json:"photo_count"`
	CoverImageURL     *string       `db:"cover_image_url" json:"cover_image_url"`
	CreatedAt         time.Time     `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time     `db:"updated_at" json:"updated_at"`
	PublishAt         *time.Time    `db:"publish_at" json:"publish_at"`
	ExpireAt          *time.Time    `db:"expire_at" json:"expire_at"`
}

type Photo struct {
	ID        string    `db:"id" json:"id"`
	AlbumID   string    `db:"album_id" json:"album_id"`
	ImageURL  string    `db:"image_url" json:"image_url"`
	Caption   *string   `db:"caption" json:"caption"`
	SortOrder int       `db:"sort_order" json:"sort_order"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type AlbumDetail struct {
	Album
	Photos []Photo `json:"photos"`
}
