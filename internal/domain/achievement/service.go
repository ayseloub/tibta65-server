package achievement

import (
	"context"
	"errors"
	"mime/multipart"

	"github.com/oklog/ulid/v2"

	"github.com/Tibta65web/tibta65-server/pkg/storage"
)

var ErrValidation = errors.New("data tidak valid")

const (
	defaultLimit = 10
	maxLimit     = 50
	uploadFolder = "achievement"
)

type ListResult struct {
	Items      []Achievement `json:"items"`
	Page       int           `json:"page"`
	Limit      int           `json:"limit"`
	Total      int           `json:"total"`
	TotalPages int           `json:"total_pages"`
}

type Service interface {
	List(ctx context.Context, page, limit int, year string) (*ListResult, error)
	Get(ctx context.Context, id string) (*Achievement, error)
	Create(ctx context.Context, year, title, description string, file *multipart.FileHeader) (*Achievement, error)
	Update(ctx context.Context, id, year, title, description string, file *multipart.FileHeader) (*Achievement, error)
	Delete(ctx context.Context, id string) error
}

type service struct {
	repo    Repository
	storage storage.Storage
}

func NewService(repo Repository, storage storage.Storage) Service {
	return &service{repo: repo, storage: storage}
}

func (s *service) List(ctx context.Context, page, limit int, year string) (*ListResult, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > maxLimit {
		limit = defaultLimit
	}

	items, total, err := s.repo.FindAll(ctx, page, limit, year)
	if err != nil {
		return nil, err
	}

	totalPages := (total + limit - 1) / limit // pembulatan ke atas

	return &ListResult{
		Items:      items,
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

func (s *service) Get(ctx context.Context, id string) (*Achievement, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *service) Create(ctx context.Context, year, title, description string, file *multipart.FileHeader) (*Achievement, error) {
	if year == "" || title == "" || description == "" {
		return nil, ErrValidation
	}
	if file == nil {
		return nil, errors.New("gambar wajib diupload")
	}

	imageURL, err := s.storage.Upload(ctx, file, uploadFolder)
	if err != nil {
		return nil, err
	}

	a := &Achievement{
		ID:          ulid.Make().String(),
		Year:        year,
		Title:       title,
		Description: description,
		ImageURL:    imageURL,
	}

	if err := s.repo.Create(ctx, a); err != nil {
		// Rollback: kalau gagal simpan ke DB, hapus lagi gambar yang udah kepalang keupload,
		// biar gak ada file "yatim" (nempel di storage tapi gak ada referensinya di DB).
		_ = s.storage.Delete(ctx, imageURL)
		return nil, err
	}

	return a, nil
}

func (s *service) Update(ctx context.Context, id, year, title, description string, file *multipart.FileHeader) (*Achievement, error) {
	if year == "" || title == "" || description == "" {
		return nil, ErrValidation
	}

	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	imageURL := existing.ImageURL

	if file != nil {
		newImageURL, err := s.storage.Upload(ctx, file, uploadFolder)
		if err != nil {
			return nil, err
		}
		imageURL = newImageURL
	}

	a := &Achievement{
		ID:          id,
		Year:        year,
		Title:       title,
		Description: description,
		ImageURL:    imageURL,
	}

	if err := s.repo.Update(ctx, a); err != nil {
		if file != nil {
			_ = s.storage.Delete(ctx, imageURL)
		}
		return nil, err
	}

	if file != nil && existing.ImageURL != "" {
		_ = s.storage.Delete(ctx, existing.ImageURL)
	}

	return a, nil
}

func (s *service) Delete(ctx context.Context, id string) error {
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	_ = s.storage.Delete(ctx, existing.ImageURL)

	return nil
}
