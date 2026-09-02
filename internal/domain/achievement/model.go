package achievement

import "time"

type Achievement struct {
	ID          string    `db:"id" json:"id"`
	Year        string    `db:"year" json:"year"`
	Title       string    `db:"title" json:"title"`
	Description string    `db:"description" json:"description"`
	ImageURL    string    `db:"image_url" json:"image_url"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}
