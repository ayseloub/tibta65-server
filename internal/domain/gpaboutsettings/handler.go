package gpaboutsettings

import (
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

func (h *Handler) Get(c echo.Context) error {
	result, err := h.service.Get(c.Request().Context())
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
	return response.Success(c, http.StatusOK, "Berhasil mengambil data", result)
}

type updateRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Stat1       string `json:"stat_1"`
	Stat2       string `json:"stat_2"`
}

func (h *Handler) Update(c echo.Context) error {
	var req updateRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "Format request tidak valid")
	}

	result, err := h.service.Update(c.Request().Context(), UpdateInput{
		Title: req.Title, Description: req.Description, Stat1: req.Stat1, Stat2: req.Stat2,
	})
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
	return response.Success(c, http.StatusOK, "Pengaturan berhasil disimpan", result)
}

func RegisterRoutes(e *echo.Echo, h *Handler, jwtSecret string) {
	e.GET("/api/gp/about-settings", h.Get)

	group := e.Group("/api/admin/gp-about-settings",
		appMiddleware.RequireAuth(jwtSecret),
		appMiddleware.RequireRole(auth.RoleAdmin, auth.RoleSuperAdmin),
	)
	group.PUT("", h.Update)
	group.GET("", h.Get)
}
