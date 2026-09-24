package gpkegiatan

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/Tibta65web/tibta65-server/internal/domain/auth"
	appMiddleware "github.com/Tibta65web/tibta65-server/pkg/middleware"
	"github.com/Tibta65web/tibta65-server/pkg/response"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func parsePublishAt(c echo.Context) (*time.Time, error) {
	raw := c.FormValue("publish_at")
	if raw == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, raw)
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
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (h *Handler) List(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	result, err := h.service.List(c.Request().Context(), ListFilter{
		KordaID: c.QueryParam("korda_id"), Page: page, Limit: limit,
	})
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
	return response.Success(c, http.StatusOK, "Berhasil mengambil data", result)
}

func (h *Handler) Get(c echo.Context) error {
	k, err := h.service.Get(c.Request().Context(), c.Param("slug"))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return response.Error(c, http.StatusNotFound, "Kegiatan tidak ditemukan")
		}
		return response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
	return response.Success(c, http.StatusOK, "Berhasil mengambil data", k)
}

func (h *Handler) Create(c echo.Context) error {
	file, _ := c.FormFile("image")
	publishAt, err := parsePublishAt(c)
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "Format waktu terbit tidak valid")
	}
	expireAt, err := parseExpireAt(c)
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "Format waktu berakhir tidak valid")
	}

	k, err := h.service.Create(c.Request().Context(), CreateInput{
		Title: c.FormValue("title"), Date: c.FormValue("date"), KordaID: c.FormValue("korda_id"),
		Location: c.FormValue("location"), Description: c.FormValue("description"),
		PublishAt: publishAt, ExpireAt: expireAt, Image: file,
	})
	if err != nil {
		return handleError(c, err)
	}
	return response.Success(c, http.StatusCreated, "Kegiatan GP TIBTA berhasil ditambahkan", k)
}

func (h *Handler) Update(c echo.Context) error {
	file, _ := c.FormFile("image")
	publishAt, err := parsePublishAt(c)
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "Format waktu terbit tidak valid")
	}
	expireAt, err := parseExpireAt(c)
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "Format waktu berakhir tidak valid")
	}

	k, err := h.service.Update(c.Request().Context(), UpdateInput{
		Slug: c.Param("slug"), Title: c.FormValue("title"), Date: c.FormValue("date"),
		KordaID: c.FormValue("korda_id"), Location: c.FormValue("location"), Description: c.FormValue("description"),
		PublishAt: publishAt, ExpireAt: expireAt, Image: file,
	})
	if err != nil {
		return handleError(c, err)
	}
	return response.Success(c, http.StatusOK, "Kegiatan GP TIBTA berhasil diperbarui", k)
}

func (h *Handler) Delete(c echo.Context) error {
	if err := h.service.Delete(c.Request().Context(), c.Param("slug")); err != nil {
		return handleError(c, err)
	}
	return response.Success(c, http.StatusOK, "Kegiatan GP TIBTA berhasil dihapus", nil)
}

func (h *Handler) ListPublic(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	result, err := h.service.List(c.Request().Context(), ListFilter{
		KordaID: c.QueryParam("korda_id"), KategoriID: c.QueryParam("kategori_id"),
		Scheduled: true, Page: page, Limit: limit,
	})
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
	return response.Success(c, http.StatusOK, "Berhasil mengambil data", result)
}

func (h *Handler) GetPublic(c echo.Context) error {
	k, err := h.service.Get(c.Request().Context(), c.Param("slug"))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return response.Error(c, http.StatusNotFound, "Kegiatan tidak ditemukan")
		}
		return response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
	if k.PublishAt != nil && k.PublishAt.After(time.Now()) {
		return response.Error(c, http.StatusNotFound, "Kegiatan tidak ditemukan")
	}
	if k.ExpireAt != nil && !k.ExpireAt.After(time.Now()) {
		return response.Error(c, http.StatusNotFound, "Kegiatan tidak ditemukan")
	}
	return response.Success(c, http.StatusOK, "Berhasil mengambil data", k)
}

func handleError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, ErrValidation):
		return response.Error(c, http.StatusBadRequest, "Semua field wajib diisi")
	case errors.Is(err, ErrNotFound):
		return response.Error(c, http.StatusNotFound, "Kegiatan tidak ditemukan")
	case errors.Is(err, ErrDuplicate):
		return response.Error(c, http.StatusConflict, err.Error())
	default:
		return response.Error(c, http.StatusBadRequest, err.Error())
	}
}

func RegisterRoutes(e *echo.Echo, h *Handler, jwtSecret string) {
	e.GET("/api/gp/kegiatan", h.ListPublic)
	e.GET("/api/gp/kegiatan/:slug", h.GetPublic)

	adminGroup := e.Group("/api/admin/gp-kegiatan",
		appMiddleware.RequireAuth(jwtSecret),
		appMiddleware.RequireRole(auth.RoleAdmin, auth.RoleSuperAdmin),
	)
	adminGroup.GET("", h.List)
	adminGroup.GET("/:slug", h.Get)
	adminGroup.POST("", h.Create)
	adminGroup.PUT("/:slug", h.Update)
	adminGroup.DELETE("/:slug", h.Delete)
}
