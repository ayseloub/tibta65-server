package adminmanagement

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/Tibta65web/tibta65-server/internal/domain/auth"
)

var (
	ErrNotFound  = errors.New("admin not found")
	ErrDuplicate = errors.New("username sudah dipakai")
)

type Repository interface {
	FindAll(ctx context.Context, page, limit int) ([]auth.Admin, int, error)
	Create(ctx context.Context, a *auth.Admin) error
	Update(ctx context.Context, a *auth.Admin) error
	UpdatePassword(ctx context.Context, id, passwordHash string) error
	Delete(ctx context.Context, id string) error
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) FindAll(ctx context.Context, page, limit int) ([]auth.Admin, int, error) {
	var total int
	if err := r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM admins"); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	admins := []auth.Admin{}
	query := `
		SELECT id, username, full_name, password_hash, role, must_change_password, last_login_at, created_at, updated_at
		FROM admins
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	if err := r.db.SelectContext(ctx, &admins, query, limit, offset); err != nil {
		return nil, 0, err
	}
	return admins, total, nil
}

func (r *repository) Create(ctx context.Context, a *auth.Admin) error {
	query := `
		INSERT INTO admins (id, username, full_name, password_hash, role, must_change_password)
		VALUES ($1, $2, $3, $4, $5, true)
		RETURNING created_at, updated_at
	`
	err := r.db.QueryRowContext(ctx, query, a.ID, a.Username, a.FullName, a.PasswordHash, a.Role).
		Scan(&a.CreatedAt, &a.UpdatedAt)
	if isDuplicateKeyError(err) {
		return ErrDuplicate
	}
	return err
}

func (r *repository) Update(ctx context.Context, a *auth.Admin) error {
	query := `
		UPDATE admins
		SET username = $1, full_name = $2, role = $3, updated_at = now()
		WHERE id = $4
		RETURNING updated_at
	`
	err := r.db.QueryRowContext(ctx, query, a.Username, a.FullName, a.Role, a.ID).Scan(&a.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if isDuplicateKeyError(err) {
		return ErrDuplicate
	}
	return err
}

func (r *repository) UpdatePassword(ctx context.Context, id, passwordHash string) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE admins SET password_hash = $1, must_change_password = false, updated_at = now() WHERE id = $2",
		passwordHash, id)
	return err
}

func (r *repository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM admins WHERE id = $1", id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func isDuplicateKeyError(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505"
}
