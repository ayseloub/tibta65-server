package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/Tibta65web/tibta65-server/pkg/jwt"
)

const ContextKeyMemberID = "member_id"

func RequireMemberAuth(jwtSecret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				return echo.NewHTTPError(http.StatusUnauthorized, "Token tidak ditemukan")
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			claims, err := jwt.ParseMemberToken(jwtSecret, tokenString)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "Token tidak valid atau kedaluwarsa")
			}

			c.Set(ContextKeyMemberID, claims.MemberID)
			return next(c)
		}
	}
}
