package member

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

var (
	ErrNotFound  = errors.New("member not found")
	ErrDuplicate = errors.New("email sudah terdaftar")
)

type Repository interface {
	FindByEmail(ctx context.Context, email string) (*Member, error)
	FindByID(ctx context.Context, id string) (*Member, error)
	FindByGoogleID(ctx context.Context, googleID string) (*Member, error)
	Create(ctx context.Context, m *Member) error
	MarkEmailVerified(ctx context.Context, id string) error
	LinkGoogleID(ctx context.Context, memberID, googleID string, avatarURL *string) error
	UpdatePassword(ctx context.Context, id, passwordHash string) error
	UpdateProfile(ctx context.Context, m *Member) error
	UpdateAvatar(ctx context.Context, id string, avatarURL *string) error
}

func (r *repository) FindByGoogleID(ctx context.Context, googleID string) (*Member, error) {
	var m Member
	err := r.db.GetContext(ctx, &m, baseSelect+" WHERE m.google_id = $1", googleID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &m, nil
}

func (r *repository) LinkGoogleID(ctx context.Context, memberID, googleID string, avatarURL *string) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE members SET google_id = $1, avatar_url = COALESCE($2, avatar_url), updated_at = now() WHERE id = $3",
		googleID, avatarURL, memberID)
	return err
}

func (r *repository) MarkEmailVerified(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE members SET email_verified_at = now() WHERE id = $1", id)
	return err
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

const baseSelect = `
	SELECT m.id, m.full_name, m.email, m.phone, m.address, m.member_number, m.password_hash, m.google_id,
	       m.avatar_url, m.korda_id, k.name AS korda_name, m.must_change_password, m.email_verified_at,
	       m.created_at, m.updated_at
	FROM members m
	LEFT JOIN kordas k ON k.id = m.korda_id
`

func (r *repository) FindByEmail(ctx context.Context, email string) (*Member, error) {
	var m Member
	err := r.db.GetContext(ctx, &m, baseSelect+" WHERE m.email = $1", email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &m, nil
}

func (r *repository) FindByID(ctx context.Context, id string) (*Member, error) {
	var m Member
	err := r.db.GetContext(ctx, &m, baseSelect+" WHERE m.id = $1", id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &m, nil
}

func (r *repository) Create(ctx context.Context, m *Member) error {
	query := `
		INSERT INTO members (id, full_name, email, password_hash, google_id, avatar_url, korda_id, member_number)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'TIBTA-1965-' || LPAD(nextval('member_number_seq')::text, 4, '0'))
		RETURNING member_number, created_at, updated_at
	`
	err := r.db.QueryRowContext(ctx, query,
		m.ID, m.FullName, m.Email, m.PasswordHash, m.GoogleID, m.AvatarURL, m.KordaID,
	).Scan(&m.MemberNumber, &m.CreatedAt, &m.UpdatedAt)

	if isDuplicateKeyError(err) {
		return ErrDuplicate
	}
	return err
}

func isDuplicateKeyError(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505"
}

func (r *repository) UpdatePassword(ctx context.Context, id, passwordHash string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE members SET password_hash = $1, updated_at = now() WHERE id = $2", passwordHash, id)
	return err
}

func (r *repository) UpdateProfile(ctx context.Context, m *Member) error {
	query := `
		UPDATE members
		SET full_name = $1, phone = $2, korda_id = $3, address = $4, updated_at = now()
		WHERE id = $5
		RETURNING updated_at
	`
	err := r.db.QueryRowContext(ctx, query, m.FullName, m.Phone, m.KordaID, m.Address, m.ID).Scan(&m.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func (r *repository) UpdateAvatar(ctx context.Context, id string, avatarURL *string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE members SET avatar_url = $1, updated_at = now() WHERE id = $2", avatarURL, id)
	return err
}
