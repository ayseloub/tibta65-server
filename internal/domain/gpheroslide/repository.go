package gpheroslide

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
)

var ErrNotFound = errors.New("hero slide not found")

type Repository interface {
	FindAllAdmin(ctx context.Context) ([]GPHeroSlide, error)
	FindAllPublic(ctx context.Context) ([]GPHeroSlide, error)
	FindByID(ctx context.Context, id string) (*GPHeroSlide, error)
	Create(ctx context.Context, s *GPHeroSlide) error
	Update(ctx context.Context, s *GPHeroSlide) error
	Delete(ctx context.Context, id string) error
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
	secondary_button_text, secondary_button_link, image_url, position, status, created_at, updated_at
`

func (r *repository) FindAllAdmin(ctx context.Context) ([]GPHeroSlide, error) {
	items := []GPHeroSlide{}
	err := r.db.SelectContext(ctx, &items, "SELECT "+selectColumns+" FROM gp_hero_slides ORDER BY position ASC")
	return items, err
}

func (r *repository) FindAllPublic(ctx context.Context) ([]GPHeroSlide, error) {
	items := []GPHeroSlide{}
	query := "SELECT " + selectColumns + " FROM gp_hero_slides WHERE status = 'published' ORDER BY position ASC"
	err := r.db.SelectContext(ctx, &items, query)
	return items, err
}

func (r *repository) FindByID(ctx context.Context, id string) (*GPHeroSlide, error) {
	var s GPHeroSlide
	err := r.db.GetContext(ctx, &s, "SELECT "+selectColumns+" FROM gp_hero_slides WHERE id = $1", id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &s, nil
}

func (r *repository) Create(ctx context.Context, s *GPHeroSlide) error {
	query := `
		INSERT INTO gp_hero_slides (id, headline, description, primary_button_text, primary_button_link,
			secondary_button_text, secondary_button_link, image_url, position, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING created_at, updated_at
	`
	return r.db.QueryRowContext(ctx, query,
		s.ID, s.Headline, s.Description, s.PrimaryButtonText, s.PrimaryButtonLink,
		s.SecondaryButtonText, s.SecondaryButtonLink, s.ImageURL, s.Position, s.Status,
	).Scan(&s.CreatedAt, &s.UpdatedAt)
}

func (r *repository) Update(ctx context.Context, s *GPHeroSlide) error {
	query := `
		UPDATE gp_hero_slides
		SET headline = $1, description = $2, primary_button_text = $3, primary_button_link = $4,
		    secondary_button_text = $5, secondary_button_link = $6, image_url = $7, status = $8, updated_at = now()
		WHERE id = $9
		RETURNING updated_at
	`
	err := r.db.QueryRowContext(ctx, query,
		s.Headline, s.Description, s.PrimaryButtonText, s.PrimaryButtonLink,
		s.SecondaryButtonText, s.SecondaryButtonLink, s.ImageURL, s.Status, s.ID,
	).Scan(&s.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func (r *repository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM gp_hero_slides WHERE id = $1", id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *repository) Reorder(ctx context.Context, orderedIDs []string) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for i, id := range orderedIDs {
		if _, err := tx.ExecContext(ctx, "UPDATE gp_hero_slides SET position = $1, updated_at = now() WHERE id = $2", i, id); err != nil {
			return err
		}
	}

	return tx.Commit()
}
