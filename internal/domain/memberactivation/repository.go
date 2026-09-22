package memberactivation

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/oklog/ulid/v2"
)

var ErrNotFound = errors.New("token not found")

type Repository interface {
	Create(ctx context.Context, t *ActivationToken) error
	FindLatestValidByMemberID(ctx context.Context, memberID, purpose string) (*ActivationToken, error)
	MarkUsed(ctx context.Context, id string) error
	RecordFailedAttempt(ctx context.Context, memberNumber, purpose string) error
	CountRecentFailedAttempts(ctx context.Context, memberNumber, purpose string, since time.Time) (int, error)
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, t *ActivationToken) error {
	query := `
		INSERT INTO member_activation_tokens (id, member_id, token_hash, purpose, expires_at, created_by_admin_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at
	`
	return r.db.QueryRowContext(ctx, query, t.ID, t.MemberID, t.TokenHash, t.Purpose, t.ExpiresAt, t.CreatedByAdminID).
		Scan(&t.CreatedAt)
}

func (r *repository) FindLatestValidByMemberID(ctx context.Context, memberID, purpose string) (*ActivationToken, error) {
	var t ActivationToken
	query := `
		SELECT id, member_id, token_hash, purpose, expires_at, created_by_admin_id, used_at, created_at
		FROM member_activation_tokens
		WHERE member_id = $1 AND purpose = $2 AND used_at IS NULL AND expires_at > now()
		ORDER BY created_at DESC
		LIMIT 1
	`
	err := r.db.GetContext(ctx, &t, query, memberID, purpose)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &t, nil
}

func (r *repository) MarkUsed(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE member_activation_tokens SET used_at = now() WHERE id = $1", id)
	return err
}

func (r *repository) RecordFailedAttempt(ctx context.Context, memberNumber, purpose string) error {
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO activation_attempts (id, member_number, purpose) VALUES ($1, $2, $3)",
		ulid.Make().String(), memberNumber, purpose,
	)
	return err
}

func (r *repository) CountRecentFailedAttempts(ctx context.Context, memberNumber, purpose string, since time.Time) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count,
		"SELECT COUNT(*) FROM activation_attempts WHERE member_number = $1 AND purpose = $2 AND created_at > $3",
		memberNumber, purpose, since,
	)
	return count, err
}
