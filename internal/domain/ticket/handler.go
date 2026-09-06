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

// ===== Member-facing =====

type createRequest struct {
	Subject string `json:"subject"`
	Message string `json:"message"`
}

func (h *Handler) Create(c echo.Context) error {
	memberID := c.Get(appMiddleware.ContextKeyMemberID).(string)

	var req createRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "Format request tidak valid")
	}
	if req.Subject == "" || req.Message == "" {
		return response.Error(c, http.StatusBadRequest, "Subject dan pesan wajib diisi")
	}

	result, err := h.service.Create(c.Request().Context(), CreateInput{
		MemberID: memberID, Subject: req.Subject, Message: req.Message,
	})
	if err != nil {
		if errors.Is(err, ErrRateLimited) {
			return response.Error(c, http.StatusTooManyRequests, "Tunggu 5 menit sebelum mengirim tiket baru")
		}
		return response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
	return response.Success(c, http.StatusCreated, "Tiket berhasil dikirim", result)
}

func (h *Handler) ListMine(c echo.Context) error {
	memberID := c.Get(appMiddleware.ContextKeyMemberID).(string)
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
	memberID := c.Get(appMiddleware.ContextKeyMemberID).(string)
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

// ===== Admin-facing =====

func (h *Handler) ListAdmin(c echo.Context) error {
	kordaID := c.QueryParam("korda_id")
	status := c.QueryParam("status")
	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	items, total, err := h.service.ListAdmin(c.Request().Context(), kordaID, status, page, limit)
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
		return response.Error(c, http.StatusBadRequest, "Balasan wajib diisi")
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
