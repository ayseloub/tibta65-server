package membermanagement

import (
	"context"
	"errors"
	"fmt"

	"github.com/oklog/ulid/v2"

	"github.com/Tibta65web/tibta65-server/internal/domain/member"
)

type ListResult struct {
	Items      []member.Member `json:"items"`
	Page       int             `json:"page"`
	Limit      int             `json:"limit"`
	Total      int             `json:"total"`
	TotalPages int             `json:"total_pages"`
}

type Service interface {
	ListPending(ctx context.Context, page, limit int) (*ListResult, error)
	ListAll(ctx context.Context, page, limit int) (*ListResult, error)
	Approve(ctx context.Context, id string) error
	Delete(ctx context.Context, id string) error
	CreateLegacyMember(ctx context.Context, in CreateLegacyMemberInput) (*member.Member, error)
	UpdateStatus(ctx context.Context, id, status string) error
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

var ErrInvalidStatusTransition = errors.New("status tujuan tidak valid")

func (s *service) UpdateStatus(ctx context.Context, id, status string) error {
	if status != member.StatusDeceased && status != member.StatusRejected {
		return ErrInvalidStatusTransition
	}
	return s.repo.UpdateStatus(ctx, id, status)
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

func normalizePaging(page, limit int) (int, int) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 5
	}
	return page, limit
}

func buildResult(items []member.Member, page, limit, total int) *ListResult {
	totalPages := (total + limit - 1) / limit
	if totalPages < 1 {
		totalPages = 1
	}
	return &ListResult{Items: items, Page: page, Limit: limit, Total: total, TotalPages: totalPages}
}

func (s *service) ListPending(ctx context.Context, page, limit int) (*ListResult, error) {
	page, limit = normalizePaging(page, limit)
	items, total, err := s.repo.FindPending(ctx, page, limit)
	if err != nil {
		return nil, err
	}
	return buildResult(items, page, limit, total), nil
}

func (s *service) ListAll(ctx context.Context, page, limit int) (*ListResult, error) {
	page, limit = normalizePaging(page, limit)
	items, total, err := s.repo.FindAll(ctx, page, limit)
	if err != nil {
		return nil, err
	}
	return buildResult(items, page, limit, total), nil
}

func (s *service) Approve(ctx context.Context, id string) error {
	return s.repo.Approve(ctx, id)
}

func (s *service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
