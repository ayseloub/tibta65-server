package kegiatan

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

var (
	ErrNotFound  = errors.New("kegiatan not found")
	ErrDuplicate = errors.New("judul kegiatan menghasilkan slug yang sama, coba judul lain")
)

type ListFilter struct {
	Search     string
	KordaID    string
	KategoriID string
	Page       int
	Limit      int
}

type Repository interface {
	FindAll(ctx context.Context, f ListFilter) ([]Kegiatan, int, error)
	FindBySlug(ctx context.Context, slug string) (*Kegiatan, error)
	Create(ctx context.Context, k *Kegiatan) error
	Update(ctx context.Context, k *Kegiatan) error
	Delete(ctx context.Context, slug string) error
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

const baseSelect = `
	SELECT
		k.id, k.slug, k.title, k.date, k.korda_id, ko.name AS korda_name,
		k.kategori_id, kt.name AS kategori_name, k.location, k.image_url,
		k.description, k.created_at, k.updated_at
	FROM kegiatans k
	JOIN kordas ko ON ko.id = k.korda_id
	JOIN kategoris kt ON kt.id = k.kategori_id
`

func (r *repository) FindAll(ctx context.Context, f ListFilter) ([]Kegiatan, int, error) {
	where := " WHERE 1=1"
	args := []interface{}{}
	argPos := 1

	if f.Search != "" {
		where += fmt.Sprintf(" AND k.title ILIKE $%d", argPos)
		args = append(args, "%"+f.Search+"%")
		argPos++
	}
	if f.KordaID != "" {
		where += fmt.Sprintf(" AND k.korda_id = $%d", argPos)
		args = append(args, f.KordaID)
		argPos++
	}
	if f.KategoriID != "" {
		where += fmt.Sprintf(" AND k.kategori_id = $%d", argPos)
		args = append(args, f.KategoriID)
		argPos++
	}

	var total int
	countQuery := "SELECT COUNT(*) FROM kegiatans k" + where
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	limit := f.Limit
	offset := (f.Page - 1) * limit

	listQuery := baseSelect + where +
		" ORDER BY k.created_at DESC LIMIT $" + strconv.Itoa(argPos) + " OFFSET $" + strconv.Itoa(argPos+1)
	args = append(args, limit, offset)

	kegiatans := []Kegiatan{}
	if err := r.db.SelectContext(ctx, &kegiatans, listQuery, args...); err != nil {
		return nil, 0, err
	}

	return kegiatans, total, nil
}

func (r *repository) FindBySlug(ctx context.Context, slug string) (*Kegiatan, error) {
	var k Kegiatan
	query := baseSelect + " WHERE k.slug = $1"
	err := r.db.GetContext(ctx, &k, query, slug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &k, nil
}

func (r *repository) Create(ctx context.Context, k *Kegiatan) error {
	query := `
		INSERT INTO kegiatans (id, slug, title, date, korda_id, kategori_id, location, image_url, description)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING created_at, updated_at
	`
	err := r.db.QueryRowContext(ctx, query,
		k.ID, k.Slug, k.Title, k.Date, k.KordaID, k.KategoriID, k.Location, k.ImageURL, k.Description,
	).Scan(&k.CreatedAt, &k.UpdatedAt)

	if isDuplicateKeyError(err) {
		return ErrDuplicate
	}
	return err
}

func (r *repository) Update(ctx context.Context, k *Kegiatan) error {
	query := `
		UPDATE kegiatans
		SET title = $1, date = $2, korda_id = $3, kategori_id = $4,
		    location = $5, image_url = $6, description = $7, updated_at = now()
		WHERE slug = $8
		RETURNING updated_at
	`
	err := r.db.QueryRowContext(ctx, query,
		k.Title, k.Date, k.KordaID, k.KategoriID, k.Location, k.ImageURL, k.Description, k.Slug,
	).Scan(&k.UpdatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func (r *repository) Delete(ctx context.Context, slug string) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM kegiatans WHERE slug = $1", slug)
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
