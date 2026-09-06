package sitesettings

import "time"

type Settings struct {
	ID           string    `db:"id" json:"id"`
	Email        *string   `db:"email" json:"email"`
	YoutubeURL   *string   `db:"youtube_url" json:"youtube_url"`
	InstagramURL *string   `db:"instagram_url" json:"instagram_url"`
	FacebookURL  *string   `db:"facebook_url" json:"facebook_url"`
	Address      *string   `db:"address" json:"address"`
	MapsEmbedURL *string   `db:"maps_embed_url" json:"maps_embed_url"`
	MapsLink     *string   `db:"maps_link" json:"maps_link"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}
