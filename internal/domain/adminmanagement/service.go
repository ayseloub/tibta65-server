package adminmanagement

import (
	"context"
	"errors"

	"github.com/oklog/ulid/v2"
	"golang.org/x/crypto/bcrypt"

	"github.com/Tibta65web/tibta65-server/internal/domain/auth"
)

var ErrValidation = errors.New("data tidak valid")

type ListResult struct {
	Items      []auth.Admin `json:"items"`
	Page       int          `json:"page"`
	Limit      int          `json:"limit"`
	Total      int          `json:"total"`
	TotalPages int          `json:"total_pages"`
}

type CreateInput struct {
	FullName string
	Username string
	Role     string
	Password string
}

type UpdateInput struct {
	ID       string
	FullName string
	Username string
	Role     string
	Password string
}

type Service interface {
	List(ctx context.Context, page, limit int) (*ListResult, error)
	Create(ctx context.Context, in CreateInput) (*auth.Admin, error)
	Update(ctx context.Context, in UpdateInput) (*auth.Admin, error)
	Delete(ctx context.Context, id string) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) List(ctx context.Context, page, limit int) (*ListResult, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 5
	}

	items, total, err := s.repo.FindAll(ctx, page, limit)
	if err != nil {
		return nil, err
	}

	totalPages := (total + limit - 1) / limit
	return &ListResult{Items: items, Page: page, Limit: limit, Total: total, TotalPages: totalPages}, nil
}

func validateUsername(u string) error {
	if len(u) < 8 || len(u) > 12 {
		return errors.New("username harus 8-12 karakter")
	}
	return nil
}

func validateRole(r string) error {
	if r != auth.RoleAdmin && r != auth.RoleSuperAdmin {
		return errors.New("role tidak valid")
	}
	return nil
}

func (s *service) Create(ctx context.Context, in CreateInput) (*auth.Admin, error) {
	if in.FullName == "" || in.Username == "" || in.Role == "" || in.Password == "" {
		return nil, ErrValidation
	}
	if err := validateUsername(in.Username); err != nil {
		return nil, err
	}
	if err := validateRole(in.Role); err != nil {
		return nil, err
	}
	if len(in.Password) < 8 {
		return nil, errors.New("password minimal 8 karakter")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	a := &auth.Admin{
		ID:           ulid.Make().String(),
		Username:     in.Username,
		FullName:     in.FullName,
		PasswordHash: string(hash),
		Role:         in.Role,
	}

	if err := s.repo.Create(ctx, a); err != nil {
		return nil, err
	}
	return a, nil
}

func (s *service) Update(ctx context.Context, in UpdateInput) (*auth.Admin, error) {
	if in.FullName == "" || in.Username == "" || in.Role == "" {
		return nil, ErrValidation
	}
	if err := validateUsername(in.Username); err != nil {
		return nil, err
	}
	if err := validateRole(in.Role); err != nil {
		return nil, err
	}

	a := &auth.Admin{ID: in.ID, Username: in.Username, FullName: in.FullName, Role: in.Role}
	if err := s.repo.Update(ctx, a); err != nil {
		return nil, err
	}

	if in.Password != "" {
		if len(in.Password) < 8 {
			return nil, errors.New("password minimal 8 karakter")
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		if err := s.repo.UpdatePassword(ctx, in.ID, string(hash)); err != nil {
			return nil, err
		}
	}

	return a, nil
}

func (s *service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
