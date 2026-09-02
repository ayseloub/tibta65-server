package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/Tibta65web/tibta65-server/pkg/jwt"
	"github.com/Tibta65web/tibta65-server/pkg/response"
)

const (
	ContextKeyAdminID = "admin_id"
	ContextKeyRole    = "role"
)

func RequireAuth(jwtSecret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return response.Error(c, http.StatusUnauthorized, "Token tidak ditemukan")
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				return response.Error(c, http.StatusUnauthorized, "Format token tidak valid")
			}

			claims, err := jwt.ParseToken(jwtSecret, parts[1])
			if err != nil {
				return response.Error(c, http.StatusUnauthorized, "Token tidak valid atau sudah kedaluwarsa")
			}

			c.Set(ContextKeyAdminID, claims.AdminID)
			c.Set(ContextKeyRole, claims.Role)

			return next(c)
		}
	}
}

func RequireRole(allowedRoles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			role, ok := c.Get(ContextKeyRole).(string)
			if !ok {
				return response.Error(c, http.StatusForbidden, "Akses ditolak")
			}

			for _, allowed := range allowedRoles {
				if role == allowed {
					return next(c)
				}
			}

			return response.Error(c, http.StatusForbidden, "Anda tidak punya izin untuk mengakses ini")
		}
	}
}
