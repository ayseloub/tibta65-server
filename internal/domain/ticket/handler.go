package ticket

import (
	"errors"
	"net/http"
	"strconv"

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

type createRequest struct {
	Subject          string `json:"subject"`
	Message          string `json:"message"`
	ReportedMemberID string `json:"reported_member_id"`
}

func (h *Handler) Create(c echo.Context) error {
	memberID, _ := c.Get(appMiddleware.ContextKeyMemberID).(string)

	var req createRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "Format request tidak valid")
	}
	if req.Subject == "" || req.Message == "" {
		return response.Error(c, http.StatusBadRequest, "Subject dan pesan wajib diisi")
	}

	result, err := h.service.Create(c.Request().Context(), CreateInput{
		MemberID: memberID, Subject: req.Subject, Message: req.Message, ReportedMemberID: req.ReportedMemberID,
	})
	if err != nil {
		if errors.Is(err, ErrRateLimited) {
			return response.Error(c, http.StatusTooManyRequests, "Tunggu 5 menit sebelum mengirim tiket baru")
		}
		if errors.Is(err, ErrNotApproved) {
			return response.Error(c, http.StatusForbidden, err.Error())
		}
		return response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
	return response.Success(c, http.StatusCreated, "Tiket berhasil dikirim", result)
}

func (h *Handler) ListMine(c echo.Context) error {
	memberID, _ := c.Get(appMiddleware.ContextKeyMemberID).(string)
	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	items, total, err := h.service.ListMine(c.Request().Context(), memberID, page, limit)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
	return response.Success(c, http.StatusOK, "Berhasil mengambil data", map[string]interface{}{
		"items": items, "page": page, "limit": limit, "total": total,
		"total_pages": (total + limit - 1) / limit,
	})
}

func (h *Handler) GetMine(c echo.Context) error {
	memberID, _ := c.Get(appMiddleware.ContextKeyMemberID).(string)
	id := c.Param("id")

	result, err := h.service.GetMine(c.Request().Context(), id, memberID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return response.Error(c, http.StatusNotFound, "Tiket tidak ditemukan")
		}
		return response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
	return response.Success(c, http.StatusOK, "Berhasil mengambil data", result)
}

func (h *Handler) UnreadCount(c echo.Context) error {
	memberID, _ := c.Get(appMiddleware.ContextKeyMemberID).(string)
	count, err := h.service.CountUnreadByMember(c.Request().Context(), memberID)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
	return response.Success(c, http.StatusOK, "Berhasil mengambil data", map[string]int{"unread_count": count})
}

func (h *Handler) MarkRead(c echo.Context) error {
	memberID, _ := c.Get(appMiddleware.ContextKeyMemberID).(string)
	id := c.Param("id")
	if err := h.service.MarkReadByMember(c.Request().Context(), id, memberID); err != nil {
		return response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
	return response.Success(c, http.StatusOK, "Berhasil menandai dibaca", nil)
}

func (h *Handler) ListAdmin(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	items, total, err := h.service.ListAdmin(c.Request().Context(), c.QueryParam("korda_id"), c.QueryParam("status"), page, limit)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
	return response.Success(c, http.StatusOK, "Berhasil mengambil data", map[string]interface{}{
		"items": items, "page": page, "limit": limit, "total": total,
		"total_pages": (total + limit - 1) / limit,
	})
}

func (h *Handler) GetAdmin(c echo.Context) error {
	id := c.Param("id")
	result, err := h.service.GetAdmin(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return response.Error(c, http.StatusNotFound, "Tiket tidak ditemukan")
		}
		return response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
	return response.Success(c, http.StatusOK, "Berhasil mengambil data", result)
}

type replyRequest struct {
	AdminReply string `json:"admin_reply"`
}

func (h *Handler) Reply(c echo.Context) error {
	id := c.Param("id")
	var req replyRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "Format request tidak valid")
	}
	if req.AdminReply == "" {
		return response.Error(c, http.StatusBadRequest, "Balasan tidak boleh kosong")
	}

	result, err := h.service.Reply(c.Request().Context(), id, req.AdminReply)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return response.Error(c, http.StatusNotFound, "Tiket tidak ditemukan")
		}
		return response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
	return response.Success(c, http.StatusOK, "Balasan berhasil dikirim", result)
}
