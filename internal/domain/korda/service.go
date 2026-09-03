package korda

import (
	"context"
	"errors"

	"github.com/oklog/ulid/v2"
)

var ErrValidation = errors.New("nama wajib diisi")

type Service interface {
	List(ctx context.Context) ([]Korda, error)
	Create(ctx context.Context, name string) (*Korda, error)
	Update(ctx context.Context, id, name string) (*Korda, error)
	Delete(ctx context.Context, id string) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) List(ctx context.Context) ([]Korda, error) {
	return s.repo.FindAll(ctx)
}

func (s *service) Create(ctx context.Context, name string) (*Korda, error) {
	if name == "" {
		return nil, ErrValidation
	}
	k := &Korda{ID: ulid.Make().String(), Name: name}
	if err := s.repo.Create(ctx, k); err != nil {
		return nil, err
	}
	return k, nil
}

func (s *service) Update(ctx context.Context, id, name string) (*Korda, error) {
	if name == "" {
		return nil, ErrValidation
	}
	k := &Korda{ID: id, Name: name}
	if err := s.repo.Update(ctx, k); err != nil {
		return nil, err
	}
	return k, nil
}

func (s *service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
