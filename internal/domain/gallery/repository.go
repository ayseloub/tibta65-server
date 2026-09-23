package gallery

import (
	"context"
	"database/sql"
	"errors"
	"strconv"

	"github.com/jmoiron/sqlx"
)

var (
	ErrNotFound      = errors.New("album not found")
	ErrPhotoNotFound = errors.New("photo not found")
	ErrMaxPhotos     = errors.New("album sudah mencapai batas maksimal 30 foto")
)

const MaxPhotosPerAlbum = 30

type ListFilter struct {
	Search           string
	Visibility       string
	Scheduled        bool
	MemberGeneration int
	Page             int
	Limit            int
}

type Repository interface {
	FindAll(ctx context.Context, f ListFilter) ([]Album, int, error)
	FindByID(ctx context.Context, id string) (*Album, error)
	FindHighlight(ctx context.Context) (*Album, error)
	Create(ctx context.Context, a *Album) error
	Update(ctx context.Context, a *Album) error
	Delete(ctx context.Context, id string) error
	SetHighlight(ctx context.Context, albumID string, highlight bool) error

	FindPhotosByAlbumID(ctx context.Context, albumID string) ([]Photo, error)
	FindPhotoByID(ctx context.Context, photoID string) (*Photo, error)
	CountPhotos(ctx context.Context, albumID string) (int, error)
	AddPhoto(ctx context.Context, p *Photo) error
	UpdatePhotoCaption(ctx context.Context, photoID string, caption *string) error
	DeletePhoto(ctx context.Context, photoID string) error

	GetMemberGeneration(ctx context.Context, memberID string) (int, error)
	GetMemberStatus(ctx context.Context, memberID string) (string, error)
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

const albumSelect = `
	SELECT
		a.id, a.title, a.description, a.event_date, a.korda_id, ko.name AS korda_name,
		a.kategori_id, kt.name AS kategori_name, a.is_highlight, a.visibility, a.target_generations,
		a.publish_at, a.expire_at, a.created_at, a.updated_at,
		COALESCE((SELECT COUNT(*) FROM gallery_photos p WHERE p.album_id = a.id), 0) AS photo_count,
		(SELECT p.image_url FROM gallery_photos p WHERE p.album_id = a.id ORDER BY p.sort_order ASC LIMIT 1) AS cover_image_url
	FROM gallery_albums a
	LEFT JOIN kordas ko ON ko.id = a.korda_id
	LEFT JOIN kategoris kt ON kt.id = a.kategori_id
`

const scheduleFilter = " AND (a.publish_at IS NULL OR a.publish_at <= now()) AND (a.expire_at IS NULL OR a.expire_at > now())"

func (r *repository) FindAll(ctx context.Context, f ListFilter) ([]Album, int, error) {
	where := " WHERE 1=1"
	args := []interface{}{}
	argPos := 1

	if f.Search != "" {
		where += " AND a.title ILIKE $" + strconv.Itoa(argPos)
		args = append(args, "%"+f.Search+"%")
		argPos++
	}
	if f.Visibility != "" {
		where += " AND a.visibility = $" + strconv.Itoa(argPos)
		args = append(args, f.Visibility)
		argPos++
	}
	if f.Scheduled {
		where += scheduleFilter
	}
	if f.MemberGeneration > 0 {
		where += " AND (a.target_generations IS NULL OR array_length(a.target_generations, 1) IS NULL OR $" + strconv.Itoa(argPos) + " = ANY(a.target_generations))"
		args = append(args, f.MemberGeneration)
		argPos++
	}

	var total int
	countQuery := "SELECT COUNT(*) FROM gallery_albums a" + where
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	limit := f.Limit
	offset := (f.Page - 1) * limit

	listQuery := albumSelect + where + " ORDER BY a.created_at DESC LIMIT $" + strconv.Itoa(argPos) + " OFFSET $" + strconv.Itoa(argPos+1)
	args = append(args, limit, offset)

	albums := []Album{}
	if err := r.db.SelectContext(ctx, &albums, listQuery, args...); err != nil {
		return nil, 0, err
	}

	return albums, total, nil
}

func (r *repository) FindByID(ctx context.Context, id string) (*Album, error) {
	var a Album
	err := r.db.GetContext(ctx, &a, albumSelect+" WHERE a.id = $1", id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &a, nil
}

func (r *repository) FindHighlight(ctx context.Context) (*Album, error) {
	var a Album
	err := r.db.GetContext(ctx, &a, albumSelect+" WHERE a.is_highlight = true LIMIT 1")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &a, nil
}

func (r *repository) Create(ctx context.Context, a *Album) error {
	query := `
		INSERT INTO gallery_albums (id, title, description, event_date, korda_id, kategori_id, visibility, target_generations, publish_at, expire_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING created_at, updated_at
	`
	return r.db.QueryRowContext(ctx, query,
		a.ID, a.Title, a.Description, a.EventDate, a.KordaID, a.KategoriID, a.Visibility,
		a.TargetGenerations, a.PublishAt, a.ExpireAt,
	).Scan(&a.CreatedAt, &a.UpdatedAt)
}

func (r *repository) Update(ctx context.Context, a *Album) error {
	query := `
		UPDATE gallery_albums
		SET title = $1, description = $2, event_date = $3, korda_id = $4, kategori_id = $5,
		    visibility = $6, target_generations = $7, publish_at = $8, expire_at = $9, updated_at = now()
		WHERE id = $10
		RETURNING updated_at
	`
	err := r.db.QueryRowContext(ctx, query,
		a.Title, a.Description, a.EventDate, a.KordaID, a.KategoriID, a.Visibility,
		a.TargetGenerations, a.PublishAt, a.ExpireAt, a.ID,
	).Scan(&a.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func (r *repository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM gallery_albums WHERE id = $1", id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *repository) SetHighlight(ctx context.Context, albumID string, highlight bool) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, "UPDATE gallery_albums SET is_highlight = false, updated_at = now() WHERE is_highlight = true"); err != nil {
		return err
	}

	if highlight {
		result, err := tx.ExecContext(ctx, "UPDATE gallery_albums SET is_highlight = true, updated_at = now() WHERE id = $1", albumID)
		if err != nil {
			return err
		}
		rows, _ := result.RowsAffected()
		if rows == 0 {
			return ErrNotFound
		}
	}

	return tx.Commit()
}

func (r *repository) FindPhotosByAlbumID(ctx context.Context, albumID string) ([]Photo, error) {
	photos := []Photo{}
	query := `
		SELECT id, album_id, image_url, caption, sort_order, created_at
		FROM gallery_photos WHERE album_id = $1 ORDER BY sort_order ASC
	`
	err := r.db.SelectContext(ctx, &photos, query, albumID)
	return photos, err
}

func (r *repository) FindPhotoByID(ctx context.Context, photoID string) (*Photo, error) {
	var p Photo
	query := `SELECT id, album_id, image_url, caption, sort_order, created_at FROM gallery_photos WHERE id = $1`
	err := r.db.GetContext(ctx, &p, query, photoID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPhotoNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (r *repository) CountPhotos(ctx context.Context, albumID string) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM gallery_photos WHERE album_id = $1", albumID)
	return count, err
}

func (r *repository) AddPhoto(ctx context.Context, p *Photo) error {
	query := `
		INSERT INTO gallery_photos (id, album_id, image_url, caption, sort_order)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at
	`
	return r.db.QueryRowContext(ctx, query, p.ID, p.AlbumID, p.ImageURL, p.Caption, p.SortOrder).Scan(&p.CreatedAt)
}

func (r *repository) UpdatePhotoCaption(ctx context.Context, photoID string, caption *string) error {
	result, err := r.db.ExecContext(ctx, "UPDATE gallery_photos SET caption = $1 WHERE id = $2", caption, photoID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrPhotoNotFound
	}
	return nil
}

func (r *repository) DeletePhoto(ctx context.Context, photoID string) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM gallery_photos WHERE id = $1", photoID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrPhotoNotFound
	}
	return nil
}

func (r *repository) GetMemberGeneration(ctx context.Context, memberID string) (int, error) {
	var generation int
	err := r.db.GetContext(ctx, &generation, "SELECT generation FROM members WHERE id = $1", memberID)
	return generation, err
}

func (r *repository) GetMemberStatus(ctx context.Context, memberID string) (string, error) {
	var status string
	err := r.db.GetContext(ctx, &status, "SELECT status FROM members WHERE id = $1", memberID)
	return status, err
}
