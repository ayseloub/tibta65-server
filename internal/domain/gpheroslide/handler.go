package gpheroslide

import (
	"errors"
	"net/http"

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

func (h *Handler) Create(c echo.Context) error {
	file, _ := c.FormFile("image")
	sl, err := h.service.Create(c.Request().Context(), CreateInput{
		Headline: c.FormValue("headline"), Description: c.FormValue("description"),
		PrimaryButtonText: c.FormValue("primary_button_text"), PrimaryButtonLink: c.FormValue("primary_button_link"),
		SecondaryButtonText: c.FormValue("secondary_button_text"), SecondaryButtonLink: c.FormValue("secondary_button_link"),
		Status: c.FormValue("status"), Image: file,
	})
	if err != nil {
		return handleError(c, err)
	}
	return response.Success(c, http.StatusCreated, "Slide berhasil ditambahkan", sl)
}

func (h *Handler) Update(c echo.Context) error {
	file, _ := c.FormFile("image")
	sl, err := h.service.Update(c.Request().Context(), UpdateInput{
		ID: c.Param("id"), Headline: c.FormValue("headline"), Description: c.FormValue("description"),
		PrimaryButtonText: c.FormValue("primary_button_text"), PrimaryButtonLink: c.FormValue("primary_button_link"),
		SecondaryButtonText: c.FormValue("secondary_button_text"), SecondaryButtonLink: c.FormValue("secondary_button_link"),
		Status: c.FormValue("status"), Image: file,
	})
	if err != nil {
		return handleError(c, err)
	}
	return response.Success(c, http.StatusOK, "Slide berhasil diperbarui", sl)
}

func (h *Handler) Delete(c echo.Context) error {
	if err := h.service.Delete(c.Request().Context(), c.Param("id")); err != nil {
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
		return response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
	return response.Success(c, http.StatusOK, "Urutan berhasil disimpan", nil)
}

func handleError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, ErrValidation):
		return response.Error(c, http.StatusBadRequest, "Judul dan gambar wajib diisi")
	case errors.Is(err, ErrNotFound):
		return response.Error(c, http.StatusNotFound, "Slide tidak ditemukan")
	default:
		return response.Error(c, http.StatusBadRequest, err.Error())
	}
}

func RegisterRoutes(e *echo.Echo, h *Handler, jwtSecret string) {
	e.GET("/api/gp/hero-slides", h.ListPublic)

	group := e.Group("/api/admin/gp-hero-slides",
		appMiddleware.RequireAuth(jwtSecret),
		appMiddleware.RequireRole(auth.RoleAdmin, auth.RoleSuperAdmin),
	)
	group.GET("", h.ListAdmin)
	group.POST("", h.Create)
	group.PUT("/:id", h.Update)
	group.DELETE("/:id", h.Delete)
	group.PUT("/reorder", h.Reorder)
}
