package korda

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

var (
	ErrNotFound  = errors.New("korda not found")
	ErrDuplicate = errors.New("nama korda sudah dipakai")
	ErrInUse     = errors.New("korda masih dipakai oleh kegiatan lain")
)

type Repository interface {
	FindAll(ctx context.Context) ([]Korda, error)
	FindByID(ctx context.Context, id string) (*Korda, error)
	Create(ctx context.Context, k *Korda) error
	Update(ctx context.Context, k *Korda) error
	Delete(ctx context.Context, id string) error
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) FindAll(ctx context.Context) ([]Korda, error) {
	kordas := []Korda{}
	query := `SELECT id, name, created_at, updated_at FROM kordas ORDER BY name ASC`
	err := r.db.SelectContext(ctx, &kordas, query)
	return kordas, err
}

func (r *repository) FindByID(ctx context.Context, id string) (*Korda, error) {
	var k Korda
	query := `SELECT id, name, created_at, updated_at FROM kordas WHERE id = $1`
	err := r.db.GetContext(ctx, &k, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &k, nil
}

func (r *repository) Create(ctx context.Context, k *Korda) error {
	query := `INSERT INTO kordas (id, name) VALUES ($1, $2) RETURNING created_at, updated_at`
	err := r.db.QueryRowContext(ctx, query, k.ID, k.Name).Scan(&k.CreatedAt, &k.UpdatedAt)
	if isDuplicateKeyError(err) {
		return ErrDuplicate
	}
	return err
}

func (r *repository) Update(ctx context.Context, k *Korda) error {
	query := `UPDATE kordas SET name = $1, updated_at = now() WHERE id = $2 RETURNING updated_at`
	err := r.db.QueryRowContext(ctx, query, k.Name, k.ID).Scan(&k.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if isDuplicateKeyError(err) {
		return ErrDuplicate
	}
	return err
}

func (r *repository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM kordas WHERE id = $1", id)
	if isForeignKeyError(err) {
		return ErrInUse
	}
	return err
}

func isDuplicateKeyError(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505"
}

func isForeignKeyError(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23503"
}
