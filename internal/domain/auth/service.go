package auth

import (
	"context"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/Tibta65web/tibta65-server/pkg/jwt"
)

var ErrInvalidCredentials = errors.New("invalid username or password")

type LoginResult struct {
	Token              string `json:"token"`
	Role               string `json:"role"`
	MustChangePassword bool   `json:"must_change_password"`
}

type Service interface {
	Login(ctx context.Context, username, password string) (*LoginResult, error)
	Me(ctx context.Context, adminID string) (*Admin, error)
}

type service struct {
	repo      Repository
	jwtSecret string
	jwtExpiry time.Duration
}

func NewService(repo Repository, jwtSecret string, jwtExpiry time.Duration) Service {
	return &service{
		repo:      repo,
		jwtSecret: jwtSecret,
		jwtExpiry: jwtExpiry,
	}
}

func (s *service) Login(ctx context.Context, username, password string) (*LoginResult, error) {
	admin, err := s.repo.FindByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, ErrAdminNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	_ = s.repo.UpdateLastLogin(ctx, admin.ID)

	token, err := jwt.GenerateToken(s.jwtSecret, admin.ID, admin.Role, s.jwtExpiry)
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		Token:              token,
		Role:               admin.Role,
		MustChangePassword: admin.MustChangePassword,
	}, nil
}

func (s *service) Me(ctx context.Context, adminID string) (*Admin, error) {
	return s.repo.FindByID(ctx, adminID)
}
