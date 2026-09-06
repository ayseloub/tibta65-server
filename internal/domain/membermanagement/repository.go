package membermanagement

import (
	"context"
	"errors"

	"github.com/jmoiron/sqlx"

	"github.com/Tibta65web/tibta65-server/internal/domain/member"
)

var ErrNotFound = errors.New("member not found")

const baseSelect = `
	SELECT m.id, m.full_name, m.email, m.phone, m.address, m.member_number, m.password_hash, m.google_id,
	       m.avatar_url, m.korda_id, k.name AS korda_name, m.must_change_password, m.email_verified_at,
	       m.approved_at, m.created_at, m.updated_at
	FROM members m
	LEFT JOIN kordas k ON k.id = m.korda_id
`

type Repository interface {
	FindPending(ctx context.Context, page, limit int) ([]member.Member, int, error)
	FindAll(ctx context.Context, page, limit int) ([]member.Member, int, error)
	CountPending(ctx context.Context) (int, error)
	Approve(ctx context.Context, id string) error
	Delete(ctx context.Context, id string) error
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) FindPending(ctx context.Context, page, limit int) ([]member.Member, int, error) {
	var total int
	if err := r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM members WHERE approved_at IS NULL"); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	members := []member.Member{}
	query := baseSelect + " WHERE m.approved_at IS NULL ORDER BY m.created_at ASC LIMIT $1 OFFSET $2"
	if err := r.db.SelectContext(ctx, &members, query, limit, offset); err != nil {
		return nil, 0, err
	}
	return members, total, nil
}

func (r *repository) FindAll(ctx context.Context, page, limit int) ([]member.Member, int, error) {
	var total int
	if err := r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM members"); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	members := []member.Member{}
	query := baseSelect + " ORDER BY m.created_at DESC LIMIT $1 OFFSET $2"
	if err := r.db.SelectContext(ctx, &members, query, limit, offset); err != nil {
		return nil, 0, err
	}
	return members, total, nil
}

func (r *repository) CountPending(ctx context.Context) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM members WHERE approved_at IS NULL")
	return count, err
}

func (r *repository) Approve(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, "UPDATE members SET approved_at = now(), updated_at = now() WHERE id = $1", id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *repository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM members WHERE id = $1", id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}
