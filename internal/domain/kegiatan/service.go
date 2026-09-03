package kegiatan

import (
	"context"
	"errors"
	"mime/multipart"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/Tibta65web/tibta65-server/pkg/slug"
	"github.com/Tibta65web/tibta65-server/pkg/storage"
)

var ErrValidation = errors.New("data tidak valid")

const uploadFolder = "kegiatan"

type ListResult struct {
	Items      []Kegiatan `json:"items"`
	Page       int        `json:"page"`
	Limit      int        `json:"limit"`
	Total      int        `json:"total"`
	TotalPages int        `json:"total_pages"`
}

type CreateInput struct {
	Title       string
	Date        string
	KordaID     string
	KategoriID  string
	Location    string
	Description string
	Image       *multipart.FileHeader
}

type UpdateInput struct {
	Slug        string
	Title       string
	Date        string
	KordaID     string
	KategoriID  string
	Location    string
	Description string
	Image       *multipart.FileHeader
}

type Service interface {
	List(ctx context.Context, f ListFilter) (*ListResult, error)
	Get(ctx context.Context, slug string) (*Kegiatan, error)
	Create(ctx context.Context, in CreateInput) (*Kegiatan, error)
	Update(ctx context.Context, in UpdateInput) (*Kegiatan, error)
	Delete(ctx context.Context, slug string) error
}

type service struct {
	repo    Repository
	storage storage.Storage
}

func NewService(repo Repository, storage storage.Storage) Service {
	return &service{repo: repo, storage: storage}
}

func (s *service) List(ctx context.Context, f ListFilter) (*ListResult, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Limit < 1 || f.Limit > 50 {
		f.Limit = 5
	}

	items, total, err := s.repo.FindAll(ctx, f)
	if err != nil {
		return nil, err
	}

	totalPages := (total + f.Limit - 1) / f.Limit

	return &ListResult{
		Items: items, Page: f.Page, Limit: f.Limit, Total: total, TotalPages: totalPages,
	}, nil
}

func (s *service) Get(ctx context.Context, slugParam string) (*Kegiatan, error) {
	return s.repo.FindBySlug(ctx, slugParam)
}

func (s *service) Create(ctx context.Context, in CreateInput) (*Kegiatan, error) {
	if in.Title == "" || in.Date == "" || in.KordaID == "" || in.KategoriID == "" || in.Location == "" || in.Description == "" {
		return nil, ErrValidation
	}
	if in.Image == nil {
		return nil, errors.New("gambar wajib diupload")
	}

	date, err := time.Parse("2006-01-02", in.Date)
	if err != nil {
		return nil, errors.New("format tanggal tidak valid")
	}

	imageURL, err := s.storage.Upload(ctx, in.Image, uploadFolder)
	if err != nil {
		return nil, err
	}

	k := &Kegiatan{
		ID:          ulid.Make().String(),
		Slug:        slug.Generate(in.Title),
		Title:       in.Title,
		Date:        date,
		KordaID:     in.KordaID,
		KategoriID:  in.KategoriID,
		Location:    in.Location,
		ImageURL:    imageURL,
		Description: in.Description,
	}

	if err := s.repo.Create(ctx, k); err != nil {
		_ = s.storage.Delete(ctx, imageURL)
		return nil, err
	}

	return k, nil
}

func (s *service) Update(ctx context.Context, in UpdateInput) (*Kegiatan, error) {
	if in.Title == "" || in.Date == "" || in.KordaID == "" || in.KategoriID == "" || in.Location == "" || in.Description == "" {
		return nil, ErrValidation
	}

	existing, err := s.repo.FindBySlug(ctx, in.Slug)
	if err != nil {
		return nil, err
	}

	date, err := time.Parse("2006-01-02", in.Date)
	if err != nil {
		return nil, errors.New("format tanggal tidak valid")
	}

	imageURL := existing.ImageURL
	if in.Image != nil {
		newImageURL, err := s.storage.Upload(ctx, in.Image, uploadFolder)
		if err != nil {
			return nil, err
		}
		imageURL = newImageURL
	}

	k := &Kegiatan{
		Slug: in.Slug, Title: in.Title, Date: date, KordaID: in.KordaID,
		KategoriID: in.KategoriID, Location: in.Location, ImageURL: imageURL, Description: in.Description,
	}

	if err := s.repo.Update(ctx, k); err != nil {
		if in.Image != nil {
			_ = s.storage.Delete(ctx, imageURL)
		}
		return nil, err
	}

	if in.Image != nil && existing.ImageURL != "" {
		_ = s.storage.Delete(ctx, existing.ImageURL)
	}

	return k, nil
}

func (s *service) Delete(ctx context.Context, slugParam string) error {
	existing, err := s.repo.FindBySlug(ctx, slugParam)
	if err != nil {
		return err
	}
	if err := s.repo.Delete(ctx, slugParam); err != nil {
		return err
	}
	_ = s.storage.Delete(ctx, existing.ImageURL)
	return nil
}
