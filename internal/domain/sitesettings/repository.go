package sitesettings

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type Repository interface {
	Find(ctx context.Context) (*Settings, error)
	Update(ctx context.Context, s *Settings) error
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

const selectQuery = `
	SELECT id, email, youtube_url, instagram_url, facebook_url, address, maps_embed_url, maps_link, created_at, updated_at
	FROM site_settings LIMIT 1
`

func (r *repository) Find(ctx context.Context) (*Settings, error) {
	var s Settings
	if err := r.db.GetContext(ctx, &s, selectQuery); err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *repository) Update(ctx context.Context, s *Settings) error {
	query := `
		UPDATE site_settings
		SET email = $1, youtube_url = $2, instagram_url = $3, facebook_url = $4,
		    address = $5, maps_embed_url = $6, maps_link = $7, updated_at = now()
		RETURNING updated_at
	`
	return r.db.QueryRowContext(ctx, query,
		s.Email, s.YoutubeURL, s.InstagramURL, s.FacebookURL, s.Address, s.MapsEmbedURL, s.MapsLink,
	).Scan(&s.UpdatedAt)
}
