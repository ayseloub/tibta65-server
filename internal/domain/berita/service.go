package berita

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"regexp"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/Tibta65web/tibta65-server/pkg/storage"
)

var ErrValidation = errors.New("validasi gagal")

const (
	TitleMaxLen       = 60
	DescriptionMaxLen = 1000
)

type CreateInput struct {
	Title       string
	Description string
	Visibility  string
	Status      string
	EventDate   time.Time
	PublishAt   *time.Time
	ExpireAt    *time.Time
	ImageHeader *multipart.FileHeader
	AuthorID    string
	AuthorName  string
}

type UpdateInput struct {
	Title       string
	Description string
	Visibility  string
	Status      string
	EventDate   time.Time
	PublishAt   *time.Time
	ExpireAt    *time.Time
	ImageHeader *multipart.FileHeader
}

type Service interface {
	Create(ctx context.Context, in CreateInput) (*Berita, error)
	Update(ctx context.Context, id string, in UpdateInput) (*Berita, error)
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*Berita, error)
	FindAll(ctx context.Context, f ListFilter) ([]Berita, int, error)
	Stats(ctx context.Context) (total, published, draft int, err error)

	ToggleHighlight(ctx context.Context, id string, enable bool) (*Berita, error)
	FindHighlightPublic(ctx context.Context) (*Berita, error)
	FindAllPublic(ctx context.Context, page, limit int) ([]Berita, int, error)
	FindBySlugPublic(ctx context.Context, slug string) (*Berita, error)

	FindAllMember(ctx context.Context, page, limit int) ([]Berita, int, error)
	FindBySlugMember(ctx context.Context, slug string) (*Berita, error)
}

type service struct {
	repo    Repository
	storage storage.Storage
}

func NewService(repo Repository, storage storage.Storage) Service {
	return &service{repo: repo, storage: storage}
}

var slugSanitizer = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(title string) string {
	s := strings.ToLower(title)
	s = slugSanitizer.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

func (s *service) generateUniqueSlug(ctx context.Context, title string) (string, error) {
	base := slugify(title)
	slug := base
	i := 2
	for {
		exists, err := s.repo.SlugExists(ctx, slug)
		if err != nil {
			return "", err
		}
		if !exists {
			return slug, nil
		}
		slug = fmt.Sprintf("%s-%d", base, i)
		i++
	}
}

func validateVisibility(v string) bool {
	return v == VisibilityPublic || v == VisibilityInternal
}

func validateStatus(v string) bool {
	return v == StatusDraft || v == StatusPublished
}

func (s *service) Create(ctx context.Context, in CreateInput) (*Berita, error) {
	if in.Title == "" || len(in.Title) > TitleMaxLen {
		return nil, ErrValidation
	}
	if in.Description == "" || len(in.Description) > DescriptionMaxLen {
		return nil, ErrValidation
	}
	if !validateVisibility(in.Visibility) || !validateStatus(in.Status) {
		return nil, ErrValidation
	}
	if in.ImageHeader == nil {
		return nil, ErrValidation
	}
	if in.PublishAt != nil && in.ExpireAt != nil && !in.ExpireAt.After(*in.PublishAt) {
		return nil, errors.New("waktu berakhir harus setelah waktu terbit")
	}

	imageURL, err := s.storage.Upload(ctx, in.ImageHeader, "berita")
	if err != nil {
		return nil, err
	}

	slug, err := s.generateUniqueSlug(ctx, in.Title)
	if err != nil {
		return nil, err
	}

	authorID := in.AuthorID
	b := &Berita{
		ID:          ulid.Make().String(),
		Title:       in.Title,
		Slug:        slug,
		Description: in.Description,
		ImageURL:    imageURL,
		Visibility:  in.Visibility,
		Status:      in.Status,
		AuthorID:    &authorID,
		AuthorName:  in.AuthorName,
		EventDate:   in.EventDate,
		PublishAt:   in.PublishAt,
		ExpireAt:    in.ExpireAt,
	}
	if err := s.repo.Create(ctx, b); err != nil {
		return nil, err
	}
	return b, nil
}

func (s *service) Update(ctx context.Context, id string, in UpdateInput) (*Berita, error) {
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if in.Title == "" || len(in.Title) > TitleMaxLen {
		return nil, ErrValidation
	}
	if in.Description == "" || len(in.Description) > DescriptionMaxLen {
		return nil, ErrValidation
	}
	if !validateVisibility(in.Visibility) || !validateStatus(in.Status) {
		return nil, ErrValidation
	}
	if in.PublishAt != nil && in.ExpireAt != nil && !in.ExpireAt.After(*in.PublishAt) {
		return nil, errors.New("waktu berakhir harus setelah waktu terbit")
	}

	imageURL := existing.ImageURL
	if in.ImageHeader != nil {
		imageURL, err = s.storage.Upload(ctx, in.ImageHeader, "berita")
		if err != nil {
			return nil, err
		}
	}

	slug := existing.Slug
	if in.Title != existing.Title {
		slug, err = s.generateUniqueSlug(ctx, in.Title)
		if err != nil {
			return nil, err
		}
	}

	existing.Title = in.Title
	existing.Slug = slug
	existing.Description = in.Description
	existing.ImageURL = imageURL
	existing.Visibility = in.Visibility
	existing.Status = in.Status
	existing.EventDate = in.EventDate
	existing.PublishAt = in.PublishAt
	existing.ExpireAt = in.ExpireAt

	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *service) FindByID(ctx context.Context, id string) (*Berita, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *service) FindAll(ctx context.Context, f ListFilter) ([]Berita, int, error) {
	return s.repo.FindAll(ctx, f)
}

func (s *service) Stats(ctx context.Context) (total, published, draft int, err error) {
	total, err = s.repo.CountAll(ctx)
	if err != nil {
		return
	}
	published, err = s.repo.CountByStatus(ctx, StatusPublished)
	if err != nil {
		return
	}
	draft, err = s.repo.CountByStatus(ctx, StatusDraft)
	return
}

func (s *service) ToggleHighlight(ctx context.Context, id string, enable bool) (*Berita, error) {
	if enable {
		if err := s.repo.SetHighlight(ctx, id); err != nil {
			return nil, err
		}
	} else {
		if err := s.repo.UnsetHighlight(ctx, id); err != nil {
			return nil, err
		}
	}
	return s.repo.FindByID(ctx, id)
}

func (s *service) FindHighlightPublic(ctx context.Context) (*Berita, error) {
	return s.repo.FindHighlightPublic(ctx)
}

func (s *service) FindAllPublic(ctx context.Context, page, limit int) ([]Berita, int, error) {
	return s.repo.FindAllPublic(ctx, page, limit)
}

func (s *service) FindBySlugPublic(ctx context.Context, slug string) (*Berita, error) {
	return s.repo.FindBySlugPublic(ctx, slug)
}

func (s *service) FindAllMember(ctx context.Context, page, limit int) ([]Berita, int, error) {
	return s.repo.FindAllMember(ctx, page, limit)
}

func (s *service) FindBySlugMember(ctx context.Context, slug string) (*Berita, error) {
	b, err := s.repo.FindBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if b.Status != StatusPublished || b.Visibility != VisibilityInternal {
		return nil, ErrNotFound
	}
	if b.PublishAt != nil && b.PublishAt.After(time.Now()) {
		return nil, ErrNotFound
	}
	if b.ExpireAt != nil && !b.ExpireAt.After(time.Now()) {
		return nil, ErrNotFound
	}
	return b, nil
}
