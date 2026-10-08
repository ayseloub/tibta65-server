package membermanagement

import (
	"context"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/Tibta65web/tibta65-server/internal/domain/member"
)

var (
	ErrNotFound      = errors.New("member not found")
	ErrHasActiveVote = errors.New("anggota ini sudah memberikan suara. Batalkan suaranya dulu di tab Suara Dibatalkan pada halaman Pemilu")
	ErrHasChildren   = errors.New("anggota ini punya keturunan yang terdaftar. Tinjau atau hapus akun keturunannya terlebih dahulu")
)

const baseSelect = `
	SELECT m.id, m.full_name, m.email, m.phone, m.address, m.legacy_member_number, m.member_number,
	       m.username, m.generation, m.parent_member_id, m.status, m.nama_suci, m.agama, m.nrp, m.no_ak,
	       m.pangkat_terakhir, m.legacy_identifier_raw, m.password_hash, m.google_id, m.avatar_url,
	       m.korda_id, k.name AS korda_name, m.must_change_password, m.email_verified_at, m.approved_at,
	       m.profile_completed, m.rejection_type, m.rejection_reason, m.rejected_at, m.created_at, m.updated_at
	FROM members m
	LEFT JOIN kordas k ON k.id = m.korda_id
`

type ListFilter struct {
	IDs        []string
	Search     string
	KordaID    string
	Generation int
	Status     string
	Page       int
	Limit      int
}

type Repository interface {
	FindPending(ctx context.Context, f ListFilter) ([]member.Member, int, error)
	FindAll(ctx context.Context, f ListFilter) ([]member.Member, int, error)
	FindForExport(ctx context.Context, f ListFilter) ([]member.Member, error)
	CountPending(ctx context.Context) (int, error)
	Approve(ctx context.Context, id string) error
	Reject(ctx context.Context, id, rejectionType, reason string) error
	Reopen(ctx context.Context, id, reason string) error
	ApproveRejected(ctx context.Context, id string) error
	UpdateStatus(ctx context.Context, id, status string) error
	Delete(ctx context.Context, id string) error
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func buildWhere(f ListFilter) (string, []interface{}) {
	where := " WHERE 1=1"
	args := []interface{}{}
	n := 1

	if f.Search != "" {
		where += fmt.Sprintf(
			" AND (m.full_name ILIKE $%d OR m.email ILIKE $%d OR m.member_number ILIKE $%d OR m.username ILIKE $%d OR m.nama_suci ILIKE $%d)",
			n, n, n, n, n,
		)
		args = append(args, "%"+f.Search+"%")
		n++
	}
	if f.KordaID != "" {
		where += fmt.Sprintf(" AND m.korda_id = $%d", n)
		args = append(args, f.KordaID)
		n++
	}
	if f.Generation > 0 {
		where += fmt.Sprintf(" AND m.generation = $%d", n)
		args = append(args, f.Generation)
		n++
	}
	if f.Status != "" {
		where += fmt.Sprintf(" AND m.status = $%d", n)
		args = append(args, f.Status)
		n++
	}
	if len(f.IDs) > 0 {
		where += fmt.Sprintf(" AND m.id = ANY($%d)", n)
		args = append(args, pq.Array(f.IDs))
		n++
	}

	return where, args
}

func (r *repository) findMembers(ctx context.Context, f ListFilter, orderBy string) ([]member.Member, int, error) {
	where, args := buildWhere(f)

	var total int
	if err := r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM members m"+where, args...); err != nil {
		return nil, 0, err
	}

	offset := (f.Page - 1) * f.Limit
	args = append(args, f.Limit, offset)
	query := baseSelect + where + fmt.Sprintf(" ORDER BY %s LIMIT $%d OFFSET $%d", orderBy, len(args)-1, len(args))

	members := []member.Member{}
	if err := r.db.SelectContext(ctx, &members, query, args...); err != nil {
		return nil, 0, err
	}
	return members, total, nil
}

