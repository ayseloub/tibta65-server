package pemilu

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	"github.com/lib/pq"
	"github.com/oklog/ulid/v2"
)

var ErrNotFound = errors.New("kandidat not found")
var ErrAlreadyVoted = errors.New("kamu sudah memberikan suara sebelumnya")
var ErrPemiluNotActive = errors.New("pemilu sedang tidak aktif")
var ErrKandidatNotFound = errors.New("kandidat tidak ditemukan")
var ErrKandidatHasVotes = errors.New("kandidat sudah memiliki suara, tidak bisa dihapus satu-satu — gunakan Reset Pemilu untuk memulai periode baru")

type Repository interface {
	FindSettings(ctx context.Context) (*Settings, error)
	UpdateSettings(ctx context.Context, startAt, endAt string) (*Settings, error)
	CloseEarly(ctx context.Context) (*Settings, error)

	FindAllKandidat(ctx context.Context) ([]Kandidat, error)
	FindKandidatByID(ctx context.Context, id string) (*Kandidat, error)
	CreateKandidat(ctx context.Context, k *Kandidat) error
	UpdateKandidat(ctx context.Context, k *Kandidat) error
	DeleteKandidat(ctx context.Context, id string) error

	CountMembers(ctx context.Context) (int, error)
	CountVotes(ctx context.Context) (int, error)

	FindMemberVote(ctx context.Context, memberID string) (*string, error)
	CreateVote(ctx context.Context, memberID, kandidatID string) error
	ResetAll(ctx context.Context) error
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) FindSettings(ctx context.Context) (*Settings, error) {
	var s Settings
	query := `SELECT id, start_at, end_at, closed_early_at, created_at, updated_at FROM pemilu_settings LIMIT 1`
	if err := r.db.GetContext(ctx, &s, query); err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *repository) UpdateSettings(ctx context.Context, startAt, endAt string) (*Settings, error) {
	var s Settings
	query := `
		UPDATE pemilu_settings
		SET start_at = $1, end_at = $2, closed_early_at = NULL, updated_at = now()
		RETURNING id, start_at, end_at, closed_early_at, created_at, updated_at
	`
	if err := r.db.GetContext(ctx, &s, query, startAt, endAt); err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *repository) CloseEarly(ctx context.Context) (*Settings, error) {
	var s Settings
	query := `
		UPDATE pemilu_settings
		SET closed_early_at = now(), updated_at = now()
		RETURNING id, start_at, end_at, closed_early_at, created_at, updated_at
	`
	if err := r.db.GetContext(ctx, &s, query); err != nil {
		return nil, err
	}
	return &s, nil
}

const kandidatSelect = `
	SELECT k.id, k.full_name, k.visi, k.misi, k.pangkat, k.created_at, k.updated_at,
	       COALESCE(COUNT(v.id), 0) AS vote_count
	FROM kandidats k
	LEFT JOIN votes v ON v.kandidat_id = k.id
`

func (r *repository) FindAllKandidat(ctx context.Context) ([]Kandidat, error) {
	kandidats := []Kandidat{}
	query := kandidatSelect + " GROUP BY k.id ORDER BY vote_count DESC, k.created_at ASC"
	err := r.db.SelectContext(ctx, &kandidats, query)
	return kandidats, err
}

func (r *repository) FindKandidatByID(ctx context.Context, id string) (*Kandidat, error) {
	var k Kandidat
	query := kandidatSelect + " WHERE k.id = $1 GROUP BY k.id"
	err := r.db.GetContext(ctx, &k, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &k, nil
}

func (r *repository) CreateKandidat(ctx context.Context, k *Kandidat) error {
	query := `
		INSERT INTO kandidats (id, full_name, visi, misi, pangkat)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at, updated_at
	`
	return r.db.QueryRowContext(ctx, query, k.ID, k.FullName, k.Visi, k.Misi, k.Pangkat).
		Scan(&k.CreatedAt, &k.UpdatedAt)
}

func (r *repository) UpdateKandidat(ctx context.Context, k *Kandidat) error {
	query := `
		UPDATE kandidats
		SET full_name = $1, visi = $2, misi = $3, pangkat = $4, updated_at = now()
		WHERE id = $5
		RETURNING updated_at
	`
	err := r.db.QueryRowContext(ctx, query, k.FullName, k.Visi, k.Misi, k.Pangkat, k.ID).Scan(&k.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func (r *repository) DeleteKandidat(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM kandidats WHERE id = $1", id)
	if err != nil {
		if isForeignKeyError(err) {
			return ErrKandidatHasVotes
		}
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *repository) CountMembers(ctx context.Context) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM members")
	return count, err
}

func (r *repository) CountVotes(ctx context.Context) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM votes")
	return count, err
}

func (r *repository) FindMemberVote(ctx context.Context, memberID string) (*string, error) {
	var kandidatID string
	err := r.db.GetContext(ctx, &kandidatID, "SELECT kandidat_id FROM votes WHERE member_id = $1", memberID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &kandidatID, nil
}

func (r *repository) CreateVote(ctx context.Context, memberID, kandidatID string) error {
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO votes (id, member_id, kandidat_id) VALUES ($1, $2, $3)",
		ulid.Make().String(), memberID, kandidatID)
	if isDuplicateKeyError(err) {
		return ErrAlreadyVoted
	}
	if isForeignKeyError(err) {
		return ErrKandidatNotFound
	}
	return err
}

func isDuplicateKeyError(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505"
}

func isForeignKeyError(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23503"
}

func (r *repository) ResetAll(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, "TRUNCATE votes, kandidats CASCADE")
	return err
}
