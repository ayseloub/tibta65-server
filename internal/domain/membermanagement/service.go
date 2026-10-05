package membermanagement

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/oklog/ulid/v2"

	"github.com/Tibta65web/tibta65-server/internal/domain/member"
)

var (
	ErrInvalidStatusTransition = errors.New("status tujuan tidak valid")
	ErrInvalidRejectionType    = errors.New("jenis penolakan tidak valid")
	ErrReasonRequired          = errors.New("alasan penolakan wajib diisi")
)

const defaultReopenReason = "Pendaftaranmu dibuka kembali oleh admin. Periksa dan lengkapi biodata, lalu kirim ulang untuk direview."

type ListResult struct {
	Items      []member.Member `json:"items"`
	Page       int             `json:"page"`
	Limit      int             `json:"limit"`
	Total      int             `json:"total"`
	TotalPages int             `json:"total_pages"`
}

type Service interface {
	ListPending(ctx context.Context, f ListFilter) (*ListResult, error)
	ListAll(ctx context.Context, f ListFilter) (*ListResult, error)
	Approve(ctx context.Context, id string) error
	Reject(ctx context.Context, id, rejectionType, reason string) error
	Reopen(ctx context.Context, id, reason string) error
	ApproveRejected(ctx context.Context, id string) error
	UpdateStatus(ctx context.Context, id, status string) error
	Delete(ctx context.Context, id string) error
	CreateLegacyMember(ctx context.Context, in CreateLegacyMemberInput) (*member.Member, error)
	Export(ctx context.Context, in ExportInput) ([]byte, error)
	ExportFields() []ExportFieldOption
}

type service struct {
	repo       Repository
	memberRepo member.Repository
}

func NewService(repo Repository, memberRepo member.Repository) Service {
	return &service{repo: repo, memberRepo: memberRepo}
}

type CreateLegacyMemberInput struct {
	FullName        string
	Generation      int
	ParentMemberID  string
	NamaSuci        string
	Agama           string
	NRP             string
	NoAK            string
	PangkatTerakhir string
}

func (s *service) CreateLegacyMember(ctx context.Context, in CreateLegacyMemberInput) (*member.Member, error) {
	if in.FullName == "" {
		return nil, member.ErrValidation
	}
	generation := in.Generation
	if generation < 1 {
		generation = 1
	}

	memberNumber, err := s.memberRepo.NextMemberNumber(ctx, generation)
	if err != nil {
		return nil, err
	}

	var namaSuciPtr, agamaPtr, nrpPtr, noAkPtr, pangkatPtr, parentPtr *string
	if in.NamaSuci != "" {
		namaSuciPtr = &in.NamaSuci
	}
	if in.Agama != "" {
		agamaPtr = &in.Agama
	}
	if in.NRP != "" {
		nrpPtr = &in.NRP
	}
	if in.NoAK != "" {
		noAkPtr = &in.NoAK
	}
	if in.PangkatTerakhir != "" {
		pangkatPtr = &in.PangkatTerakhir
	}
	if in.ParentMemberID != "" {
		parentPtr = &in.ParentMemberID
	}

	placeholderEmail := fmt.Sprintf("unclaimed+%s@tibta65.local", memberNumber)

	m := &member.Member{
		ID:               ulid.Make().String(),
		FullName:         in.FullName,
		Email:            placeholderEmail,
		MemberNumber:     memberNumber,
		Generation:       generation,
		ParentMemberID:   parentPtr,
		Status:           member.StatusUnclaimed,
		ProfileCompleted: true,
		NamaSuci:         namaSuciPtr,
		Agama:            agamaPtr,
		NRP:              nrpPtr,
		NoAK:             noAkPtr,
		PangkatTerakhir:  pangkatPtr,
	}

	if err := s.memberRepo.Create(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

func normalizePaging(f ListFilter) ListFilter {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Limit < 1 || f.Limit > 50 {
		f.Limit = 20
	}
	return f
}

func buildResult(items []member.Member, f ListFilter, total int) *ListResult {
	totalPages := (total + f.Limit - 1) / f.Limit
	if totalPages < 1 {
		totalPages = 1
	}
	return &ListResult{Items: items, Page: f.Page, Limit: f.Limit, Total: total, TotalPages: totalPages}
}

func (s *service) ListPending(ctx context.Context, f ListFilter) (*ListResult, error) {
	f = normalizePaging(f)
	items, total, err := s.repo.FindPending(ctx, f)
	if err != nil {
		return nil, err
	}
	return buildResult(items, f, total), nil
}

func (s *service) ListAll(ctx context.Context, f ListFilter) (*ListResult, error) {
	f = normalizePaging(f)
	items, total, err := s.repo.FindAll(ctx, f)
	if err != nil {
		return nil, err
	}
	return buildResult(items, f, total), nil
}

func (s *service) Approve(ctx context.Context, id string) error {
	return s.repo.Approve(ctx, id)
}

func (s *service) Reject(ctx context.Context, id, rejectionType, reason string) error {
	if rejectionType != member.RejectionIncomplete && rejectionType != member.RejectionSuspicious {
		return ErrInvalidRejectionType
	}
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return ErrReasonRequired
	}
	return s.repo.Reject(ctx, id, rejectionType, reason)
}

func (s *service) Reopen(ctx context.Context, id, reason string) error {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		reason = defaultReopenReason
	}
	return s.repo.Reopen(ctx, id, reason)
}

func (s *service) ApproveRejected(ctx context.Context, id string) error {
	return s.repo.ApproveRejected(ctx, id)
}

func (s *service) UpdateStatus(ctx context.Context, id, status string) error {
	if status != member.StatusDeceased {
		return ErrInvalidStatusTransition
	}
	return s.repo.UpdateStatus(ctx, id, status)
}

func (s *service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