func (r *repository) FindPending(ctx context.Context, f ListFilter) ([]member.Member, int, error) {
	f.Status = member.StatusPendingReview
	return r.findMembers(ctx, f, "m.created_at ASC")
}

func (r *repository) FindAll(ctx context.Context, f ListFilter) ([]member.Member, int, error) {
	return r.findMembers(ctx, f, "m.created_at DESC")
}

const maxExportRows = 5000

func (r *repository) FindForExport(ctx context.Context, f ListFilter) ([]member.Member, error) {
	where, args := buildWhere(f)
	query := baseSelect + where + fmt.Sprintf(" ORDER BY m.generation ASC, m.member_number ASC LIMIT %d", maxExportRows)

	members := []member.Member{}
	if err := r.db.SelectContext(ctx, &members, query, args...); err != nil {
		return nil, err
	}
	return members, nil
}

func (r *repository) CountPending(ctx context.Context) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM members WHERE status = $1", member.StatusPendingReview)
	return count, err
}

func (r *repository) execOne(ctx context.Context, query string, args ...interface{}) error {
	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *repository) Approve(ctx context.Context, id string) error {
	return r.execOne(ctx,
		"UPDATE members SET status = $1, approved_at = now(), updated_at = now() WHERE id = $2 AND status = $3",
		member.StatusActive, id, member.StatusPendingReview,
	)
}

func (r *repository) Reject(ctx context.Context, id, rejectionType, reason string) error {
	return r.execOne(ctx, `
		UPDATE members
		SET status = $1, rejection_type = $2, rejection_reason = $3, rejected_at = now(), updated_at = now()
		WHERE id = $4 AND status = $5`,
		member.StatusRejected, rejectionType, reason, id, member.StatusPendingReview,
	)
}

func (r *repository) Reopen(ctx context.Context, id, reason string) error {
	return r.execOne(ctx, `
		UPDATE members
		SET rejection_type = $1, rejection_reason = $2, updated_at = now()
		WHERE id = $3 AND status = $4`,
		member.RejectionIncomplete, reason, id, member.StatusRejected,
	)
}

func (r *repository) ApproveRejected(ctx context.Context, id string) error {
	return r.execOne(ctx, `
		UPDATE members
		SET status = $1, approved_at = now(), rejection_type = NULL, rejection_reason = NULL,
		    rejected_at = NULL, updated_at = now()
		WHERE id = $2 AND status = $3`,
		member.StatusActive, id, member.StatusRejected,
	)
}

func (r *repository) UpdateStatus(ctx context.Context, id, status string) error {
	return r.execOne(ctx, "UPDATE members SET status = $1, updated_at = now() WHERE id = $2", status, id)
}

func (r *repository) Delete(ctx context.Context, id string) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var exists int
	if err := tx.GetContext(ctx, &exists, "SELECT COUNT(*) FROM members WHERE id = $1", id); err != nil {
		return err
	}
	if exists == 0 {
		return ErrNotFound
	}

	var activeVotes int
	if err := tx.GetContext(ctx, &activeVotes, "SELECT COUNT(*) FROM votes WHERE member_id = $1 AND voided_at IS NULL", id); err != nil {
		return err
	}
	if activeVotes > 0 {
		return ErrHasActiveVote
	}

	var children int
	if err := tx.GetContext(ctx, &children, "SELECT COUNT(*) FROM members WHERE parent_member_id = $1", id); err != nil {
		return err
	}
	if children > 0 {
		return ErrHasChildren
	}

	if _, err := tx.ExecContext(ctx, "DELETE FROM votes WHERE member_id = $1 AND voided_at IS NOT NULL", id); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM member_activation_tokens WHERE member_id = $1", id); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM member_pending_changes WHERE member_id = $1", id); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, "UPDATE tickets SET reported_member_id = NULL WHERE reported_member_id = $1", id); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, "DELETE FROM members WHERE id = $1", id); err != nil {
		return err
	}
	return tx.Commit()
}
