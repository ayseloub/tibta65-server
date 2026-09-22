package berita

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"

	appMiddleware "github.com/Tibta65web/tibta65-server/pkg/middleware"
	"github.com/Tibta65web/tibta65-server/pkg/response"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func parseEventDate(c echo.Context) (time.Time, error) {
	return time.Parse("2006-01-02", c.FormValue("event_date"))
}

func parsePublishAt(c echo.Context) (*time.Time, error) {
	raw := c.FormValue("publish_at")
	if raw == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02T15:04", raw)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func parseExpireAt(c echo.Context) (*time.Time, error) {
	raw := c.FormValue("expire_at")
	if raw == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02T15:04", raw)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (h *Handler) Create(c echo.Context) error {
	adminID, _ := c.Get(appMiddleware.ContextKeyAdminID).(string)

	eventDate, err := parseEventDate(c)
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "Format tanggal tidak valid")
	}

	publishAt, err := parsePublishAt(c)
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "Format waktu terbit tidak valid")
	}

	expireAt, err := parseExpireAt(c)
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "Format waktu berakhir tidak valid")
	}

	fileHeader, err := c.FormFile("image")
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "Gambar wajib diupload")
	}

	result, err := h.service.Create(c.Request().Context(), CreateInput{
		Title:       c.FormValue("title"),
		Description: c.FormValue("description"),
		Visibility:  c.FormValue("visibility"),
		Status:      c.FormValue("status"),
		EventDate:   eventDate,
		PublishAt:   publishAt,
		ExpireAt:    expireAt,
		ImageHeader: fileHeader,
		AuthorID:    adminID,
		AuthorName:  c.FormValue("author_name"),
	})
	if err != nil {
		if errors.Is(err, ErrValidation) {
			return response.Error(c, http.StatusBadRequest, "Data tidak valid, cek kembali form")
		}
		return response.Error(c, http.StatusInternalServerError, err.Error())
	}
	return response.Success(c, http.StatusCreated, "Berita berhasil dibuat", result)
}

func (h *Handler) ListMember(c echo.Context) error {
	memberID, _ := c.Get(appMiddleware.ContextKeyMemberID).(string)
	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	generation, err := h.service.GetMemberGeneration(c.Request().Context(), memberID)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
	}

	items, total, err := h.service.FindAllMember(c.Request().Context(), page, limit, generation)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
	return response.Success(c, http.StatusOK, "Berhasil mengambil data", map[string]interface{}{
		"items": items, "page": page, "limit": limit, "total": total,
		"total_pages": (total + limit - 1) / limit,
	})
}

func (h *Handler) GetMember(c echo.Context) error {
	memberID, _ := c.Get(appMiddleware.ContextKeyMemberID).(string)
	slug := c.Param("slug")

	result, err := h.service.FindBySlugMember(c.Request().Context(), slug)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return response.Error(c, http.StatusNotFound, "Berita tidak ditemukan")
		}
		return response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
	}

	if len(result.TargetGenerations) > 0 {
		generation, err := h.service.GetMemberGeneration(c.Request().Context(), memberID)
		if err != nil {
			return response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
		}
		eligible := false
		for _, g := range result.TargetGenerations {
			if int(g) == generation {
				eligible = true
				break
			}
		}
		if !eligible {
			return response.Error(c, http.StatusNotFound, "Berita tidak ditemukan")
		}
	}

	return response.Success(c, http.StatusOK, "Berhasil mengambil data", result)
}

func (h *Handler) Update(c echo.Context) error {
	id := c.Param("id")

	eventDate, err := parseEventDate(c)
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "Format tanggal tidak valid")
	}

	publishAt, err := parsePublishAt(c)
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "Format waktu terbit tidak valid")
	}

	expireAt, err := parseExpireAt(c)
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "Format waktu berakhir tidak valid")
	}

	fileHeader, _ := c.FormFile("image")

	result, err := h.service.Update(c.Request().Context(), id, UpdateInput{
		Title:       c.FormValue("title"),
		Description: c.FormValue("description"),
		Visibility:  c.FormValue("visibility"),
		Status:      c.FormValue("status"),
		EventDate:   eventDate,
		PublishAt:   publishAt,
		ExpireAt:    expireAt,
		ImageHeader: fileHeader,
	})
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return response.Error(c, http.StatusNotFound, "Berita tidak ditemukan")
		}
		if errors.Is(err, ErrValidation) {
			return response.Error(c, http.StatusBadRequest, "Data tidak valid, cek kembali form")
		}
		return response.Error(c, http.StatusInternalServerError, err.Error())
	}
	return response.Success(c, http.StatusOK, "Berita berhasil diperbarui", result)
}

func (h *Handler) Delete(c echo.Context) error {
	id := c.Param("id")
	if err := h.service.Delete(c.Request().Context(), id); err != nil {
		if errors.Is(err, ErrNotFound) {
			return response.Error(c, http.StatusNotFound, "Berita tidak ditemukan")
		}
		return response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
	return response.Success(c, http.StatusOK, "Berita berhasil dihapus", nil)
}

func (h *Handler) GetByID(c echo.Context) error {
	id := c.Param("id")
	result, err := h.service.FindByID(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return response.Error(c, http.StatusNotFound, "Berita tidak ditemukan")
		}
		return response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
	return response.Success(c, http.StatusOK, "Berhasil mengambil data", result)
}

func (h *Handler) List(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	items, total, err := h.service.FindAll(c.Request().Context(), ListFilter{
		Search:     c.QueryParam("search"),
		Status:     c.QueryParam("status"),
		Visibility: c.QueryParam("visibility"),
		Page:       page,
		Limit:      limit,
	})
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
	}

	return response.Success(c, http.StatusOK, "Berhasil mengambil data", map[string]interface{}{
		"items": items, "page": page, "limit": limit, "total": total,
		"total_pages": (total + limit - 1) / limit,
	})
}

func (h *Handler) Stats(c echo.Context) error {
	total, published, draft, err := h.service.Stats(c.Request().Context())
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
	return response.Success(c, http.StatusOK, "Berhasil mengambil data", map[string]interface{}{
		"total": total, "published": published, "draft": draft,
	})
}

type toggleHighlightRequest struct {
	Enable bool `json:"enable"`
}

func (h *Handler) ToggleHighlight(c echo.Context) error {
	id := c.Param("id")
	var req toggleHighlightRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "Format request tidak valid")
	}

	result, err := h.service.ToggleHighlight(c.Request().Context(), id, req.Enable)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return response.Error(c, http.StatusNotFound, "Berita tidak ditemukan")
		}
		return response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
	return response.Success(c, http.StatusOK, "Highlight berhasil diperbarui", result)
}

func (h *Handler) GetHighlightPublic(c echo.Context) error {
	result, err := h.service.FindHighlightPublic(c.Request().Context())
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
	return response.Success(c, http.StatusOK, "Berhasil mengambil data", result)
}

func (h *Handler) ListPublic(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 6
	}

	items, total, err := h.service.FindAllPublic(c.Request().Context(), page, limit)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
	return response.Success(c, http.StatusOK, "Berhasil mengambil data", map[string]interface{}{
		"items": items, "page": page, "limit": limit, "total": total,
		"total_pages": (total + limit - 1) / limit,
	})
}

func (h *Handler) GetBySlugPublic(c echo.Context) error {
	slug := c.Param("slug")
	result, err := h.service.FindBySlugPublic(c.Request().Context(), slug)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return response.Error(c, http.StatusNotFound, "Berita tidak ditemukan")
		}
		return response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
	return response.Success(c, http.StatusOK, "Berhasil mengambil data", result)
}
