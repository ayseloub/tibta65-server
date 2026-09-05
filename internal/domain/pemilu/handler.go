package pemilu

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

type settingsRequest struct {
	StartAt string `json:"start_at"`
	EndAt   string `json:"end_at"`
}

type kandidatRequest struct {
	FullName string `json:"full_name"`
	Visi     string `json:"visi"`
	Misi     string `json:"misi"`
	Pangkat  string `json:"pangkat"`
}

func (h *Handler) MemberDashboard(c echo.Context) error {
	memberID, _ := c.Get(appMiddleware.ContextKeyMemberID).(string)
	result, err := h.service.MemberDashboard(c.Request().Context(), memberID)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
	return response.Success(c, http.StatusOK, "Berhasil mengambil data", result)
}

type voteRequest struct {
	KandidatID string `json:"kandidat_id"`
}

func (h *Handler) CastVote(c echo.Context) error {
	memberID, _ := c.Get(appMiddleware.ContextKeyMemberID).(string)
	var req voteRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "Format request tidak valid")
	}

	if err := h.service.CastVote(c.Request().Context(), memberID, req.KandidatID); err != nil {
		switch {
		case errors.Is(err, ErrAlreadyVoted):
			return response.Error(c, http.StatusConflict, err.Error())
		case errors.Is(err, ErrPemiluNotActive):
			return response.Error(c, http.StatusForbidden, err.Error())
		case errors.Is(err, ErrKandidatNotFound):
			return response.Error(c, http.StatusNotFound, err.Error())
		default:
			return response.Error(c, http.StatusBadRequest, err.Error())
		}
	}
	return response.Success(c, http.StatusOK, "Suara berhasil dikirim", nil)
}

func (h *Handler) Dashboard(c echo.Context) error {
	result, err := h.service.Dashboard(c.Request().Context())
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
	return response.Success(c, http.StatusOK, "Berhasil mengambil data", result)
}

func (h *Handler) UpdateSettings(c echo.Context) error {
	var req settingsRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "Format request tidak valid")
	}
	result, err := h.service.UpdateSettings(c.Request().Context(), req.StartAt, req.EndAt)
	if err != nil {
		return handleError(c, err)
	}
	return response.Success(c, http.StatusOK, "Periode pemilu berhasil diperbarui", result)
}

func (h *Handler) CloseEarly(c echo.Context) error {
	result, err := h.service.CloseEarly(c.Request().Context())
	if err != nil {
		return handleError(c, err)
	}
	return response.Success(c, http.StatusOK, "Pemilu berhasil ditutup", result)
}

func (h *Handler) GetKandidat(c echo.Context) error {
	id := c.Param("id")
	k, err := h.service.GetKandidat(c.Request().Context(), id)
	if err != nil {
		return handleError(c, err)
	}
	return response.Success(c, http.StatusOK, "Berhasil mengambil data", k)
}

func (h *Handler) CreateKandidat(c echo.Context) error {
	var req kandidatRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "Format request tidak valid")
	}
	k, err := h.service.CreateKandidat(c.Request().Context(), KandidatInput{
		FullName: req.FullName, Visi: req.Visi, Misi: req.Misi, Pangkat: req.Pangkat,
	})
	if err != nil {
		return handleError(c, err)
	}
	return response.Success(c, http.StatusCreated, "Kandidat berhasil ditambahkan", k)
}

func (h *Handler) UpdateKandidat(c echo.Context) error {
	id := c.Param("id")
	var req kandidatRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "Format request tidak valid")
	}
	k, err := h.service.UpdateKandidat(c.Request().Context(), id, KandidatInput{
		FullName: req.FullName, Visi: req.Visi, Misi: req.Misi, Pangkat: req.Pangkat,
	})
	if err != nil {
		return handleError(c, err)
	}
	return response.Success(c, http.StatusOK, "Kandidat berhasil diperbarui", k)
}

func (h *Handler) DeleteKandidat(c echo.Context) error {
	id := c.Param("id")
	if err := h.service.DeleteKandidat(c.Request().Context(), id); err != nil {
		return handleError(c, err)
	}
	return response.Success(c, http.StatusOK, "Kandidat berhasil dihapus", nil)
}

func handleError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, ErrValidation):
		return response.Error(c, http.StatusBadRequest, "Semua field wajib diisi")
	case errors.Is(err, ErrNotFound):
		return response.Error(c, http.StatusNotFound, "Kandidat tidak ditemukan")
	default:
		return response.Error(c, http.StatusBadRequest, err.Error())
	}
}

func RegisterRoutes(e *echo.Echo, h *Handler, jwtSecret, memberJWTSecret string) {
	group := e.Group("/api/admin/pemilu",
		appMiddleware.RequireAuth(jwtSecret),
		appMiddleware.RequireRole(auth.RoleAdmin, auth.RoleSuperAdmin),
	)
	group.GET("/dashboard", h.Dashboard)
	group.PUT("/settings", h.UpdateSettings)
	group.POST("/close-early", h.CloseEarly)
	group.GET("/kandidat/:id", h.GetKandidat)
	group.POST("/kandidat", h.CreateKandidat)
	group.PUT("/kandidat/:id", h.UpdateKandidat)
	group.DELETE("/kandidat/:id", h.DeleteKandidat)

	memberGroup := e.Group("/api/member-auth/pemilu", appMiddleware.RequireMemberAuth(memberJWTSecret))
	memberGroup.GET("", h.MemberDashboard)
	memberGroup.POST("/vote", h.CastVote)
}
