package heroslide

import (
	"errors"
	"net/http"
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

func parseScheduleField(raw string) (*time.Time, error) {
	if raw == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02T15:04", raw)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (h *Handler) ListAdmin(c echo.Context) error {
	items, err := h.service.ListAdmin(c.Request().Context())
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
	return response.Success(c, http.StatusOK, "Berhasil mengambil data", items)
}

func (h *Handler) ListPublic(c echo.Context) error {
	items, err := h.service.ListPublic(c.Request().Context())
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
	return response.Success(c, http.StatusOK, "Berhasil mengambil data", items)
}

func (h *Handler) Get(c echo.Context) error {
	id := c.Param("id")
	result, err := h.service.Get(c.Request().Context(), id)
	if err != nil {
		return handleError(c, err)
	}
	return response.Success(c, http.StatusOK, "Berhasil mengambil data", result)
}

func (h *Handler) Create(c echo.Context) error {
	file, _ := c.FormFile("image")

	publishAt, err := parseScheduleField(c.FormValue("publish_at"))
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "Format waktu terbit tidak valid")
	}
	expireAt, err := parseScheduleField(c.FormValue("expire_at"))
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "Format waktu berakhir tidak valid")
	}

	in := CreateInput{
		Headline:            c.FormValue("headline"),
		Description:         c.FormValue("description"),
		PrimaryButtonText:   c.FormValue("primary_button_text"),
		PrimaryButtonLink:   c.FormValue("primary_button_link"),
		SecondaryButtonText: c.FormValue("secondary_button_text"),
		SecondaryButtonLink: c.FormValue("secondary_button_link"),
		BgColor:             c.FormValue("bg_color"),
		Status:              c.FormValue("status"),
		PublishAt:           publishAt,
		ExpireAt:            expireAt,
		Image:               file,
	}

	result, err := h.service.Create(c.Request().Context(), in)
	if err != nil {
		return handleError(c, err)
	}
	return response.Success(c, http.StatusCreated, "Slide berhasil ditambahkan", result)
}

func (h *Handler) Update(c echo.Context) error {
	id := c.Param("id")
	file, _ := c.FormFile("image")

	publishAt, err := parseScheduleField(c.FormValue("publish_at"))
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "Format waktu terbit tidak valid")
	}
	expireAt, err := parseScheduleField(c.FormValue("expire_at"))
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "Format waktu berakhir tidak valid")
	}

	in := UpdateInput{
		Headline:            c.FormValue("headline"),
		Description:         c.FormValue("description"),
		PrimaryButtonText:   c.FormValue("primary_button_text"),
		PrimaryButtonLink:   c.FormValue("primary_button_link"),
		SecondaryButtonText: c.FormValue("secondary_button_text"),
		SecondaryButtonLink: c.FormValue("secondary_button_link"),
		BgColor:             c.FormValue("bg_color"),
		Status:              c.FormValue("status"),
		PublishAt:           publishAt,
		ExpireAt:            expireAt,
		Image:               file,
	}

	result, err := h.service.Update(c.Request().Context(), id, in)
	if err != nil {
		return handleError(c, err)
	}
	return response.Success(c, http.StatusOK, "Slide berhasil diperbarui", result)
}

func (h *Handler) Delete(c echo.Context) error {
	id := c.Param("id")
	if err := h.service.Delete(c.Request().Context(), id); err != nil {
		return handleError(c, err)
	}
	return response.Success(c, http.StatusOK, "Slide berhasil dihapus", nil)
}

type reorderRequest struct {
	OrderedIDs []string `json:"ordered_ids"`
}

func (h *Handler) Reorder(c echo.Context) error {
	var req reorderRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "Format request tidak valid")
	}
	if err := h.service.Reorder(c.Request().Context(), req.OrderedIDs); err != nil {
		return handleError(c, err)
	}
	return response.Success(c, http.StatusOK, "Urutan slide berhasil disimpan", nil)
}

func handleError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, ErrValidation):
		return response.Error(c, http.StatusBadRequest, "Data tidak valid, cek kembali form")
	case errors.Is(err, ErrNotFound):
		return response.Error(c, http.StatusNotFound, "Slide tidak ditemukan")
	default:
		return response.Error(c, http.StatusBadRequest, err.Error())
	}
}

func RegisterRoutes(e *echo.Echo, h *Handler, jwtSecret string) {
	e.GET("/api/hero-slides", h.ListPublic)

	group := e.Group("/api/admin/hero-slides",
		appMiddleware.RequireAuth(jwtSecret),
		appMiddleware.RequireRole(auth.RoleAdmin, auth.RoleSuperAdmin),
	)
	group.GET("", h.ListAdmin)
	group.GET("/:id", h.Get)
	group.POST("", h.Create)
	group.PUT("/reorder", h.Reorder)
	group.PUT("/:id", h.Update)
	group.DELETE("/:id", h.Delete)
}
