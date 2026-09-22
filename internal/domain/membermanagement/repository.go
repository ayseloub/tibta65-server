package membermanagement

import (
	"context"
	"errors"

	"github.com/jmoiron/sqlx"

	"github.com/Tibta65web/tibta65-server/internal/domain/member"
)

var ErrNotFound = errors.New("member not found")

const baseSelect = `
	SELECT m.id, m.full_name, m.email, m.phone, m.address, m.legacy_member_number, m.member_number,
	       m.username, m.generation, m.parent_member_id, m.status, m.nama_suci, m.agama, m.nrp, m.no_ak,
	       m.pangkat_terakhir, m.legacy_identifier_raw, m.password_hash, m.google_id, m.avatar_url,
	       m.korda_id, k.name AS korda_name, m.must_change_password, m.email_verified_at, m.approved_at,
	       m.profile_completed, m.created_at, m.updated_at
	FROM members m
	LEFT JOIN kordas k ON k.id = m.korda_id
`

type Repository interface {
	FindPending(ctx context.Context, page, limit int) ([]member.Member, int, error)
	FindAll(ctx context.Context, page, limit int) ([]member.Member, int, error)
	CountPending(ctx context.Context) (int, error)
	Approve(ctx context.Context, id string) error
	UpdateStatus(ctx context.Context, id, status string) error
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
	if err := r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM members WHERE status = $1", member.StatusPendingReview); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	members := []member.Member{}
	query := baseSelect + " WHERE m.status = $1 ORDER BY m.created_at ASC LIMIT $2 OFFSET $3"
	if err := r.db.SelectContext(ctx, &members, query, member.StatusPendingReview, limit, offset); err != nil {
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
	err := r.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM members WHERE status = $1", member.StatusPendingReview)
	return count, err
}

func (r *repository) Approve(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx,
		"UPDATE members SET status = $1, approved_at = now(), updated_at = now() WHERE id = $2 AND status = $3",
		member.StatusActive, id, member.StatusPendingReview,
	)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *repository) UpdateStatus(ctx context.Context, id, status string) error {
	result, err := r.db.ExecContext(ctx, "UPDATE members SET status = $1, updated_at = now() WHERE id = $2", status, id)
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
