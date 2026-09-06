package membermanagement

import (
	"errors"
	"net/http"
	"strconv"

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

func (h *Handler) ListPending(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	result, err := h.service.ListPending(c.Request().Context(), page, limit)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
	return response.Success(c, http.StatusOK, "Berhasil mengambil data", result)
}

func (h *Handler) ListAll(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	result, err := h.service.ListAll(c.Request().Context(), page, limit)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
	return response.Success(c, http.StatusOK, "Berhasil mengambil data", result)
}

func (h *Handler) Approve(c echo.Context) error {
	id := c.Param("id")
	if err := h.service.Approve(c.Request().Context(), id); err != nil {
		return handleError(c, err)
	}
	return response.Success(c, http.StatusOK, "Anggota berhasil disetujui", nil)
}

func (h *Handler) Delete(c echo.Context) error {
	id := c.Param("id")
	if err := h.service.Delete(c.Request().Context(), id); err != nil {
		return handleError(c, err)
	}
	return response.Success(c, http.StatusOK, "Anggota berhasil dihapus", nil)
}

func handleError(c echo.Context, err error) error {
	if errors.Is(err, ErrNotFound) {
		return response.Error(c, http.StatusNotFound, "Anggota tidak ditemukan")
	}
	return response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
}

func RegisterRoutes(e *echo.Echo, h *Handler, jwtSecret string) {
	group := e.Group("/api/admin/members",
		appMiddleware.RequireAuth(jwtSecret),
		appMiddleware.RequireRole(auth.RoleAdmin, auth.RoleSuperAdmin),
	)
	group.GET("/pending", h.ListPending)
	group.GET("", h.ListAll)
	group.POST("/:id/approve", h.Approve)
	group.DELETE("/:id", h.Delete)
}
