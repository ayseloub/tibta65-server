package ticket

import (
	"github.com/labstack/echo/v4"

	appMiddleware "github.com/Tibta65web/tibta65-server/pkg/middleware"
)

func RegisterRoutes(e *echo.Echo, h *Handler, memberJWTSecret, adminJWTSecret string) {
	memberGroup := e.Group("/api/member/tickets", appMiddleware.RequireMemberAuth(memberJWTSecret))
	memberGroup.POST("", h.Create)
	memberGroup.GET("", h.ListMine)
	memberGroup.GET("/:id", h.GetMine)

	adminGroup := e.Group("/api/admin/tickets", appMiddleware.RequireAuth(adminJWTSecret))
	adminGroup.GET("", h.ListAdmin)
	adminGroup.GET("/:id", h.GetAdmin)
	adminGroup.PUT("/:id/reply", h.Reply)
}
