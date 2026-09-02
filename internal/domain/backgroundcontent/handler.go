package backgroundcontent

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

type updateRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// Get godoc
// @Summary      Ambil konten background (about/sejarah)
// @Description  Endpoint publik, dipakai landing page
// @Tags         background-content
// @Produce      json
// @Param        section path string true "about atau sejarah"
// @Success      200 {object} response.Envelope
// @Failure      400 {object} response.Envelope
// @Failure      404 {object} response.Envelope
// @Router       /api/background-content/{section} [get]
func (h *Handler) Get(c echo.Context) error {
	section := c.Param("section")

	content, err := h.service.Get(c.Request().Context(), section)
	if err != nil {
		if errors.Is(err, ErrInvalidSection) {
			return response.Error(c, http.StatusBadRequest, "Section tidak valid")
		}
		if errors.Is(err, ErrContentNotFound) {
			return response.Error(c, http.StatusNotFound, "Konten tidak ditemukan")
		}
		return response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
	}

	return response.Success(c, http.StatusOK, "Berhasil mengambil konten", content)
}

// Update godoc
// @Summary      Update konten background (about/sejarah)
// @Description  Endpoint admin, butuh token
// @Tags         background-content
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        section path string true "about atau sejarah"
// @Param        request body updateRequest true "Judul dan deskripsi baru"
// @Success      200 {object} response.Envelope
// @Failure      400 {object} response.Envelope
// @Failure      401 {object} response.Envelope
// @Router       /api/admin/background-content/{section} [put]
func (h *Handler) Update(c echo.Context) error {
	section := c.Param("section")

	var req updateRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "Format request tidak valid")
	}

	content, err := h.service.Update(c.Request().Context(), section, req.Title, req.Description)
	if err != nil {
		if errors.Is(err, ErrInvalidSection) {
			return response.Error(c, http.StatusBadRequest, "Section tidak valid")
		}
		if errors.Is(err, ErrContentNotFound) {
			return response.Error(c, http.StatusNotFound, "Konten tidak ditemukan")
		}
		return response.Error(c, http.StatusBadRequest, err.Error())
	}

	return response.Success(c, http.StatusOK, "Berhasil memperbarui konten", content)
}

// RegisterRoutes daftarin route publik (GET) dan route admin (PUT, protected).
func RegisterRoutes(e *echo.Echo, h *Handler, jwtSecret string) {
	e.GET("/api/background-content/:section", h.Get)

	adminGroup := e.Group("/api/admin/background-content",
		appMiddleware.RequireAuth(jwtSecret),
		appMiddleware.RequireRole(auth.RoleAdmin, auth.RoleSuperAdmin),
	)
	adminGroup.PUT("/:section", h.Update)
}
