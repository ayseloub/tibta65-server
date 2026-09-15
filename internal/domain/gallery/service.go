package gallery

import (
	"context"
	"errors"
	"mime/multipart"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/Tibta65web/tibta65-server/pkg/storage"
)

var ErrValidation = errors.New("data tidak valid")

const uploadFolder = "gallery"

const (
	VisibilityPublic   = "public"
	VisibilityInternal = "internal"
)

type ListResult struct {
	Items      []Album `json:"items"`
	Page       int     `json:"page"`
	Limit      int     `json:"limit"`
	Total      int     `json:"total"`
	TotalPages int     `json:"total_pages"`
}

type AlbumInput struct {
	Title       string
	Description string
	EventDate   string // format "2006-01-02"
	KordaID     string
	KategoriID  string
	Visibility  string
}

type Service interface {
	List(ctx context.Context, search, visibility string, page, limit int) (*ListResult, error)
	Get(ctx context.Context, id string) (*AlbumDetail, error)
	GetHighlight(ctx context.Context) (*AlbumDetail, error)
	Create(ctx context.Context, in AlbumInput) (*Album, error)
	Update(ctx context.Context, id string, in AlbumInput) (*Album, error)
	Delete(ctx context.Context, id string) error
	SetHighlight(ctx context.Context, id string, highlight bool) error

	AddPhoto(ctx context.Context, albumID string, file *multipart.FileHeader, caption string) (*Photo, error)
	UpdatePhotoCaption(ctx context.Context, photoID, caption string) error
	DeletePhoto(ctx context.Context, photoID string) error
}

type service struct {
	repo    Repository
	storage storage.Storage
}

func NewService(repo Repository, storage storage.Storage) Service {
	return &service{repo: repo, storage: storage}
}

func validateVisibility(v string) bool {
	return v == VisibilityPublic || v == VisibilityInternal
}

func (s *service) List(ctx context.Context, search, visibility string, page, limit int) (*ListResult, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 8
	}

	items, total, err := s.repo.FindAll(ctx, ListFilter{Search: search, Visibility: visibility, Page: page, Limit: limit})
	if err != nil {
		return nil, err
	}

	totalPages := (total + limit - 1) / limit
	return &ListResult{Items: items, Page: page, Limit: limit, Total: total, TotalPages: totalPages}, nil
}

func (s *service) Get(ctx context.Context, id string) (*AlbumDetail, error) {
	album, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	photos, err := s.repo.FindPhotosByAlbumID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &AlbumDetail{Album: *album, Photos: photos}, nil
}

func (s *service) GetHighlight(ctx context.Context) (*AlbumDetail, error) {
	album, err := s.repo.FindHighlight(ctx)
	if err != nil {
		return nil, err
	}
	photos, err := s.repo.FindPhotosByAlbumID(ctx, album.ID)
	if err != nil {
		return nil, err
	}
	return &AlbumDetail{Album: *album, Photos: photos}, nil
}

func parseAlbumInput(in AlbumInput) (title, description string, eventDate time.Time, kordaID, kategoriID *string, visibility string, err error) {
	if in.Title == "" || in.Description == "" || in.EventDate == "" {
		err = ErrValidation
		return
	}
	if !validateVisibility(in.Visibility) {
		err = ErrValidation
		return
	}

	eventDate, err = time.Parse("2006-01-02", in.EventDate)
	if err != nil {
		err = errors.New("format tanggal tidak valid")
		return
	}

	if in.KordaID != "" {
		kordaID = &in.KordaID
	}
	if in.KategoriID != "" {
		kategoriID = &in.KategoriID
	}

	return in.Title, in.Description, eventDate, kordaID, kategoriID, in.Visibility, nil
}

func (s *service) Create(ctx context.Context, in AlbumInput) (*Album, error) {
	title, description, eventDate, kordaID, kategoriID, visibility, err := parseAlbumInput(in)
	if err != nil {
		return nil, err
	}

	a := &Album{
		ID: ulid.Make().String(), Title: title, Description: description,
		EventDate: eventDate, KordaID: kordaID, KategoriID: kategoriID, Visibility: visibility,
	}

	if err := s.repo.Create(ctx, a); err != nil {
		return nil, err
	}
	return a, nil
}

func (s *service) Update(ctx context.Context, id string, in AlbumInput) (*Album, error) {
	title, description, eventDate, kordaID, kategoriID, visibility, err := parseAlbumInput(in)
	if err != nil {
		return nil, err
	}

	a := &Album{ID: id, Title: title, Description: description, EventDate: eventDate, KordaID: kordaID, KategoriID: kategoriID, Visibility: visibility}
	if err := s.repo.Update(ctx, a); err != nil {
		return nil, err
	}
	return a, nil
}

func (s *service) Delete(ctx context.Context, id string) error {
	photos, err := s.repo.FindPhotosByAlbumID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	for _, p := range photos {
		_ = s.storage.Delete(ctx, p.ImageURL)
	}

	return nil
}

func (s *service) SetHighlight(ctx context.Context, id string, highlight bool) error {
	return s.repo.SetHighlight(ctx, id, highlight)
}

func (s *service) AddPhoto(ctx context.Context, albumID string, file *multipart.FileHeader, caption string) (*Photo, error) {
	if _, err := s.repo.FindByID(ctx, albumID); err != nil {
		return nil, err
	}

	count, err := s.repo.CountPhotos(ctx, albumID)
	if err != nil {
		return nil, err
	}
	if count >= MaxPhotosPerAlbum {
		return nil, ErrMaxPhotos
	}

	imageURL, err := s.storage.Upload(ctx, file, uploadFolder)
	if err != nil {
		return nil, err
	}

	var captionPtr *string
	if caption != "" {
		captionPtr = &caption
	}

	p := &Photo{
		ID: ulid.Make().String(), AlbumID: albumID, ImageURL: imageURL,
		Caption: captionPtr, SortOrder: count,
	}

	if err := s.repo.AddPhoto(ctx, p); err != nil {
		_ = s.storage.Delete(ctx, imageURL)
		return nil, err
	}

	return p, nil
}

func (s *service) UpdatePhotoCaption(ctx context.Context, photoID, caption string) error {
	var captionPtr *string
	if caption != "" {
		captionPtr = &caption
	}
	return s.repo.UpdatePhotoCaption(ctx, photoID, captionPtr)
}

func (s *service) DeletePhoto(ctx context.Context, photoID string) error {
	photo, err := s.repo.FindPhotoByID(ctx, photoID)
	if err != nil {
		return err
	}

	if err := s.repo.DeletePhoto(ctx, photoID); err != nil {
		return err
	}

	_ = s.storage.Delete(ctx, photo.ImageURL)
	return nil
}
