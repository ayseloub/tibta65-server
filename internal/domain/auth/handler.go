package auth

import (
	"errors"
	"net/http"

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

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Login godoc
// @Summary      Login admin
// @Description  Login menggunakan username dan password, mengembalikan JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body loginRequest true "Kredensial login"
// @Success      200 {object} response.Envelope
// @Failure      400 {object} response.Envelope
// @Failure      401 {object} response.Envelope
// @Router       /api/auth/login [post]
func (h *Handler) Login(c echo.Context) error {
	var req loginRequest

	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "Format request tidak valid")
	}

	if req.Username == "" || req.Password == "" {
		return response.Error(c, http.StatusBadRequest, "Username dan password wajib diisi")
	}

	result, err := h.service.Login(c.Request().Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			return response.Error(c, http.StatusUnauthorized, "Username atau password salah")
		}
		return response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
	}

	return response.Success(c, http.StatusOK, "Login berhasil", result)
}

// Me godoc
// @Summary      Ambil profil admin yang sedang login
// @Tags         auth
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} response.Envelope
// @Failure      401 {object} response.Envelope
// @Router       /api/auth/me [get]
func (h *Handler) Me(c echo.Context) error {
	adminID, _ := c.Get(appMiddleware.ContextKeyAdminID).(string)

	admin, err := h.service.Me(c.Request().Context(), adminID)
	if err != nil {
		return response.Error(c, http.StatusNotFound, "Admin tidak ditemukan")
	}

	me := map[string]interface{}{
		"id":       admin.ID,
		"username": admin.Username,
		"role":     admin.Role,
	}

	return response.Success(c, http.StatusOK, "Berhasil mengambil profil", me)
}

func RegisterRoutes(e *echo.Echo, h *Handler, jwtSecret string) {
	group := e.Group("/api/auth")
	group.POST("/login", h.Login)

	protected := group.Group("", appMiddleware.RequireAuth(jwtSecret))
	protected.GET("/me", h.Me)
}
