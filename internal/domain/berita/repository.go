package berita

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

var ErrNotFound = errors.New("berita not found")

type ListFilter struct {
	Search     string
	Status     string
	Visibility string
	Page       int
	Limit      int
}

type Repository interface {
	Create(ctx context.Context, b *Berita) error
	Update(ctx context.Context, b *Berita) error
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*Berita, error)
	FindBySlug(ctx context.Context, slug string) (*Berita, error)
	FindAll(ctx context.Context, f ListFilter) ([]Berita, int, error)
	CountAll(ctx context.Context) (int, error)
	CountByStatus(ctx context.Context, status string) (int, error)
	SlugExists(ctx context.Context, slug string) (bool, error)

	SetHighlight(ctx context.Context, id string) error
	UnsetHighlight(ctx context.Context, id string) error
	FindHighlightPublic(ctx context.Context) (*Berita, error)
	FindAllPublic(ctx context.Context, page, limit int) ([]Berita, int, error)
	FindBySlugPublic(ctx context.Context, slug string) (*Berita, error)

	FindAllMember(ctx context.Context, page, limit int) ([]Berita, int, error)
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

const selectColumns = `
	id, title, slug, description, image_url, visibility, status, is_highlight,
	author_id, author_name, event_date, publish_at, expire_at, created_at, updated_at
`

func (r *repository) Create(ctx context.Context, b *Berita) error {
	query := `
		INSERT INTO beritas (id, title, slug, description, image_url, visibility, status, author_id, author_name, event_date, publish_at, expire_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING created_at, updated_at
	`
	return r.db.QueryRowContext(ctx, query,
		b.ID, b.Title, b.Slug, b.Description, b.ImageURL, b.Visibility, b.Status, b.AuthorID, b.AuthorName, b.EventDate, b.PublishAt, b.ExpireAt,
	).Scan(&b.CreatedAt, &b.UpdatedAt)
}

func (r *repository) Update(ctx context.Context, b *Berita) error {
	query := `
		UPDATE beritas
		SET title = $1, slug = $2, description = $3, image_url = $4, visibility = $5,
		    status = $6, event_date = $7, publish_at = $8, expire_at = $9, updated_at = now()
		WHERE id = $10
		RETURNING updated_at
	`
	err := r.db.QueryRowContext(ctx, query,
		b.Title, b.Slug, b.Description, b.ImageURL, b.Visibility, b.Status, b.EventDate, b.PublishAt, b.ExpireAt, b.ID,
	).Scan(&b.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func (r *repository) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM beritas WHERE id = $1", id)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *repository) FindByID(ctx context.Context, id string) (*Berita, error) {
	var b Berita
	query := "SELECT " + selectColumns + " FROM beritas WHERE id = $1"
	if err := r.db.GetContext(ctx, &b, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &b, nil
}

func (r *repository) FindBySlug(ctx context.Context, slug string) (*Berita, error) {
	var b Berita
	query := "SELECT " + selectColumns + " FROM beritas WHERE slug = $1"
	if err := r.db.GetContext(ctx, &b, query, slug); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &b, nil
}

func (r *repository) FindAll(ctx context.Context, f ListFilter) ([]Berita, int, error) {
	where := " WHERE 1=1"
	args := []interface{}{}
	argN := 1

	if f.Search != "" {
		where += fmt.Sprintf(" AND title ILIKE $%d", argN)
		args = append(args, "%"+f.Search+"%")
		argN++
	}
	if f.Status != "" {
		where += fmt.Sprintf(" AND status = $%d", argN)
		args = append(args, f.Status)
		argN++
	}
	if f.Visibility != "" {
		where += fmt.Sprintf(" AND visibility = $%d", argN)
		args = append(args, f.Visibility)
		argN++
	}

	var total int
	countQuery := "SELECT COUNT(*) FROM beritas" + where
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	offset := (f.Page - 1) * f.Limit
	args = append(args, f.Limit, offset)
	query := "SELECT " + selectColumns + " FROM beritas" + where +
		fmt.Sprintf(" ORDER BY event_date DESC, created_at DESC LIMIT $%d OFFSET $%d", argN, argN+1)

	items := []Berita{}
	if err := r.db.SelectContext(ctx, &items, query, args...); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *repository) CountAll(ctx context.Context) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM beritas")
	return count, err
}

func (r *repository) CountByStatus(ctx context.Context, status string) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM beritas WHERE status = $1", status)
	return count, err
}

func (r *repository) SlugExists(ctx context.Context, slug string) (bool, error) {
	var exists bool
	err := r.db.GetContext(ctx, &exists, "SELECT EXISTS(SELECT 1 FROM beritas WHERE slug = $1)", slug)
	return exists, err
}

func (r *repository) SetHighlight(ctx context.Context, id string) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx,
		"UPDATE beritas SET is_highlight = false, updated_at = now() WHERE is_highlight = true AND id != $1", id,
	); err != nil {
		return err
	}

	res, err := tx.ExecContext(ctx,
		"UPDATE beritas SET is_highlight = true, updated_at = now() WHERE id = $1", id,
	)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}

	return tx.Commit()
}

func (r *repository) UnsetHighlight(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx,
		"UPDATE beritas SET is_highlight = false, updated_at = now() WHERE id = $1", id,
	)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

const scheduleFilter = " AND (publish_at IS NULL OR publish_at <= now()) AND (expire_at IS NULL OR expire_at > now())"

func (r *repository) FindHighlightPublic(ctx context.Context) (*Berita, error) {
	var b Berita
	query := "SELECT " + selectColumns + ` FROM beritas
		WHERE is_highlight = true AND status = 'published' AND visibility = 'public'` + scheduleFilter + `
		LIMIT 1`
	if err := r.db.GetContext(ctx, &b, query); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &b, nil
}

func (r *repository) FindAllPublic(ctx context.Context, page, limit int) ([]Berita, int, error) {
	where := " WHERE status = 'published' AND visibility = 'public' AND is_highlight = false" + scheduleFilter

	var total int
	if err := r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM beritas"+where); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	items := []Berita{}
	query := "SELECT " + selectColumns + " FROM beritas" + where +
		" ORDER BY event_date DESC, created_at DESC LIMIT $1 OFFSET $2"
	if err := r.db.SelectContext(ctx, &items, query, limit, offset); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *repository) FindBySlugPublic(ctx context.Context, slug string) (*Berita, error) {
	var b Berita
	query := "SELECT " + selectColumns + ` FROM beritas
		WHERE slug = $1 AND status = 'published' AND visibility = 'public'` + scheduleFilter
	if err := r.db.GetContext(ctx, &b, query, slug); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &b, nil
}

func (r *repository) FindAllMember(ctx context.Context, page, limit int) ([]Berita, int, error) {
	where := " WHERE status = 'published' AND visibility = 'internal'" + scheduleFilter

	var total int
	if err := r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM beritas"+where); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	items := []Berita{}
	query := "SELECT " + selectColumns + " FROM beritas" + where +
		" ORDER BY event_date DESC, created_at DESC LIMIT $1 OFFSET $2"
	if err := r.db.SelectContext(ctx, &items, query, limit, offset); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}
