package backgroundcontent

import (
	"context"
	"errors"
)

var ErrInvalidSection = errors.New("section tidak valid")

type Service interface {
	Get(ctx context.Context, section string) (*BackgroundContent, error)
	Update(ctx context.Context, section, title, description string) (*BackgroundContent, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Get(ctx context.Context, section string) (*BackgroundContent, error) {
	if !IsValidSection(section) {
		return nil, ErrInvalidSection
	}

	return s.repo.FindBySection(ctx, section)
}

func (s *service) Update(ctx context.Context, section, title, description string) (*BackgroundContent, error) {
	if !IsValidSection(section) {
		return nil, ErrInvalidSection
	}

	if title == "" || description == "" {
		return nil, errors.New("judul dan deskripsi wajib diisi")
	}

	return s.repo.UpdateBySection(ctx, section, title, description)
}
