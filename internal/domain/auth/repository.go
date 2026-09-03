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
	UpdateLastLogin(ctx context.Context, id string) error
}

func (r *repository) UpdateLastLogin(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE admins SET last_login_at = now() WHERE id = $1", id)
	return err
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
    SELECT id, username, full_name, password_hash, role, must_change_password, last_login_at, created_at, updated_at
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
