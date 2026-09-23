package ticket

import (
	"context"
	"errors"
	"time"

	"github.com/oklog/ulid/v2"
)

var (
	ErrRateLimited = errors.New("rate limited")
	ErrNotApproved = errors.New("akun kamu masih menunggu persetujuan admin")
)

const rateLimitWindow = 5 * time.Minute

type CreateInput struct {
	MemberID         string
	Subject          string
	Message          string
	ReportedMemberID string
}

type Service interface {
	Create(ctx context.Context, in CreateInput) (*Ticket, error)
	ListMine(ctx context.Context, memberID string, page, limit int) ([]Ticket, int, error)
	GetMine(ctx context.Context, id, memberID string) (*Ticket, error)
	ListAdmin(ctx context.Context, kordaID, status string, page, limit int) ([]Ticket, int, error)
	GetAdmin(ctx context.Context, id string) (*Ticket, error)
	Reply(ctx context.Context, id, adminReply string) (*Ticket, error)
	MarkReadByMember(ctx context.Context, id, memberID string) error
	CountUnreadByMember(ctx context.Context, memberID string) (int, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, in CreateInput) (*Ticket, error) {
	status, err := s.repo.GetMemberStatus(ctx, in.MemberID)
	if err != nil {
		return nil, err
	}
	if status != "active" {
		return nil, ErrNotApproved
	}

	count, err := s.repo.CountRecentByMember(ctx, in.MemberID, time.Now().Add(-rateLimitWindow))
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, ErrRateLimited
	}

	var reportedPtr *string
	if in.ReportedMemberID != "" {
		reportedPtr = &in.ReportedMemberID
	}

	t := &Ticket{
		ID:               ulid.Make().String(),
		MemberID:         in.MemberID,
		Subject:          in.Subject,
		Message:          in.Message,
		Status:           StatusOpen,
		ReportedMemberID: reportedPtr,
	}
	if err := s.repo.Create(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *service) ListMine(ctx context.Context, memberID string, page, limit int) ([]Ticket, int, error) {
	return s.repo.FindAllByMember(ctx, memberID, page, limit)
}

func (s *service) GetMine(ctx context.Context, id, memberID string) (*Ticket, error) {
	return s.repo.FindByIDForMember(ctx, id, memberID)
}

func (s *service) ListAdmin(ctx context.Context, kordaID, status string, page, limit int) ([]Ticket, int, error) {
	return s.repo.FindAllAdmin(ctx, kordaID, status, page, limit)
}

func (s *service) GetAdmin(ctx context.Context, id string) (*Ticket, error) {
	return s.repo.FindByIDAdmin(ctx, id)
}

func (s *service) Reply(ctx context.Context, id, adminReply string) (*Ticket, error) {
	// TODO: setelah reply sukses, trigger email ke member (pola BrevoEmailSender
	// yang udah dipakai di reminder/broadcast), pakai t.MemberEmail dari hasil Reply().
	return s.repo.Reply(ctx, id, adminReply)
}

func (s *service) MarkReadByMember(ctx context.Context, id, memberID string) error {
	return s.repo.MarkReadByMember(ctx, id, memberID)
}

func (s *service) CountUnreadByMember(ctx context.Context, memberID string) (int, error) {
	return s.repo.CountUnreadByMember(ctx, memberID)
}
