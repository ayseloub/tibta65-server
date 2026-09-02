package backgroundcontent

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
)

var ErrContentNotFound = errors.New("background content not found")

type Repository interface {
	FindBySection(ctx context.Context, section string) (*BackgroundContent, error)
	UpdateBySection(ctx context.Context, section, title, description string) (*BackgroundContent, error)
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) FindBySection(ctx context.Context, section string) (*BackgroundContent, error) {
	var content BackgroundContent

	query := `
		SELECT id, section, title, description, updated_at
		FROM background_content
		WHERE section = $1
	`

	err := r.db.GetContext(ctx, &content, query, section)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrContentNotFound
		}
		return nil, err
	}

	return &content, nil
}

func (r *repository) UpdateBySection(ctx context.Context, section, title, description string) (*BackgroundContent, error) {
	query := `
		UPDATE background_content
		SET title = $1, description = $2, updated_at = now()
		WHERE section = $3
		RETURNING id, section, title, description, updated_at
	`

	var content BackgroundContent

	err := r.db.GetContext(ctx, &content, query, title, description, section)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrContentNotFound
		}
		return nil, err
	}

	return &content, nil
}
