package membermanagement

import (
	"context"

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
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
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
