package memberactivation

import (
	"errors"
	"net/http"

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

type generateTokenRequest struct {
	Purpose string `json:"purpose"`
}

func (h *Handler) AdminGenerateToken(c echo.Context) error {
	memberID := c.Param("id")
	adminID, _ := c.Get(appMiddleware.ContextKeyAdminID).(string)

	var req generateTokenRequest
	_ = c.Bind(&req)
	if req.Purpose == "" {
		req.Purpose = PurposeClaim
	}

	rawToken, err := h.service.AdminGenerateToken(c.Request().Context(), memberID, adminID, req.Purpose)
	if err != nil {
		return handleError(c, err)
	}
	return response.Success(c, http.StatusOK, "Token berhasil dibuat", map[string]string{"token": rawToken})
}

type checkTokenRequest struct {
	MemberNumber string `json:"member_number"`
	Token        string `json:"token"`
}

func (h *Handler) CheckClaimToken(c echo.Context) error {
	var req checkTokenRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "Format request tidak valid")
	}

	m, err := h.service.CheckClaimToken(c.Request().Context(), req.MemberNumber, req.Token)
	if err != nil {
		return handleError(c, err)
	}
	return response.Success(c, http.StatusOK, "Token valid", m)
}

type claimRequest struct {
	MemberNumber    string `json:"member_number"`
	Token           string `json:"token"`
	Username        string `json:"username"`
	Password        string `json:"password"`
	KordaID         string `json:"korda_id"`
	Phone           string `json:"phone"`
	Address         string `json:"address"`
	FullName        string `json:"full_name"`
	NamaSuci        string `json:"nama_suci"`
	Agama           string `json:"agama"`
	NRP             string `json:"nrp"`
	NoAK            string `json:"no_ak"`
	PangkatTerakhir string `json:"pangkat_terakhir"`
}

type claimGoogleRequest struct {
	MemberNumber    string `json:"member_number"`
	Token           string `json:"token"`
	IDToken         string `json:"id_token"`
	KordaID         string `json:"korda_id"`
	Phone           string `json:"phone"`
	Address         string `json:"address"`
	FullName        string `json:"full_name"`
	NamaSuci        string `json:"nama_suci"`
	Agama           string `json:"agama"`
	NRP             string `json:"nrp"`
	NoAK            string `json:"no_ak"`
	PangkatTerakhir string `json:"pangkat_terakhir"`
}

func (h *Handler) ClaimWithGoogle(c echo.Context) error {
	var req claimGoogleRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "Format request tidak valid")
	}

	result, err := h.service.ClaimWithGoogle(c.Request().Context(), ClaimWithGoogleInput{
		MemberNumber: req.MemberNumber, Token: req.Token, IDToken: req.IDToken, KordaID: req.KordaID,
		Phone: req.Phone, Address: req.Address, FullName: req.FullName, NamaSuci: req.NamaSuci,
		Agama: req.Agama, NRP: req.NRP, NoAK: req.NoAK, PangkatTerakhir: req.PangkatTerakhir,
	})
	if err != nil {
		return handleError(c, err)
	}
	return response.Success(c, http.StatusOK, "Akun berhasil diaktivasi", result)
}

type resetPasswordTokenRequest struct {
	MemberNumber string `json:"member_number"`
	Token        string `json:"token"`
	NewPassword  string `json:"new_password"`
}

func (h *Handler) ResetPasswordWithToken(c echo.Context) error {
	var req resetPasswordTokenRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "Format request tidak valid")
	}

	if err := h.service.ResetPasswordWithToken(c.Request().Context(), req.MemberNumber, req.Token, req.NewPassword); err != nil {
		return handleError(c, err)
	}
	return response.Success(c, http.StatusOK, "Password berhasil direset, silakan login", nil)
}

func (h *Handler) ClaimWithPassword(c echo.Context) error {
	var req claimRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "Format request tidak valid")
	}

	result, err := h.service.ClaimWithPassword(c.Request().Context(), ClaimInput{
		MemberNumber: req.MemberNumber, Token: req.Token, Username: req.Username, Password: req.Password,
		KordaID: req.KordaID, Phone: req.Phone, Address: req.Address, FullName: req.FullName,
		NamaSuci: req.NamaSuci, Agama: req.Agama, NRP: req.NRP, NoAK: req.NoAK, PangkatTerakhir: req.PangkatTerakhir,
	})
	if err != nil {
		return handleError(c, err)
	}
	return response.Success(c, http.StatusOK, "Akun berhasil diaktivasi", result)
}

func handleError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, ErrInvalidToken):
		return response.Error(c, http.StatusBadRequest, err.Error())
	case errors.Is(err, ErrAlreadyClaimed):
		return response.Error(c, http.StatusConflict, err.Error())
	case errors.Is(err, ErrMemberDeceased):
		return response.Error(c, http.StatusForbidden, err.Error())
	case errors.Is(err, ErrValidation):
		return response.Error(c, http.StatusBadRequest, "Semua field wajib diisi")
	case errors.Is(err, member.ErrNotFound):
		return response.Error(c, http.StatusNotFound, "Anggota tidak ditemukan")
	case errors.Is(err, member.ErrDuplicateUsername):
		return response.Error(c, http.StatusConflict, err.Error())
	case errors.Is(err, ErrGoogleAlreadyUsed):
		return response.Error(c, http.StatusConflict, err.Error())
	case errors.Is(err, ErrEmailAlreadyUsed):
		return response.Error(c, http.StatusConflict, err.Error())
	case errors.Is(err, ErrNotClaimedYet):
		return response.Error(c, http.StatusBadRequest, err.Error())
	case errors.Is(err, ErrTooManyAttempts):
		return response.Error(c, http.StatusTooManyRequests, err.Error())
	default:
		return response.Error(c, http.StatusBadRequest, err.Error())
	}
}

func RegisterRoutes(e *echo.Echo, h *Handler, jwtSecret string) {
	adminGroup := e.Group("/api/admin/members",
		appMiddleware.RequireAuth(jwtSecret),
		appMiddleware.RequireRole(auth.RoleAdmin, auth.RoleSuperAdmin),
	)
	adminGroup.POST("/:id/generate-token", h.AdminGenerateToken)

	e.POST("/api/member-activation/check", h.CheckClaimToken)
	e.POST("/api/member-activation/claim", h.ClaimWithPassword)
	e.POST("/api/member-activation/claim-google", h.ClaimWithGoogle)
	e.POST("/api/member-activation/reset-password", h.ResetPasswordWithToken)
}
