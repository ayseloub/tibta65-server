package ticket

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

var ErrNotFound = errors.New("ticket not found")

type Repository interface {
	Create(ctx context.Context, t *Ticket) error
	CountRecentByMember(ctx context.Context, memberID string, since time.Time) (int, error)
	FindAllByMember(ctx context.Context, memberID string, page, limit int) ([]Ticket, int, error)
	FindByIDForMember(ctx context.Context, id, memberID string) (*Ticket, error)
	FindAllAdmin(ctx context.Context, kordaID, status string, page, limit int) ([]Ticket, int, error)
	FindByIDAdmin(ctx context.Context, id string) (*Ticket, error)
	Reply(ctx context.Context, id, adminReply string) (*Ticket, error)
	MarkReadByMember(ctx context.Context, id, memberID string) error
	CountUnreadByMember(ctx context.Context, memberID string) (int, error)
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

const baseSelect = `
	SELECT t.id, t.member_id, m.full_name AS member_name, m.email AS member_email,
	       k.name AS korda_name, t.subject, t.message, t.status, t.admin_reply, t.replied_at,
	       t.member_read_at, t.created_at, t.updated_at
	FROM tickets t
	JOIN members m ON m.id = t.member_id
	LEFT JOIN kordas k ON k.id = m.korda_id
`

func (r *repository) Create(ctx context.Context, t *Ticket) error {
	query := `
		INSERT INTO tickets (id, member_id, subject, message, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at, updated_at
	`
	return r.db.QueryRowContext(ctx, query, t.ID, t.MemberID, t.Subject, t.Message, t.Status).
		Scan(&t.CreatedAt, &t.UpdatedAt)
}

func (r *repository) CountRecentByMember(ctx context.Context, memberID string, since time.Time) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count,
		`SELECT COUNT(*) FROM tickets WHERE member_id = $1 AND created_at > $2`,
		memberID, since,
	)
	return count, err
}

func (r *repository) FindAllByMember(ctx context.Context, memberID string, page, limit int) ([]Ticket, int, error) {
	var total int
	if err := r.db.GetContext(ctx, &total, `SELECT COUNT(*) FROM tickets WHERE member_id = $1`, memberID); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	tickets := []Ticket{}
	query := baseSelect + " WHERE t.member_id = $1 ORDER BY t.created_at DESC LIMIT $2 OFFSET $3"
	if err := r.db.SelectContext(ctx, &tickets, query, memberID, limit, offset); err != nil {
		return nil, 0, err
	}
	return tickets, total, nil
}

func (r *repository) FindByIDForMember(ctx context.Context, id, memberID string) (*Ticket, error) {
	var t Ticket
	query := baseSelect + " WHERE t.id = $1 AND t.member_id = $2"
	if err := r.db.GetContext(ctx, &t, query, id, memberID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &t, nil
}

func (r *repository) FindAllAdmin(ctx context.Context, kordaID, status string, page, limit int) ([]Ticket, int, error) {
	where := " WHERE 1=1"
	args := []interface{}{}
	argN := 1

	if kordaID != "" {
		where += fmt.Sprintf(" AND m.korda_id = $%d", argN)
		args = append(args, kordaID)
		argN++
	}
	if status != "" {
		where += fmt.Sprintf(" AND t.status = $%d", argN)
		args = append(args, status)
		argN++
	}

	var total int
	countQuery := "SELECT COUNT(*) FROM tickets t JOIN members m ON m.id = t.member_id" + where
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	args = append(args, limit, offset)
	query := baseSelect + where + fmt.Sprintf(" ORDER BY t.created_at DESC LIMIT $%d OFFSET $%d", argN, argN+1)

	tickets := []Ticket{}
	if err := r.db.SelectContext(ctx, &tickets, query, args...); err != nil {
		return nil, 0, err
	}
	return tickets, total, nil
}

func (r *repository) FindByIDAdmin(ctx context.Context, id string) (*Ticket, error) {
	var t Ticket
	query := baseSelect + " WHERE t.id = $1"
	if err := r.db.GetContext(ctx, &t, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &t, nil
}

func (r *repository) Reply(ctx context.Context, id, adminReply string) (*Ticket, error) {
	query := `
		UPDATE tickets
		SET admin_reply = $1, status = 'closed', replied_at = now(), updated_at = now()
		WHERE id = $2
		RETURNING id
	`
	var returnedID string
	if err := r.db.QueryRowContext(ctx, query, adminReply, id).Scan(&returnedID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return r.FindByIDAdmin(ctx, id)
}

func (r *repository) MarkReadByMember(ctx context.Context, id, memberID string) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE tickets SET member_read_at = now() WHERE id = $1 AND member_id = $2",
		id, memberID,
	)
	return err
}

func (r *repository) CountUnreadByMember(ctx context.Context, memberID string) (int, error) {
	var count int
	query := `
		SELECT COUNT(*) FROM tickets
		WHERE member_id = $1 AND status = 'closed'
		AND (member_read_at IS NULL OR member_read_at < replied_at)
	`
	err := r.db.GetContext(ctx, &count, query, memberID)
	return count, err
}
