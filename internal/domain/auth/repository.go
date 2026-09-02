package auth

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
)

var ErrAdminNotFound = errors.New("admin not found")

type Repository interface {
	FindByUsername(ctx context.Context, username string) (*Admin, error)
	FindByID(ctx context.Context, id string) (*Admin, error)
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) FindByUsername(ctx context.Context, username string) (*Admin, error) {
	var admin Admin

	query := `
		SELECT id, username, password_hash, role, must_change_password, created_at, updated_at
		FROM admins
		WHERE username = $1
	`

	err := r.db.GetContext(ctx, &admin, query, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAdminNotFound
		}
		return nil, err
	}

	return &admin, nil
}

func (r *repository) FindByID(ctx context.Context, id string) (*Admin, error) {
	var admin Admin

	query := `
		SELECT id, username, password_hash, role, must_change_password, created_at, updated_at
		FROM admins
		WHERE id = $1
	`

	err := r.db.GetContext(ctx, &admin, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAdminNotFound
		}
		return nil, err
	}

	return &admin, nil
}
