package achievement

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

// List godoc
// @Summary      List achievement (publik)
// @Tags         achievement
// @Produce      json
// @Param        page query int false "Halaman"
// @Param        limit query int false "Jumlah per halaman"
// @Param        year query string false "Filter tahun"
// @Success      200 {object} response.Envelope
// @Router       /api/achievements [get]
func (h *Handler) List(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	year := c.QueryParam("year")

	result, err := h.service.List(c.Request().Context(), page, limit, year)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
	}

	return response.Success(c, http.StatusOK, "Berhasil mengambil data", result)
}

// Get godoc
// @Summary      Detail achievement (admin, buat form edit)
// @Tags         achievement
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "ID achievement"
// @Success      200 {object} response.Envelope
// @Failure      404 {object} response.Envelope
// @Router       /api/admin/achievements/{id} [get]
func (h *Handler) Get(c echo.Context) error {
	id := c.Param("id")

	a, err := h.service.Get(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, ErrAchievementNotFound) {
			return response.Error(c, http.StatusNotFound, "Achievement tidak ditemukan")
		}
		return response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
	}

	return response.Success(c, http.StatusOK, "Berhasil mengambil data", a)
}

// Create godoc
// @Summary      Tambah achievement baru
// @Tags         achievement
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        year formData string true "Tahun"
// @Param        title formData string true "Judul"
// @Param        description formData string true "Deskripsi"
// @Param        image formData file true "Gambar"
// @Success      201 {object} response.Envelope
// @Failure      400 {object} response.Envelope
// @Router       /api/admin/achievements [post]
func (h *Handler) Create(c echo.Context) error {
	year := c.FormValue("year")
	title := c.FormValue("title")
	description := c.FormValue("description")

	file, err := c.FormFile("image")
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "Gambar wajib diupload")
	}

	a, err := h.service.Create(c.Request().Context(), year, title, description, file)
	if err != nil {
		if errors.Is(err, ErrValidation) {
			return response.Error(c, http.StatusBadRequest, "Semua field wajib diisi")
		}
		return response.Error(c, http.StatusBadRequest, err.Error())
	}

	return response.Success(c, http.StatusCreated, "Achievement berhasil ditambahkan", a)
}

// Update godoc
// @Summary      Update achievement
// @Tags         achievement
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "ID achievement"
// @Param        year formData string true "Tahun"
// @Param        title formData string true "Judul"
// @Param        description formData string true "Deskripsi"
// @Param        image formData file false "Gambar baru (opsional)"
// @Success      200 {object} response.Envelope
// @Failure      400 {object} response.Envelope
// @Router       /api/admin/achievements/{id} [put]
func (h *Handler) Update(c echo.Context) error {
	id := c.Param("id")
	year := c.FormValue("year")
	title := c.FormValue("title")
	description := c.FormValue("description")

	// Gambar opsional pas update — kalau admin gak pilih file baru, ini bakal error "http: no such file",
	// jadi kita anggap nil (bukan gagal total), bukan di-treat sebagai error fatal.
	file, _ := c.FormFile("image")

	a, err := h.service.Update(c.Request().Context(), id, year, title, description, file)
	if err != nil {
		if errors.Is(err, ErrValidation) {
			return response.Error(c, http.StatusBadRequest, "Semua field wajib diisi")
		}
		if errors.Is(err, ErrAchievementNotFound) {
			return response.Error(c, http.StatusNotFound, "Achievement tidak ditemukan")
		}
		return response.Error(c, http.StatusBadRequest, err.Error())
	}

	return response.Success(c, http.StatusOK, "Achievement berhasil diperbarui", a)
}

// Delete godoc
// @Summary      Hapus achievement
// @Tags         achievement
// @Security     BearerAuth
// @Param        id path string true "ID achievement"
// @Success      200 {object} response.Envelope
// @Failure      404 {object} response.Envelope
// @Router       /api/admin/achievements/{id} [delete]
func (h *Handler) Delete(c echo.Context) error {
	id := c.Param("id")

	if err := h.service.Delete(c.Request().Context(), id); err != nil {
		if errors.Is(err, ErrAchievementNotFound) {
			return response.Error(c, http.StatusNotFound, "Achievement tidak ditemukan")
		}
		return response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
	}

	return response.Success(c, http.StatusOK, "Achievement berhasil dihapus", nil)
}

func RegisterRoutes(e *echo.Echo, h *Handler, jwtSecret string) {
	e.GET("/api/achievements", h.List)

	adminGroup := e.Group("/api/admin/achievements",
		appMiddleware.RequireAuth(jwtSecret),
		appMiddleware.RequireRole(auth.RoleAdmin, auth.RoleSuperAdmin),
	)
	adminGroup.GET("", h.List)
	adminGroup.GET("/:id", h.Get)
	adminGroup.POST("", h.Create)
	adminGroup.PUT("/:id", h.Update)
	adminGroup.DELETE("/:id", h.Delete)
}
