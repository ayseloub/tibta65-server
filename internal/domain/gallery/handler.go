package gallery

import (
	"errors"
	"net/http"
	"strconv"

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

type albumRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	EventDate   string `json:"event_date"`
	KordaID     string `json:"korda_id"`
	KategoriID  string `json:"kategori_id"`
	Visibility  string `json:"visibility"`
}

type highlightRequest struct {
	Highlight bool `json:"highlight"`
}

type captionRequest struct {
	Caption string `json:"caption"`
}

func (h *Handler) List(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	search := c.QueryParam("search")
	visibility := c.QueryParam("visibility") // admin boleh filter/lihat semua

	result, err := h.service.List(c.Request().Context(), search, visibility, page, limit)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
	return response.Success(c, http.StatusOK, "Berhasil mengambil data", result)
}

func (h *Handler) Get(c echo.Context) error {
	id := c.Param("id")
	result, err := h.service.Get(c.Request().Context(), id)
	if err != nil {
		return handleError(c, err)
	}
	return response.Success(c, http.StatusOK, "Berhasil mengambil data", result)
}

func (h *Handler) GetHighlight(c echo.Context) error {
	result, err := h.service.GetHighlight(c.Request().Context())
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return response.Success(c, http.StatusOK, "Belum ada album highlight", nil)
		}
		return response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
	return response.Success(c, http.StatusOK, "Berhasil mengambil data", result)
}

func (h *Handler) Create(c echo.Context) error {
	var req albumRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "Format request tidak valid")
	}
	a, err := h.service.Create(c.Request().Context(), AlbumInput{
		Title: req.Title, Description: req.Description, EventDate: req.EventDate,
		KordaID: req.KordaID, KategoriID: req.KategoriID, Visibility: req.Visibility,
	})
	if err != nil {
		return handleError(c, err)
	}
	return response.Success(c, http.StatusCreated, "Album berhasil ditambahkan", a)
}

func (h *Handler) Update(c echo.Context) error {
	id := c.Param("id")
	var req albumRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "Format request tidak valid")
	}
	a, err := h.service.Update(c.Request().Context(), id, AlbumInput{
		Title: req.Title, Description: req.Description, EventDate: req.EventDate,
		KordaID: req.KordaID, KategoriID: req.KategoriID, Visibility: req.Visibility,
	})
	if err != nil {
		return handleError(c, err)
	}
	return response.Success(c, http.StatusOK, "Album berhasil diperbarui", a)
}

func (h *Handler) Delete(c echo.Context) error {
	id := c.Param("id")
	if err := h.service.Delete(c.Request().Context(), id); err != nil {
		return handleError(c, err)
	}
	return response.Success(c, http.StatusOK, "Album berhasil dihapus", nil)
}

func (h *Handler) SetHighlight(c echo.Context) error {
	id := c.Param("id")
	var req highlightRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "Format request tidak valid")
	}
	if err := h.service.SetHighlight(c.Request().Context(), id, req.Highlight); err != nil {
		return handleError(c, err)
	}
	msg := "Album dijadikan highlight"
	if !req.Highlight {
		msg = "Highlight album dimatikan"
	}
	return response.Success(c, http.StatusOK, msg, nil)
}

func (h *Handler) AddPhoto(c echo.Context) error {
	albumID := c.Param("id")
	caption := c.FormValue("caption")

	file, err := c.FormFile("image")
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "Gambar wajib diupload")
	}

	photo, err := h.service.AddPhoto(c.Request().Context(), albumID, file, caption)
	if err != nil {
		return handleError(c, err)
	}
	return response.Success(c, http.StatusCreated, "Foto berhasil ditambahkan", photo)
}

func (h *Handler) UpdatePhotoCaption(c echo.Context) error {
	photoID := c.Param("photoId")
	var req captionRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "Format request tidak valid")
	}
	if err := h.service.UpdatePhotoCaption(c.Request().Context(), photoID, req.Caption); err != nil {
		return handleError(c, err)
	}
	return response.Success(c, http.StatusOK, "Caption berhasil diperbarui", nil)
}

func (h *Handler) DeletePhoto(c echo.Context) error {
	photoID := c.Param("photoId")
	if err := h.service.DeletePhoto(c.Request().Context(), photoID); err != nil {
		return handleError(c, err)
	}
	return response.Success(c, http.StatusOK, "Foto berhasil dihapus", nil)
}

func (h *Handler) ListPublic(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	search := c.QueryParam("search")

	result, err := h.service.List(c.Request().Context(), search, VisibilityPublic, page, limit)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
	return response.Success(c, http.StatusOK, "Berhasil mengambil data", result)
}

func (h *Handler) GetPublic(c echo.Context) error {
	id := c.Param("id")
	result, err := h.service.Get(c.Request().Context(), id)
	if err != nil {
		return handleError(c, err)
	}
	if result.Visibility != VisibilityPublic {
		return response.Error(c, http.StatusNotFound, "Album tidak ditemukan")
	}
	return response.Success(c, http.StatusOK, "Berhasil mengambil data", result)
}

func (h *Handler) ListInternal(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))

	result, err := h.service.List(c.Request().Context(), "", VisibilityInternal, page, limit)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
	return response.Success(c, http.StatusOK, "Berhasil mengambil data", result)
}

func (h *Handler) GetInternal(c echo.Context) error {
	id := c.Param("id")
	result, err := h.service.Get(c.Request().Context(), id)
	if err != nil {
		return handleError(c, err)
	}
	return response.Success(c, http.StatusOK, "Berhasil mengambil data", result)
}

func handleError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, ErrValidation):
		return response.Error(c, http.StatusBadRequest, "Semua field wajib diisi")
	case errors.Is(err, ErrNotFound):
		return response.Error(c, http.StatusNotFound, "Album tidak ditemukan")
	case errors.Is(err, ErrPhotoNotFound):
		return response.Error(c, http.StatusNotFound, "Foto tidak ditemukan")
	case errors.Is(err, ErrMaxPhotos):
		return response.Error(c, http.StatusConflict, err.Error())
	default:
		return response.Error(c, http.StatusBadRequest, err.Error())
	}
}

func RegisterRoutes(e *echo.Echo, h *Handler, jwtSecret, memberJWTSecret string) {
	e.GET("/api/gallery", h.ListPublic)
	e.GET("/api/gallery/highlight", h.GetHighlight)
	e.GET("/api/gallery/:id", h.GetPublic)

	group := e.Group("/api/admin/gallery",
		appMiddleware.RequireAuth(jwtSecret),
		appMiddleware.RequireRole(auth.RoleAdmin, auth.RoleSuperAdmin),
	)
	group.GET("", h.List)
	group.GET("/:id", h.Get)
	group.POST("", h.Create)
	group.PUT("/:id", h.Update)
	group.DELETE("/:id", h.Delete)
	group.PUT("/:id/highlight", h.SetHighlight)
	group.POST("/:id/photos", h.AddPhoto)
	group.PUT("/photos/:photoId", h.UpdatePhotoCaption)
	group.DELETE("/photos/:photoId", h.DeletePhoto)

	memberGroup := e.Group("/api/member/gallery", appMiddleware.RequireMemberAuth(memberJWTSecret))
	memberGroup.GET("", h.ListInternal)
	memberGroup.GET("/:id", h.GetInternal)
}
