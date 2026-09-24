package gpheroslide

import (
	"context"
	"errors"
	"mime/multipart"

	"github.com/oklog/ulid/v2"

	"github.com/Tibta65web/tibta65-server/pkg/storage"
)

var ErrValidation = errors.New("data tidak valid")

const uploadFolder = "gp-hero-slide"

type CreateInput struct {
	Headline            string
	Description         string
	PrimaryButtonText   string
	PrimaryButtonLink   string
	SecondaryButtonText string
	SecondaryButtonLink string
	Status              string
	Image               *multipart.FileHeader
}

type UpdateInput struct {
	ID                  string
	Headline            string
	Description         string
	PrimaryButtonText   string
	PrimaryButtonLink   string
	SecondaryButtonText string
	SecondaryButtonLink string
	Status              string
	Image               *multipart.FileHeader
}

type Service interface {
	ListAdmin(ctx context.Context) ([]GPHeroSlide, error)
	ListPublic(ctx context.Context) ([]GPHeroSlide, error)
	Create(ctx context.Context, in CreateInput) (*GPHeroSlide, error)
	Update(ctx context.Context, in UpdateInput) (*GPHeroSlide, error)
	Delete(ctx context.Context, id string) error
	Reorder(ctx context.Context, orderedIDs []string) error
}

type service struct {
	repo    Repository
	storage storage.Storage
}

func NewService(repo Repository, storage storage.Storage) Service {
	return &service{repo: repo, storage: storage}
}

func validateStatus(v string) bool {
	return v == StatusDraft || v == StatusPublished
}

func ptrOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func (s *service) ListAdmin(ctx context.Context) ([]GPHeroSlide, error) {
	return s.repo.FindAllAdmin(ctx)
}

func (s *service) ListPublic(ctx context.Context) ([]GPHeroSlide, error) {
	return s.repo.FindAllPublic(ctx)
}

func (s *service) Create(ctx context.Context, in CreateInput) (*GPHeroSlide, error) {
	if in.Headline == "" {
		return nil, ErrValidation
	}
	if in.Image == nil {
		return nil, errors.New("gambar wajib diupload")
	}
	if !validateStatus(in.Status) {
		return nil, ErrValidation
	}

	imageURL, err := s.storage.Upload(ctx, in.Image, uploadFolder)
	if err != nil {
		return nil, err
	}

	existing, err := s.repo.FindAllAdmin(ctx)
	if err != nil {
		_ = s.storage.Delete(ctx, imageURL)
		return nil, err
	}

	sl := &GPHeroSlide{
		ID: ulid.Make().String(), Headline: in.Headline, Description: ptrOrNil(in.Description),
		PrimaryButtonText: ptrOrNil(in.PrimaryButtonText), PrimaryButtonLink: ptrOrNil(in.PrimaryButtonLink),
		SecondaryButtonText: ptrOrNil(in.SecondaryButtonText), SecondaryButtonLink: ptrOrNil(in.SecondaryButtonLink),
		ImageURL: imageURL, Position: len(existing), Status: in.Status,
	}

	if err := s.repo.Create(ctx, sl); err != nil {
		_ = s.storage.Delete(ctx, imageURL)
		return nil, err
	}
	return sl, nil
}

func (s *service) Update(ctx context.Context, in UpdateInput) (*GPHeroSlide, error) {
	if in.Headline == "" {
		return nil, ErrValidation
	}
	if !validateStatus(in.Status) {
		return nil, ErrValidation
	}

	existing, err := s.repo.FindByID(ctx, in.ID)
	if err != nil {
		return nil, err
	}

	imageURL := existing.ImageURL
	if in.Image != nil {
		newImageURL, err := s.storage.Upload(ctx, in.Image, uploadFolder)
		if err != nil {
			return nil, err
		}
		imageURL = newImageURL
	}

	sl := &GPHeroSlide{
		ID: in.ID, Headline: in.Headline, Description: ptrOrNil(in.Description),
		PrimaryButtonText: ptrOrNil(in.PrimaryButtonText), PrimaryButtonLink: ptrOrNil(in.PrimaryButtonLink),
		SecondaryButtonText: ptrOrNil(in.SecondaryButtonText), SecondaryButtonLink: ptrOrNil(in.SecondaryButtonLink),
		ImageURL: imageURL, Status: in.Status,
	}

	if err := s.repo.Update(ctx, sl); err != nil {
		if in.Image != nil {
			_ = s.storage.Delete(ctx, imageURL)
		}
		return nil, err
	}
	if in.Image != nil && existing.ImageURL != "" {
		_ = s.storage.Delete(ctx, existing.ImageURL)
	}
	return sl, nil
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

func (s *service) Reorder(ctx context.Context, orderedIDs []string) error {
	return s.repo.Reorder(ctx, orderedIDs)
}
