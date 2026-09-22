package kegiatan

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

	f := ListFilter{
		Search:     c.QueryParam("search"),
		KordaID:    c.QueryParam("korda_id"),
		KategoriID: c.QueryParam("kategori_id"),
		Visibility: c.QueryParam("visibility"),
		Scheduled:  false,
		Page:       page,
		Limit:      limit,
	}

	result, err := h.service.List(c.Request().Context(), f)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
	return response.Success(c, http.StatusOK, "Berhasil mengambil data", result)
}

func (h *Handler) Get(c echo.Context) error {
	slugParam := c.Param("slug")
	k, err := h.service.Get(c.Request().Context(), slugParam)
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

	in := CreateInput{
		Title:       c.FormValue("title"),
		Date:        c.FormValue("date"),
		KordaID:     c.FormValue("korda_id"),
		KategoriID:  c.FormValue("kategori_id"),
		Location:    c.FormValue("location"),
		Description: c.FormValue("description"),
		Visibility:  c.FormValue("visibility"),
		PublishAt:   publishAt,
		ExpireAt:    expireAt,
		Image:       file,
	}

	k, err := h.service.Create(c.Request().Context(), in)
	if err != nil {
		return handleError(c, err)
	}
	return response.Success(c, http.StatusCreated, "Kegiatan berhasil ditambahkan", k)
}

func (h *Handler) Update(c echo.Context) error {
	slugParam := c.Param("slug")
	file, _ := c.FormFile("image")

	publishAt, err := parsePublishAt(c)
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "Format waktu terbit tidak valid")
	}
	expireAt, err := parseExpireAt(c)
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "Format waktu berakhir tidak valid")
	}

	in := UpdateInput{
		Slug:        slugParam,
		Title:       c.FormValue("title"),
		Date:        c.FormValue("date"),
		KordaID:     c.FormValue("korda_id"),
		KategoriID:  c.FormValue("kategori_id"),
		Location:    c.FormValue("location"),
		Description: c.FormValue("description"),
		Visibility:  c.FormValue("visibility"),
		PublishAt:   publishAt,
		ExpireAt:    expireAt,
		Image:       file,
	}

	k, err := h.service.Update(c.Request().Context(), in)
	if err != nil {
		return handleError(c, err)
	}
	return response.Success(c, http.StatusOK, "Kegiatan berhasil diperbarui", k)
}

func (h *Handler) Delete(c echo.Context) error {
	slugParam := c.Param("slug")
	if err := h.service.Delete(c.Request().Context(), slugParam); err != nil {
		return handleError(c, err)
	}
	return response.Success(c, http.StatusOK, "Kegiatan berhasil dihapus", nil)
}

func (h *Handler) ListPublic(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))

	f := ListFilter{
		Search:     c.QueryParam("search"),
		KordaID:    c.QueryParam("korda_id"),
		KategoriID: c.QueryParam("kategori_id"),
		Visibility: VisibilityPublic,
		Scheduled:  true,
		Page:       page,
		Limit:      limit,
	}

	result, err := h.service.List(c.Request().Context(), f)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
	return response.Success(c, http.StatusOK, "Berhasil mengambil data", result)
}

func (h *Handler) GetPublic(c echo.Context) error {
	slugParam := c.Param("slug")
	k, err := h.service.Get(c.Request().Context(), slugParam)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return response.Error(c, http.StatusNotFound, "Kegiatan tidak ditemukan")
		}
		return response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
	if k.Visibility != VisibilityPublic {
		return response.Error(c, http.StatusNotFound, "Kegiatan tidak ditemukan")
	}
	if k.PublishAt != nil && k.PublishAt.After(time.Now()) {
		return response.Error(c, http.StatusNotFound, "Kegiatan tidak ditemukan")
	}
	if k.ExpireAt != nil && !k.ExpireAt.After(time.Now()) {
		return response.Error(c, http.StatusNotFound, "Kegiatan tidak ditemukan")
	}
	return response.Success(c, http.StatusOK, "Berhasil mengambil data", k)
}

func (h *Handler) ListInternal(c echo.Context) error {
	memberID, _ := c.Get(appMiddleware.ContextKeyMemberID).(string)
	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))

	generation, err := h.service.GetMemberGeneration(c.Request().Context(), memberID)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
	}

	f := ListFilter{
		Visibility:       VisibilityInternal,
		Scheduled:        true,
		MemberGeneration: generation,
		Page:             page,
		Limit:            limit,
	}

	result, err := h.service.List(c.Request().Context(), f)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
	return response.Success(c, http.StatusOK, "Berhasil mengambil data", result)
}

func (h *Handler) GetInternal(c echo.Context) error {
	memberID, _ := c.Get(appMiddleware.ContextKeyMemberID).(string)
	slugParam := c.Param("slug")

	k, err := h.service.Get(c.Request().Context(), slugParam)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return response.Error(c, http.StatusNotFound, "Kegiatan tidak ditemukan")
		}
		return response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
	if k.Visibility != VisibilityInternal {
		return response.Error(c, http.StatusNotFound, "Kegiatan tidak ditemukan")
	}
	if k.PublishAt != nil && k.PublishAt.After(time.Now()) {
		return response.Error(c, http.StatusNotFound, "Kegiatan tidak ditemukan")
	}
	if k.ExpireAt != nil && !k.ExpireAt.After(time.Now()) {
		return response.Error(c, http.StatusNotFound, "Kegiatan tidak ditemukan")
	}

	if len(k.TargetGenerations) > 0 {
		generation, err := h.service.GetMemberGeneration(c.Request().Context(), memberID)
		if err != nil {
			return response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
		}
		eligible := false
		for _, g := range k.TargetGenerations {
			if int(g) == generation {
				eligible = true
				break
			}
		}
		if !eligible {
			return response.Error(c, http.StatusNotFound, "Kegiatan tidak ditemukan")
		}
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

func RegisterRoutes(e *echo.Echo, h *Handler, jwtSecret, memberJWTSecret string) {
	e.GET("/api/kegiatan", h.ListPublic)
	e.GET("/api/kegiatan/:slug", h.GetPublic)

	adminGroup := e.Group("/api/admin/kegiatan",
		appMiddleware.RequireAuth(jwtSecret),
		appMiddleware.RequireRole(auth.RoleAdmin, auth.RoleSuperAdmin),
	)
	adminGroup.GET("", h.List)
	adminGroup.GET("/:slug", h.Get)
	adminGroup.POST("", h.Create)
	adminGroup.PUT("/:slug", h.Update)
	adminGroup.DELETE("/:slug", h.Delete)

	memberGroup := e.Group("/api/member/kegiatan", appMiddleware.RequireMemberAuth(memberJWTSecret))
	memberGroup.GET("", h.ListInternal)
	memberGroup.GET("/:slug", h.GetInternal)
}
