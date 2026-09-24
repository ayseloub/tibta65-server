package gpkegiatan

import (
	"context"
	"database/sql"
	"errors"
	"strconv"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

var (
	ErrNotFound  = errors.New("kegiatan GP TIBTA not found")
	ErrDuplicate = errors.New("judul kegiatan menghasilkan slug yang sama, coba judul lain")
)

type ListFilter struct {
	KordaID    string
	KategoriID string
	Scheduled  bool
	Page       int
	Limit      int
}

type Repository interface {
	FindAll(ctx context.Context, f ListFilter) ([]GPKegiatan, int, error)
	FindBySlug(ctx context.Context, slug string) (*GPKegiatan, error)
	Create(ctx context.Context, k *GPKegiatan) error
	Update(ctx context.Context, k *GPKegiatan) error
	Delete(ctx context.Context, slug string) error
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

const baseSelect = `
	SELECT k.id, k.slug, k.title, k.date, k.korda_id, ko.name AS korda_name,
	       k.kategori_id, kt.name AS kategori_name,
	       k.location, k.image_url, k.description, k.publish_at, k.expire_at, k.created_at, k.updated_at
	FROM gp_kegiatans k
	JOIN kordas ko ON ko.id = k.korda_id
	LEFT JOIN gp_kategoris kt ON kt.id = k.kategori_id
`

const scheduleFilter = " AND (k.publish_at IS NULL OR k.publish_at <= now()) AND (k.expire_at IS NULL OR k.expire_at > now())"

func (r *repository) FindAll(ctx context.Context, f ListFilter) ([]GPKegiatan, int, error) {
	where := " WHERE 1=1"
	args := []interface{}{}
	argPos := 1

	if f.KordaID != "" {
		where += " AND k.korda_id = $" + strconv.Itoa(argPos)
		args = append(args, f.KordaID)
		argPos++
	}
	if f.KategoriID != "" {
		where += " AND k.kategori_id = $" + strconv.Itoa(argPos)
		args = append(args, f.KategoriID)
		argPos++
	}
	if f.Scheduled {
		where += scheduleFilter
	}

	var total int
	countQuery := "SELECT COUNT(*) FROM gp_kegiatans k" + where
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	limit := f.Limit
	offset := (f.Page - 1) * limit
	listQuery := baseSelect + where + " ORDER BY k.date DESC, k.created_at DESC LIMIT $" + strconv.Itoa(argPos) + " OFFSET $" + strconv.Itoa(argPos+1)
	args = append(args, limit, offset)

	items := []GPKegiatan{}
	if err := r.db.SelectContext(ctx, &items, listQuery, args...); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *repository) FindBySlug(ctx context.Context, slug string) (*GPKegiatan, error) {
	var k GPKegiatan
	err := r.db.GetContext(ctx, &k, baseSelect+" WHERE k.slug = $1", slug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &k, nil
}

func (r *repository) Create(ctx context.Context, k *GPKegiatan) error {
	query := `
		INSERT INTO gp_kegiatans (id, slug, title, date, korda_id, kategori_id, location, image_url, description, publish_at, expire_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING created_at, updated_at
	`
	err := r.db.QueryRowContext(ctx, query,
		k.ID, k.Slug, k.Title, k.Date, k.KordaID, k.KategoriID, k.Location, k.ImageURL, k.Description, k.PublishAt, k.ExpireAt,
	).Scan(&k.CreatedAt, &k.UpdatedAt)

	if isDuplicateKeyError(err) {
		return ErrDuplicate
	}
	return err
}

func (r *repository) Update(ctx context.Context, k *GPKegiatan) error {
	query := `
		UPDATE gp_kegiatans
		SET title = $1, date = $2, korda_id = $3, kategori_id = $4, location = $5, image_url = $6,
		    description = $7, publish_at = $8, expire_at = $9, updated_at = now()
		WHERE slug = $10
		RETURNING updated_at
	`
	err := r.db.QueryRowContext(ctx, query,
		k.Title, k.Date, k.KordaID, k.KategoriID, k.Location, k.ImageURL, k.Description, k.PublishAt, k.ExpireAt, k.Slug,
	).Scan(&k.UpdatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func (r *repository) Delete(ctx context.Context, slug string) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM gp_kegiatans WHERE slug = $1", slug)
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
