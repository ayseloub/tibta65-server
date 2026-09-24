package gpaboutsettings

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type Repository interface {
	Get(ctx context.Context) (*GPAboutSettings, error)
	Update(ctx context.Context, s *GPAboutSettings) error
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Get(ctx context.Context) (*GPAboutSettings, error) {
	var s GPAboutSettings
	err := r.db.GetContext(ctx, &s, "SELECT id, title, description, stat_1, stat_2, updated_at FROM gp_about_settings LIMIT 1")
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *repository) Update(ctx context.Context, s *GPAboutSettings) error {
	query := `
		UPDATE gp_about_settings
		SET title = $1, description = $2, stat_1 = $3, stat_2 = $4, updated_at = now()
		RETURNING id, updated_at
	`
	return r.db.QueryRowContext(ctx, query, s.Title, s.Description, s.Stat1, s.Stat2).Scan(&s.ID, &s.UpdatedAt)
}
