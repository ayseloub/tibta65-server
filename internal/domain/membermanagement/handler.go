package membermanagement

import (
	"errors"
	"net/http"
	"strconv"
	"time"

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

func parseFilter(c echo.Context) ListFilter {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	generation, _ := strconv.Atoi(c.QueryParam("generation"))
	return ListFilter{
		Search:     c.QueryParam("search"),
		KordaID:    c.QueryParam("korda_id"),
		Generation: generation,
		Status:     c.QueryParam("status"),
		Page:       page,
		Limit:      limit,
	}
}

func (h *Handler) ListPending(c echo.Context) error {
	result, err := h.service.ListPending(c.Request().Context(), parseFilter(c))
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
	return response.Success(c, http.StatusOK, "Berhasil mengambil data", result)
}

func (h *Handler) ListAll(c echo.Context) error {
	result, err := h.service.ListAll(c.Request().Context(), parseFilter(c))
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
	return response.Success(c, http.StatusOK, "Berhasil mengambil data", result)
}

func (h *Handler) Approve(c echo.Context) error {
	if err := h.service.Approve(c.Request().Context(), c.Param("id")); err != nil {
		return handleError(c, err)
	}
	return response.Success(c, http.StatusOK, "Anggota berhasil disetujui", nil)
}

type rejectRequest struct {
	Type   string `json:"type"`
	Reason string `json:"reason"`
}

func (h *Handler) Reject(c echo.Context) error {
	var req rejectRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "Format request tidak valid")
	}
	if err := h.service.Reject(c.Request().Context(), c.Param("id"), req.Type, req.Reason); err != nil {
		return handleError(c, err)
	}
	return response.Success(c, http.StatusOK, "Pendaftaran berhasil ditolak", nil)
}

type reopenRequest struct {
	Reason string `json:"reason"`
}

func (h *Handler) Reopen(c echo.Context) error {
	var req reopenRequest
	_ = c.Bind(&req)
	if err := h.service.Reopen(c.Request().Context(), c.Param("id"), req.Reason); err != nil {
		return handleError(c, err)
	}
	return response.Success(c, http.StatusOK, "Akun diberi kesempatan untuk memperbaiki biodata", nil)
}

func (h *Handler) ApproveRejected(c echo.Context) error {
	if err := h.service.ApproveRejected(c.Request().Context(), c.Param("id")); err != nil {
		return handleError(c, err)
	}
	return response.Success(c, http.StatusOK, "Akun berhasil diaktifkan", nil)
}

func (h *Handler) Delete(c echo.Context) error {
	if err := h.service.Delete(c.Request().Context(), c.Param("id")); err != nil {
		return handleError(c, err)
	}
	return response.Success(c, http.StatusOK, "Anggota berhasil dihapus", nil)
}

type updateStatusRequest struct {
	Status string `json:"status"`
}

func (h *Handler) UpdateStatus(c echo.Context) error {
	var req updateStatusRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "Format request tidak valid")
	}
	if err := h.service.UpdateStatus(c.Request().Context(), c.Param("id"), req.Status); err != nil {
		return handleError(c, err)
	}
	return response.Success(c, http.StatusOK, "Status berhasil diperbarui", nil)
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

func (h *Handler) ExportFields(c echo.Context) error {
	return response.Success(c, http.StatusOK, "Berhasil mengambil data", h.service.ExportFields())
}

type exportRequest struct {
	MemberIDs  []string `json:"member_ids"`
	Search     string   `json:"search"`
	KordaID    string   `json:"korda_id"`
	Generation int      `json:"generation"`
	Status     string   `json:"status"`
	Fields     []string `json:"fields"`
}

func (h *Handler) Export(c echo.Context) error {
	var req exportRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "Format request tidak valid")
	}

	data, err := h.service.Export(c.Request().Context(), ExportInput{
		MemberIDs: req.MemberIDs,
		Filter: ListFilter{
			Search: req.Search, KordaID: req.KordaID, Generation: req.Generation, Status: req.Status,
		},
		Fields: req.Fields,
	})
	if err != nil {
		return handleError(c, err)
	}

	filename := "anggota-tibta65-" + time.Now().In(wib).Format("2006-01-02") + ".xlsx"
	c.Response().Header().Set(echo.HeaderContentDisposition, `attachment; filename="`+filename+`"`)
	return c.Blob(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}

func handleError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, ErrNotFound):
		return response.Error(c, http.StatusNotFound, "Anggota tidak ditemukan")
	case errors.Is(err, ErrInvalidStatusTransition), errors.Is(err, ErrInvalidRejectionType), errors.Is(err, ErrReasonRequired):
		return response.Error(c, http.StatusBadRequest, err.Error())
	case errors.Is(err, ErrNoExportField):
		return response.Error(c, http.StatusBadRequest, err.Error())
	case errors.Is(err, member.ErrValidation):
		return response.Error(c, http.StatusBadRequest, "Nama lengkap wajib diisi")
	case errors.Is(err, member.ErrDuplicateMemberNumber):
		return response.Error(c, http.StatusConflict, err.Error())
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
	group.GET("/export/fields", h.ExportFields)
	group.POST("/export", h.Export)
	group.POST("/legacy", h.CreateLegacyMember)
	group.POST("/:id/approve", h.Approve)
	group.POST("/:id/reject", h.Reject)
	group.PUT("/:id/status", h.UpdateStatus)
	group.DELETE("/:id", h.Delete)

	group.POST("/:id/reopen", h.Reopen, appMiddleware.RequireRole(auth.RoleSuperAdmin))
	group.POST("/:id/approve-rejected", h.ApproveRejected, appMiddleware.RequireRole(auth.RoleSuperAdmin))
}
