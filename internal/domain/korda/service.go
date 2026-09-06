package korda

import (
	"context"
	"errors"

	"github.com/oklog/ulid/v2"
)

var ErrValidation = errors.New("nama wajib diisi")

type UpsertInput struct {
	Name      string
	Phone     string
	AdminName string
	Address   string
}

type Service interface {
	List(ctx context.Context) ([]Korda, error)
	Create(ctx context.Context, in UpsertInput) (*Korda, error)
	Update(ctx context.Context, id string, in UpsertInput) (*Korda, error)
	Delete(ctx context.Context, id string) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func strPtr(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}

func (s *service) List(ctx context.Context) ([]Korda, error) {
	return s.repo.FindAll(ctx)
}

func (s *service) Create(ctx context.Context, in UpsertInput) (*Korda, error) {
	if in.Name == "" {
		return nil, ErrValidation
	}
	k := &Korda{
		ID: ulid.Make().String(), Name: in.Name,
		Phone: strPtr(in.Phone), AdminName: strPtr(in.AdminName), Address: strPtr(in.Address),
	}
	if err := s.repo.Create(ctx, k); err != nil {
		return nil, err
	}
	return k, nil
}

func (s *service) Update(ctx context.Context, id string, in UpsertInput) (*Korda, error) {
	if in.Name == "" {
		return nil, ErrValidation
	}
	k := &Korda{ID: id, Name: in.Name, Phone: strPtr(in.Phone), AdminName: strPtr(in.AdminName), Address: strPtr(in.Address)}
	if err := s.repo.Update(ctx, k); err != nil {
		return nil, err
	}
	return k, nil
}

func (s *service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
