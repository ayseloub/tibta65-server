package pemilu

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"

	appMiddleware "github.com/Tibta65web/tibta65-server/pkg/middleware"
	"github.com/Tibta65web/tibta65-server/pkg/response"
)

type VoteNumberHandler struct {
	db *sqlx.DB
}

func NewVoteNumberHandler(db *sqlx.DB) *VoteNumberHandler {
	return &VoteNumberHandler{db: db}
}

func (h *VoteNumberHandler) Get(c echo.Context) error {
	memberID, _ := c.Get(appMiddleware.ContextKeyMemberID).(string)

	var number int
	err := h.db.GetContext(c.Request().Context(), &number, "SELECT vote_number FROM votes WHERE member_id = $1", memberID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return response.Error(c, http.StatusNotFound, "Kamu belum memberikan suara")
		}
		return response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
	}

	return response.Success(c, http.StatusOK, "Berhasil mengambil data", map[string]int{"vote_number": number})
}

func RegisterVoteNumberRoutes(e *echo.Echo, h *VoteNumberHandler, memberJWTSecret string) {
	e.GET("/api/member/pemilu/vote-number", h.Get, appMiddleware.RequireMemberAuth(memberJWTSecret))
}
