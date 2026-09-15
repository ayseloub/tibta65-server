package heroslide

import (
	"context"
	"errors"
	"mime/multipart"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/Tibta65web/tibta65-server/pkg/storage"
)

var ErrValidation = errors.New("data tidak valid")

const uploadFolder = "hero-slide"

type CreateInput struct {
	Headline            string
	Description         string
	PrimaryButtonText   string
	PrimaryButtonLink   string
	SecondaryButtonText string
	SecondaryButtonLink string
	BgColor             string
	Status              string
	PublishAt           *time.Time
	ExpireAt            *time.Time
	Image               *multipart.FileHeader
}

type UpdateInput struct {
	Headline            string
	Description         string
	PrimaryButtonText   string
	PrimaryButtonLink   string
	SecondaryButtonText string
	SecondaryButtonLink string
	BgColor             string
	Status              string
	PublishAt           *time.Time
	ExpireAt            *time.Time
	Image               *multipart.FileHeader
}

type Service interface {
	ListAdmin(ctx context.Context) ([]HeroSlide, error)
	ListPublic(ctx context.Context) ([]HeroSlide, error)
	Get(ctx context.Context, id string) (*HeroSlide, error)
	Create(ctx context.Context, in CreateInput) (*HeroSlide, error)
	Update(ctx context.Context, id string, in UpdateInput) (*HeroSlide, error)
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

func (s *service) ListAdmin(ctx context.Context) ([]HeroSlide, error) {
	return s.repo.FindAllAdmin(ctx)
}

func (s *service) ListPublic(ctx context.Context) ([]HeroSlide, error) {
	return s.repo.FindAllPublic(ctx)
}

func (s *service) Get(ctx context.Context, id string) (*HeroSlide, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *service) Create(ctx context.Context, in CreateInput) (*HeroSlide, error) {
	if in.Headline == "" || in.Description == "" || in.PrimaryButtonText == "" || in.PrimaryButtonLink == "" {
		return nil, ErrValidation
	}
	if !validateStatus(in.Status) {
		return nil, ErrValidation
	}
	if in.Image == nil {
		return nil, errors.New("gambar wajib diupload")
	}
	if in.PublishAt != nil && in.ExpireAt != nil && !in.ExpireAt.After(*in.PublishAt) {
		return nil, errors.New("waktu berakhir harus setelah waktu terbit")
	}

	imageURL, err := s.storage.Upload(ctx, in.Image, uploadFolder)
	if err != nil {
		return nil, err
	}

	maxPos, err := s.repo.MaxPosition(ctx)
	if err != nil {
		return nil, err
	}

	var secondaryText, secondaryLink *string
	if in.SecondaryButtonText != "" {
		secondaryText = &in.SecondaryButtonText
	}
	if in.SecondaryButtonLink != "" {
		secondaryLink = &in.SecondaryButtonLink
	}

	bgColor := in.BgColor
	if bgColor == "" {
		bgColor = "#F8EFDC"
	}

	slide := &HeroSlide{
		ID:                  ulid.Make().String(),
		Headline:            in.Headline,
		Description:         in.Description,
		PrimaryButtonText:   in.PrimaryButtonText,
		PrimaryButtonLink:   in.PrimaryButtonLink,
		SecondaryButtonText: secondaryText,
		SecondaryButtonLink: secondaryLink,
		ImageURL:            imageURL,
		BgColor:             bgColor,
		Position:            maxPos + 1,
		Status:              in.Status,
		PublishAt:           in.PublishAt,
		ExpireAt:            in.ExpireAt,
	}

	if err := s.repo.Create(ctx, slide); err != nil {
		_ = s.storage.Delete(ctx, imageURL)
		return nil, err
	}

	return slide, nil
}

func (s *service) Update(ctx context.Context, id string, in UpdateInput) (*HeroSlide, error) {
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if in.Headline == "" || in.Description == "" || in.PrimaryButtonText == "" || in.PrimaryButtonLink == "" {
		return nil, ErrValidation
	}
	if !validateStatus(in.Status) {
		return nil, ErrValidation
	}
	if in.PublishAt != nil && in.ExpireAt != nil && !in.ExpireAt.After(*in.PublishAt) {
		return nil, errors.New("waktu berakhir harus setelah waktu terbit")
	}

	imageURL := existing.ImageURL
	if in.Image != nil {
		imageURL, err = s.storage.Upload(ctx, in.Image, uploadFolder)
		if err != nil {
			return nil, err
		}
	}

	var secondaryText, secondaryLink *string
	if in.SecondaryButtonText != "" {
		secondaryText = &in.SecondaryButtonText
	}
	if in.SecondaryButtonLink != "" {
		secondaryLink = &in.SecondaryButtonLink
	}

	bgColor := in.BgColor
	if bgColor == "" {
		bgColor = existing.BgColor
	}

	existing.Headline = in.Headline
	existing.Description = in.Description
	existing.PrimaryButtonText = in.PrimaryButtonText
	existing.PrimaryButtonLink = in.PrimaryButtonLink
	existing.SecondaryButtonText = secondaryText
	existing.SecondaryButtonLink = secondaryLink
	existing.ImageURL = imageURL
	existing.BgColor = bgColor
	existing.Status = in.Status
	existing.PublishAt = in.PublishAt
	existing.ExpireAt = in.ExpireAt

	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}
	return existing, nil
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
	if len(orderedIDs) == 0 {
		return ErrValidation
	}
	return s.repo.Reorder(ctx, orderedIDs)
}
