package gpkategori

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
	items, err := h.service.List(c.Request().Context())
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
	return response.Success(c, http.StatusOK, "Berhasil mengambil data", items)
}

func (h *Handler) ListPublic(c echo.Context) error {
	items, err := h.service.List(c.Request().Context())
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
	return response.Success(c, http.StatusOK, "Berhasil mengambil data", items)
}

type kategoriRequest struct {
	Name string `json:"name"`
}

func (h *Handler) Create(c echo.Context) error {
	var req kategoriRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "Format request tidak valid")
	}
	k, err := h.service.Create(c.Request().Context(), req.Name)
	if err != nil {
		return handleError(c, err)
	}
	return response.Success(c, http.StatusCreated, "Kategori berhasil ditambahkan", k)
}

func (h *Handler) Update(c echo.Context) error {
	var req kategoriRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "Format request tidak valid")
	}
	k, err := h.service.Update(c.Request().Context(), c.Param("id"), req.Name)
	if err != nil {
		return handleError(c, err)
	}
	return response.Success(c, http.StatusOK, "Kategori berhasil diperbarui", k)
}

func (h *Handler) Delete(c echo.Context) error {
	if err := h.service.Delete(c.Request().Context(), c.Param("id")); err != nil {
		return handleError(c, err)
	}
	return response.Success(c, http.StatusOK, "Kategori berhasil dihapus", nil)
}

func handleError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, ErrValidation):
		return response.Error(c, http.StatusBadRequest, err.Error())
	case errors.Is(err, ErrNotFound):
		return response.Error(c, http.StatusNotFound, "Kategori tidak ditemukan")
	case errors.Is(err, ErrDuplicate):
		return response.Error(c, http.StatusConflict, err.Error())
	case errors.Is(err, ErrInUse):
		return response.Error(c, http.StatusConflict, err.Error())
	default:
		return response.Error(c, http.StatusBadRequest, err.Error())
	}
}

func RegisterRoutes(e *echo.Echo, h *Handler, jwtSecret string) {
	e.GET("/api/gp/kategori", h.ListPublic)

	group := e.Group("/api/admin/gp-kategori",
		appMiddleware.RequireAuth(jwtSecret),
		appMiddleware.RequireRole(auth.RoleAdmin, auth.RoleSuperAdmin),
	)
	group.GET("", h.ListAdmin)
	group.POST("", h.Create)
	group.PUT("/:id", h.Update)
	group.DELETE("/:id", h.Delete)
}
