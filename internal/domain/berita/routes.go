package berita

import (
	"github.com/labstack/echo/v4"

	appMiddleware "github.com/Tibta65web/tibta65-server/pkg/middleware"
)

func RegisterRoutes(e *echo.Echo, h *Handler, jwtSecret string) {
	admin := e.Group("/api/admin/berita", appMiddleware.RequireAuth(jwtSecret))
	admin.GET("/stats", h.Stats)
	admin.POST("", h.Create)
	admin.PUT("/:id", h.Update)
	admin.PUT("/:id/highlight", h.ToggleHighlight)
	admin.DELETE("/:id", h.Delete)
	admin.GET("/:id", h.GetByID)
	admin.GET("", h.List)

	public := e.Group("/api/public/berita")
	public.GET("/highlight", h.GetHighlightPublic)
	public.GET("/:slug", h.GetBySlugPublic)
	public.GET("", h.ListPublic)
}
