package heroslide

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
)

var ErrNotFound = errors.New("hero slide not found")

type Repository interface {
	FindAllAdmin(ctx context.Context) ([]HeroSlide, error)
	FindAllPublic(ctx context.Context) ([]HeroSlide, error)
	FindByID(ctx context.Context, id string) (*HeroSlide, error)
	Create(ctx context.Context, s *HeroSlide) error
	Update(ctx context.Context, s *HeroSlide) error
	Delete(ctx context.Context, id string) error
	MaxPosition(ctx context.Context) (int, error)
	Reorder(ctx context.Context, orderedIDs []string) error
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

const selectColumns = `
	id, headline, description, primary_button_text, primary_button_link,
	secondary_button_text, secondary_button_link, image_url, bg_color, position,
	status, publish_at, expire_at, created_at, updated_at
`

func (r *repository) FindAllAdmin(ctx context.Context) ([]HeroSlide, error) {
	items := []HeroSlide{}
	query := "SELECT " + selectColumns + " FROM hero_slides ORDER BY position ASC"
	err := r.db.SelectContext(ctx, &items, query)
	return items, err
}

func (r *repository) FindAllPublic(ctx context.Context) ([]HeroSlide, error) {
	items := []HeroSlide{}
	query := "SELECT " + selectColumns + ` FROM hero_slides
		WHERE status = 'published'
		AND (publish_at IS NULL OR publish_at <= now())
		AND (expire_at IS NULL OR expire_at > now())
		ORDER BY position ASC`
	err := r.db.SelectContext(ctx, &items, query)
	return items, err
}

func (r *repository) FindByID(ctx context.Context, id string) (*HeroSlide, error) {
	var s HeroSlide
	query := "SELECT " + selectColumns + " FROM hero_slides WHERE id = $1"
	if err := r.db.GetContext(ctx, &s, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &s, nil
}

func (r *repository) Create(ctx context.Context, s *HeroSlide) error {
	query := `
		INSERT INTO hero_slides (id, headline, description, primary_button_text, primary_button_link,
			secondary_button_text, secondary_button_link, image_url, bg_color, position, status, publish_at, expire_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING created_at, updated_at
	`
	return r.db.QueryRowContext(ctx, query,
		s.ID, s.Headline, s.Description, s.PrimaryButtonText, s.PrimaryButtonLink,
		s.SecondaryButtonText, s.SecondaryButtonLink, s.ImageURL, s.BgColor, s.Position,
		s.Status, s.PublishAt, s.ExpireAt,
	).Scan(&s.CreatedAt, &s.UpdatedAt)
}

func (r *repository) Update(ctx context.Context, s *HeroSlide) error {
	query := `
		UPDATE hero_slides
		SET headline = $1, description = $2, primary_button_text = $3, primary_button_link = $4,
		    secondary_button_text = $5, secondary_button_link = $6, image_url = $7, bg_color = $8,
		    status = $9, publish_at = $10, expire_at = $11, updated_at = now()
		WHERE id = $12
		RETURNING updated_at
	`
	err := r.db.QueryRowContext(ctx, query,
		s.Headline, s.Description, s.PrimaryButtonText, s.PrimaryButtonLink,
		s.SecondaryButtonText, s.SecondaryButtonLink, s.ImageURL, s.BgColor,
		s.Status, s.PublishAt, s.ExpireAt, s.ID,
	).Scan(&s.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func (r *repository) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM hero_slides WHERE id = $1", id)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *repository) MaxPosition(ctx context.Context) (int, error) {
	var max sql.NullInt64
	err := r.db.GetContext(ctx, &max, "SELECT MAX(position) FROM hero_slides")
	if err != nil {
		return 0, err
	}
	if !max.Valid {
		return -1, nil
	}
	return int(max.Int64), nil
}

func (r *repository) Reorder(ctx context.Context, orderedIDs []string) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for i, id := range orderedIDs {
		if _, err := tx.ExecContext(ctx, "UPDATE hero_slides SET position = $1, updated_at = now() WHERE id = $2", i, id); err != nil {
			return err
		}
	}

	return tx.Commit()
}
