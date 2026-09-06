package sitesettings

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
	Email        string `json:"email"`
	Phone        string `json:"phone"`
	YoutubeURL   string `json:"youtube_url"`
	InstagramURL string `json:"instagram_url"`
	FacebookURL  string `json:"facebook_url"`
	Address      string `json:"address"`
	MapsEmbedURL string `json:"maps_embed_url"`
	MapsLink     string `json:"maps_link"`
}

func (h *Handler) Update(c echo.Context) error {
	var req updateRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "Format request tidak valid")
	}

	result, err := h.service.Update(c.Request().Context(), UpdateInput{
		Email: req.Email, Phone: req.Phone, YoutubeURL: req.YoutubeURL, InstagramURL: req.InstagramURL,
		FacebookURL: req.FacebookURL, Address: req.Address, MapsEmbedURL: req.MapsEmbedURL, MapsLink: req.MapsLink,
	})
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
	return response.Success(c, http.StatusOK, "Pengaturan berhasil disimpan", result)
}

func RegisterRoutes(e *echo.Echo, h *Handler, jwtSecret string) {
	e.GET("/api/site-settings", h.Get)

	group := e.Group("/api/admin/site-settings",
		appMiddleware.RequireAuth(jwtSecret),
		appMiddleware.RequireRole(auth.RoleSuperAdmin),
	)
	group.PUT("", h.Update)
}
