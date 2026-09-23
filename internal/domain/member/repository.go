package member

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

var (
	ErrNotFound              = errors.New("member not found")
	ErrDuplicate             = errors.New("email sudah terdaftar")
	ErrDuplicateUsername     = errors.New("username sudah digunakan")
	ErrDuplicateMemberNumber = errors.New("nomor induk sudah digunakan")
)

type Repository interface {
	FindByEmail(ctx context.Context, email string) (*Member, error)
	FindByID(ctx context.Context, id string) (*Member, error)
	FindByGoogleID(ctx context.Context, googleID string) (*Member, error)
	FindByMemberNumber(ctx context.Context, memberNumber string) (*Member, error)
	FindDirectChildren(ctx context.Context, memberID string) ([]Member, error)
	FindDescendants(ctx context.Context, memberID string) ([]Member, error)
	FindSiblings(ctx context.Context, memberID string) ([]Member, error)
	NextMemberNumber(ctx context.Context, generation int) (string, error)
	Create(ctx context.Context, m *Member) error
	Delete(ctx context.Context, id string) error
	MarkEmailVerified(ctx context.Context, id string) error
	LinkGoogleID(ctx context.Context, memberID, googleID string, avatarURL *string) error
	UpdatePassword(ctx context.Context, id, passwordHash string) error
	UpdateProfile(ctx context.Context, m *Member) error
	UpdateAvatar(ctx context.Context, id string, avatarURL *string) error
	ClaimAccount(ctx context.Context, m *Member) error
	FindByUsername(ctx context.Context, username string) (*Member, error)
	CompleteProfile(ctx context.Context, memberID, memberNumber string, generation int, parentMemberID *string, kordaID string) error
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

const memberColumns = `
	m.id, m.full_name, m.email, m.phone, m.address, m.legacy_member_number, m.member_number,
	m.username, m.generation, m.parent_member_id, m.status, m.nama_suci, m.agama, m.nrp, m.no_ak,
	m.pangkat_terakhir, m.legacy_identifier_raw, m.password_hash, m.google_id, m.avatar_url,
	m.korda_id, k.name AS korda_name, m.must_change_password, m.email_verified_at, m.approved_at,
	m.profile_completed, m.created_at, m.updated_at
`

const baseSelect = "SELECT " + memberColumns + " FROM members m LEFT JOIN kordas k ON k.id = m.korda_id"

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

func (r *repository) FindByUsername(ctx context.Context, username string) (*Member, error) {
	var m Member
	err := r.db.GetContext(ctx, &m, baseSelect+" WHERE m.username = $1", username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &m, nil
}

func (r *repository) CompleteProfile(ctx context.Context, memberID, memberNumber string, generation int, parentMemberID *string, kordaID string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE members SET member_number = $1, generation = $2, parent_member_id = $3, korda_id = $4, profile_completed = true, updated_at = now() WHERE id = $5`,
		memberNumber, generation, parentMemberID, kordaID, memberID,
	)
	return err
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

func (r *repository) FindByMemberNumber(ctx context.Context, memberNumber string) (*Member, error) {
	var m Member
	err := r.db.GetContext(ctx, &m, baseSelect+" WHERE m.member_number = $1", memberNumber)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &m, nil
}

func (r *repository) FindDirectChildren(ctx context.Context, memberID string) ([]Member, error) {
	items := []Member{}
	err := r.db.SelectContext(ctx, &items, baseSelect+" WHERE m.parent_member_id = $1 ORDER BY m.created_at", memberID)
	return items, err
}

func (r *repository) FindDescendants(ctx context.Context, memberID string) ([]Member, error) {
	items := []Member{}
	query := `
		WITH RECURSIVE descendants AS (
			SELECT id, parent_member_id FROM members WHERE parent_member_id = $1
			UNION ALL
			SELECT m.id, m.parent_member_id FROM members m
			JOIN descendants d ON m.parent_member_id = d.id
		)
		SELECT ` + memberColumns + `
		FROM members m
		LEFT JOIN kordas k ON k.id = m.korda_id
		JOIN descendants d ON d.id = m.id
		ORDER BY m.generation, m.created_at
	`
	err := r.db.SelectContext(ctx, &items, query, memberID)
	return items, err
}

func (r *repository) FindSiblings(ctx context.Context, memberID string) ([]Member, error) {
	items := []Member{}
	query := baseSelect + `
		WHERE m.parent_member_id = (SELECT parent_member_id FROM members WHERE id = $1)
		AND m.id != $1
		ORDER BY m.created_at
	`
	err := r.db.SelectContext(ctx, &items, query, memberID)
	return items, err
}

func (r *repository) NextMemberNumber(ctx context.Context, generation int) (string, error) {
	var maxSeq int
	query := `SELECT COALESCE(MAX(SUBSTRING(member_number FROM 8)::int), 0) FROM members WHERE generation = $1`
	if err := r.db.GetContext(ctx, &maxSeq, query, generation); err != nil {
		return "", err
	}
	return fmt.Sprintf("65-G%02d-%04d", generation, maxSeq+1), nil
}

func (r *repository) Create(ctx context.Context, m *Member) error {
	query := `
		INSERT INTO members (id, full_name, email, password_hash, google_id, avatar_url, korda_id,
			member_number, generation, parent_member_id, status, profile_completed,
			nama_suci, agama, nrp, no_ak, pangkat_terakhir)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		RETURNING created_at, updated_at
	`
	err := r.db.QueryRowContext(ctx, query,
		m.ID, m.FullName, m.Email, m.PasswordHash, m.GoogleID, m.AvatarURL, m.KordaID,
		m.MemberNumber, m.Generation, m.ParentMemberID, m.Status, m.ProfileCompleted,
		m.NamaSuci, m.Agama, m.NRP, m.NoAK, m.PangkatTerakhir,
	).Scan(&m.CreatedAt, &m.UpdatedAt)

	if err != nil {
		return mapDuplicateError(err)
	}
	return nil
}

func (r *repository) ClaimAccount(ctx context.Context, m *Member) error {
	query := `
		UPDATE members
		SET username = $1, password_hash = $2, google_id = COALESCE($3, google_id),
		    avatar_url = COALESCE($4, avatar_url), korda_id = $5, phone = $6, address = $7,
		    nama_suci = $8, agama = $9, nrp = $10, no_ak = $11, pangkat_terakhir = $12,
		    status = $13, profile_completed = true, email_verified_at = now(), updated_at = now()
		WHERE id = $14
		RETURNING updated_at
	`
	err := r.db.QueryRowContext(ctx, query,
		m.Username, m.PasswordHash, m.GoogleID, m.AvatarURL, m.KordaID, m.Phone, m.Address,
		m.NamaSuci, m.Agama, m.NRP, m.NoAK, m.PangkatTerakhir, m.Status, m.ID,
	).Scan(&m.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return mapDuplicateError(err)
	}
	return nil
}

func (r *repository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM members WHERE id = $1", id)
	return err
}

func mapDuplicateError(err error) error {
	var pqErr *pq.Error
	if !errors.As(err, &pqErr) || pqErr.Code != "23505" {
		return err
	}
	switch pqErr.Constraint {
	case "members_email_key":
		return ErrDuplicate
	case "members_username_key":
		return ErrDuplicateUsername
	case "members_member_number_key":
		return ErrDuplicateMemberNumber
	default:
		return ErrDuplicate
	}
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

func (r *repository) UpdatePassword(ctx context.Context, id, passwordHash string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE members SET password_hash = $1, updated_at = now() WHERE id = $2", passwordHash, id)
	return err
}

func (r *repository) UpdateProfile(ctx context.Context, m *Member) error {
	query := `
		UPDATE members
		SET full_name = $1, phone = $2, korda_id = $3, address = $4,
		    nama_suci = $5, agama = $6, nrp = $7, no_ak = $8, pangkat_terakhir = $9, updated_at = now()
		WHERE id = $10
		RETURNING updated_at
	`
	err := r.db.QueryRowContext(ctx, query,
		m.FullName, m.Phone, m.KordaID, m.Address,
		m.NamaSuci, m.Agama, m.NRP, m.NoAK, m.PangkatTerakhir, m.ID,
	).Scan(&m.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func (r *repository) UpdateAvatar(ctx context.Context, id string, avatarURL *string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE members SET avatar_url = $1, updated_at = now() WHERE id = $2", avatarURL, id)
	return err
}
