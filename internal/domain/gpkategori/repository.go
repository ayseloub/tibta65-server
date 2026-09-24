package gpkategori

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

var (
	ErrNotFound  = errors.New("kategori not found")
	ErrDuplicate = errors.New("kategori dengan nama ini sudah ada")
	ErrInUse     = errors.New("kategori masih dipakai oleh kegiatan, tidak bisa dihapus")
)

type Repository interface {
	FindAll(ctx context.Context) ([]GPKategori, error)
	Create(ctx context.Context, k *GPKategori) error
	Update(ctx context.Context, id, name string) (*GPKategori, error)
	Delete(ctx context.Context, id string) error
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) FindAll(ctx context.Context) ([]GPKategori, error) {
	items := []GPKategori{}
	err := r.db.SelectContext(ctx, &items, "SELECT id, name, created_at, updated_at FROM gp_kategoris ORDER BY name ASC")
	return items, err
}

func (r *repository) Create(ctx context.Context, k *GPKategori) error {
	query := `INSERT INTO gp_kategoris (id, name) VALUES ($1, $2) RETURNING created_at, updated_at`
	err := r.db.QueryRowContext(ctx, query, k.ID, k.Name).Scan(&k.CreatedAt, &k.UpdatedAt)
	if isDuplicateKeyError(err) {
		return ErrDuplicate
	}
	return err
}

func (r *repository) Update(ctx context.Context, id, name string) (*GPKategori, error) {
	var k GPKategori
	query := `UPDATE gp_kategoris SET name = $1, updated_at = now() WHERE id = $2 RETURNING id, name, created_at, updated_at`
	err := r.db.GetContext(ctx, &k, query, name, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		if isDuplicateKeyError(err) {
			return nil, ErrDuplicate
		}
		return nil, err
	}
	return &k, nil
}

func (r *repository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM gp_kategoris WHERE id = $1", id)
	if err != nil {
		if isForeignKeyError(err) {
			return ErrInUse
		}
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

func isForeignKeyError(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23503"
}
