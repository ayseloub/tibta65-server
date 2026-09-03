package achievement

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
)

var ErrAchievementNotFound = errors.New("achievement not found")

type Repository interface {
	FindAll(ctx context.Context, page, limit int, year string) ([]Achievement, int, error)
	FindByID(ctx context.Context, id string) (*Achievement, error)
	Create(ctx context.Context, a *Achievement) error
	Update(ctx context.Context, a *Achievement) error
	Delete(ctx context.Context, id string) error
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) FindAll(ctx context.Context, page, limit int, year string) ([]Achievement, int, error) {
	offset := (page - 1) * limit

	achievements := []Achievement{}
	var total int

	whereClause := ""
	args := []interface{}{}
	argPos := 1

	if year != "" {
		whereClause = " WHERE year = $1"
		args = append(args, year)
		argPos++
	}

	countQuery := "SELECT COUNT(*) FROM achievements" + whereClause
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	listQuery := `
		SELECT id, year, title, description, image_url, created_at, updated_at
		FROM achievements
	` + whereClause + `
		ORDER BY created_at DESC
		LIMIT $` + itoa(argPos) + ` OFFSET $` + itoa(argPos+1)

	args = append(args, limit, offset)

	if err := r.db.SelectContext(ctx, &achievements, listQuery, args...); err != nil {
		return nil, 0, err
	}

	return achievements, total, nil
}

func (r *repository) FindByID(ctx context.Context, id string) (*Achievement, error) {
	var a Achievement

	query := `
		SELECT id, year, title, description, image_url, created_at, updated_at
		FROM achievements
		WHERE id = $1
	`

	err := r.db.GetContext(ctx, &a, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAchievementNotFound
		}
		return nil, err
	}

	return &a, nil
}

func (r *repository) Create(ctx context.Context, a *Achievement) error {
	query := `
		INSERT INTO achievements (id, year, title, description, image_url)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at, updated_at
	`

	return r.db.QueryRowContext(ctx, query, a.ID, a.Year, a.Title, a.Description, a.ImageURL).
		Scan(&a.CreatedAt, &a.UpdatedAt)
}

func (r *repository) Update(ctx context.Context, a *Achievement) error {
	query := `
		UPDATE achievements
		SET year = $1, title = $2, description = $3, image_url = $4, updated_at = now()
		WHERE id = $5
		RETURNING updated_at
	`

	err := r.db.QueryRowContext(ctx, query, a.Year, a.Title, a.Description, a.ImageURL, a.ID).
		Scan(&a.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrAchievementNotFound
		}
		return err
	}

	return nil
}

func (r *repository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM achievements WHERE id = $1", id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrAchievementNotFound
	}

	return nil
}

func itoa(n int) string {
	return string(rune('0' + n))
}
