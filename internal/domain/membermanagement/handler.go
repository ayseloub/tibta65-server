package membermanagement

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/Tibta65web/tibta65-server/internal/domain/auth"
	"github.com/Tibta65web/tibta65-server/internal/domain/member"
	appMiddleware "github.com/Tibta65web/tibta65-server/pkg/middleware"
	"github.com/Tibta65web/tibta65-server/pkg/response"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

type updateStatusRequest struct {
	Status string `json:"status"`
}

func (h *Handler) UpdateStatus(c echo.Context) error {
	id := c.Param("id")
	var req updateStatusRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "Format request tidak valid")
	}
	if err := h.service.UpdateStatus(c.Request().Context(), id, req.Status); err != nil {
		return handleError(c, err)
	}
	return response.Success(c, http.StatusOK, "Status berhasil diperbarui", nil)
}

func (h *Handler) ListPending(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	result, err := h.service.ListPending(c.Request().Context(), page, limit)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
	return response.Success(c, http.StatusOK, "Berhasil mengambil data", result)
}

func (h *Handler) ListAll(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	result, err := h.service.ListAll(c.Request().Context(), page, limit)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
	return response.Success(c, http.StatusOK, "Berhasil mengambil data", result)
}

func (h *Handler) Approve(c echo.Context) error {
	id := c.Param("id")
	if err := h.service.Approve(c.Request().Context(), id); err != nil {
		return handleError(c, err)
	}
	return response.Success(c, http.StatusOK, "Anggota berhasil disetujui", nil)
}

func (h *Handler) Delete(c echo.Context) error {
	id := c.Param("id")
	if err := h.service.Delete(c.Request().Context(), id); err != nil {
		return handleError(c, err)
	}
	return response.Success(c, http.StatusOK, "Anggota berhasil dihapus", nil)
}

type createLegacyMemberRequest struct {
	FullName        string `json:"full_name"`
	Generation      int    `json:"generation"`
	ParentMemberID  string `json:"parent_member_id"`
	NamaSuci        string `json:"nama_suci"`
	Agama           string `json:"agama"`
	NRP             string `json:"nrp"`
	NoAK            string `json:"no_ak"`
	PangkatTerakhir string `json:"pangkat_terakhir"`
}

func (h *Handler) CreateLegacyMember(c echo.Context) error {
	var req createLegacyMemberRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "Format request tidak valid")
	}

	m, err := h.service.CreateLegacyMember(c.Request().Context(), CreateLegacyMemberInput{
		FullName: req.FullName, Generation: req.Generation, ParentMemberID: req.ParentMemberID,
		NamaSuci: req.NamaSuci, Agama: req.Agama, NRP: req.NRP, NoAK: req.NoAK, PangkatTerakhir: req.PangkatTerakhir,
	})
	if err != nil {
		return handleError(c, err)
	}
	return response.Success(c, http.StatusCreated, "Anggota berhasil ditambahkan", m)
}

func handleError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, ErrNotFound):
		return response.Error(c, http.StatusNotFound, "Anggota tidak ditemukan")
	case errors.Is(err, member.ErrValidation):
		return response.Error(c, http.StatusBadRequest, "Nama lengkap wajib diisi")
	case errors.Is(err, member.ErrDuplicateMemberNumber):
		return response.Error(c, http.StatusConflict, err.Error())
	case errors.Is(err, ErrInvalidStatusTransition):
		return response.Error(c, http.StatusBadRequest, err.Error())
	default:
		return response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
}

func RegisterRoutes(e *echo.Echo, h *Handler, jwtSecret string) {
	group := e.Group("/api/admin/members",
		appMiddleware.RequireAuth(jwtSecret),
		appMiddleware.RequireRole(auth.RoleAdmin, auth.RoleSuperAdmin),
	)
	group.GET("/pending", h.ListPending)
	group.GET("", h.ListAll)
	group.POST("/legacy", h.CreateLegacyMember)
	group.POST("/:id/approve", h.Approve)
	group.DELETE("/:id", h.Delete)
	group.PUT("/:id/status", h.UpdateStatus)
}
