package kategori

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

type upsertRequest struct {
	Name string `json:"name"`
}

func (h *Handler) List(c echo.Context) error {
	result, err := h.service.List(c.Request().Context())
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
	return response.Success(c, http.StatusOK, "Berhasil mengambil data", result)
}

func (h *Handler) Create(c echo.Context) error {
	var req upsertRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "Format request tidak valid")
	}

	result, err := h.service.Create(c.Request().Context(), req.Name)
	if err != nil {
		return handleError(c, err)
	}
	return response.Success(c, http.StatusCreated, "kategori berhasil ditambahkan", result)
}

func (h *Handler) Update(c echo.Context) error {
	id := c.Param("id")
	var req upsertRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "Format request tidak valid")
	}

	result, err := h.service.Update(c.Request().Context(), id, req.Name)
	if err != nil {
		return handleError(c, err)
	}
	return response.Success(c, http.StatusOK, "kategori berhasil diperbarui", result)
}

func (h *Handler) Delete(c echo.Context) error {
	id := c.Param("id")
	if err := h.service.Delete(c.Request().Context(), id); err != nil {
		return handleError(c, err)
	}
	return response.Success(c, http.StatusOK, "kategori berhasil dihapus", nil)
}

func handleError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, ErrValidation):
		return response.Error(c, http.StatusBadRequest, "Nama wajib diisi")
	case errors.Is(err, ErrDuplicate):
		return response.Error(c, http.StatusConflict, "Nama kategori sudah dipakai")
	case errors.Is(err, ErrNotFound):
		return response.Error(c, http.StatusNotFound, "kategori tidak ditemukan")
	case errors.Is(err, ErrInUse):
		return response.Error(c, http.StatusConflict, "kategori masih dipakai oleh kegiatan lain")
	default:
		return response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
}

func RegisterRoutes(e *echo.Echo, h *Handler, jwtSecret string) {
	e.GET("/api/kategoris", h.List)

	adminGroup := e.Group("/api/admin/kategoris",
		appMiddleware.RequireAuth(jwtSecret),
		appMiddleware.RequireRole(auth.RoleAdmin, auth.RoleSuperAdmin),
	)
	adminGroup.POST("", h.Create)
	adminGroup.PUT("/:id", h.Update)
	adminGroup.DELETE("/:id", h.Delete)
}
