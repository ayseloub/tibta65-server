package member

import (
	"errors"
	"net/http"

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

type registerRequest struct {
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	KordaID  string `json:"korda_id"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type googleLoginRequest struct {
	IDToken string `json:"id_token"`
}

type forgotPasswordRequest struct {
	Email string `json:"email"`
}

type resetPasswordRequest struct {
	Email       string `json:"email"`
	Code        string `json:"code"`
	NewPassword string `json:"new_password"`
}

type updateProfileRequest struct {
	FullName string `json:"full_name"`
	Phone    string `json:"phone"`
	KordaID  string `json:"korda_id"`
	Address  string `json:"address"`
}

type changePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

func (h *Handler) UpdateProfile(c echo.Context) error {
	memberID, _ := c.Get(appMiddleware.ContextKeyMemberID).(string)

	var req updateProfileRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "Format request tidak valid")
	}

	m, err := h.service.UpdateProfile(c.Request().Context(), memberID, UpdateProfileInput{
		FullName: req.FullName, Phone: req.Phone, KordaID: req.KordaID, Address: req.Address,
	})
	if err != nil {
		return handleError(c, err)
	}
	return response.Success(c, http.StatusOK, "Profil berhasil diperbarui", m)
}

func (h *Handler) UpdateAvatar(c echo.Context) error {
	memberID, _ := c.Get(appMiddleware.ContextKeyMemberID).(string)

	file, err := c.FormFile("avatar")
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "Gambar wajib diupload")
	}

	m, err := h.service.UpdateAvatar(c.Request().Context(), memberID, file)
	if err != nil {
		return handleError(c, err)
	}
	return response.Success(c, http.StatusOK, "Foto profil berhasil diperbarui", m)
}

func (h *Handler) DeleteAvatar(c echo.Context) error {
	memberID, _ := c.Get(appMiddleware.ContextKeyMemberID).(string)

	if err := h.service.DeleteAvatar(c.Request().Context(), memberID); err != nil {
		return handleError(c, err)
	}
	return response.Success(c, http.StatusOK, "Foto profil berhasil dihapus", nil)
}

func (h *Handler) ChangePassword(c echo.Context) error {
	memberID, _ := c.Get(appMiddleware.ContextKeyMemberID).(string)

	var req changePasswordRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "Format request tidak valid")
	}

	if err := h.service.ChangePassword(c.Request().Context(), memberID, req.CurrentPassword, req.NewPassword); err != nil {
		return handleError(c, err)
	}
	return response.Success(c, http.StatusOK, "Password berhasil diubah", nil)
}

func (h *Handler) ForgotPassword(c echo.Context) error {
	var req forgotPasswordRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "Format request tidak valid")
	}

	_ = h.service.ForgotPassword(c.Request().Context(), req.Email)

	return response.Success(c, http.StatusOK, "Kalau email terdaftar, kode reset password sudah dikirim", nil)
}

func (h *Handler) ResetPassword(c echo.Context) error {
	var req resetPasswordRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "Format request tidak valid")
	}

	if err := h.service.ResetPassword(c.Request().Context(), req.Email, req.Code, req.NewPassword); err != nil {
		return handleError(c, err)
	}
	return response.Success(c, http.StatusOK, "Password berhasil direset, silakan login", nil)
}

func (h *Handler) LoginWithGoogle(c echo.Context) error {
	var req googleLoginRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "Format request tidak valid")
	}

	result, err := h.service.LoginWithGoogle(c.Request().Context(), req.IDToken)
	if err != nil {
		return response.Error(c, http.StatusUnauthorized, err.Error())
	}
	return response.Success(c, http.StatusOK, "Login berhasil", result)
}

func (h *Handler) Register(c echo.Context) error {
	var req registerRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "Format request tidak valid")
	}

	m, err := h.service.Register(c.Request().Context(), RegisterInput{
		FullName: req.FullName, Email: req.Email, KordaID: req.KordaID, Password: req.Password,
	})
	if err != nil {
		return handleError(c, err)
	}
	return response.Success(c, http.StatusCreated, "Registrasi berhasil", m)
}

func (h *Handler) Login(c echo.Context) error {
	var req loginRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "Format request tidak valid")
	}

	result, err := h.service.Login(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		return handleError(c, err)
	}
	return response.Success(c, http.StatusOK, "Login berhasil", result)
}

func (h *Handler) Me(c echo.Context) error {
	memberID, _ := c.Get(appMiddleware.ContextKeyMemberID).(string)
	m, err := h.service.Me(c.Request().Context(), memberID)
	if err != nil {
		return response.Error(c, http.StatusUnauthorized, "Sesi tidak valid")
	}
	return response.Success(c, http.StatusOK, "Berhasil mengambil data", m)
}

func handleError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, ErrValidation):
		return response.Error(c, http.StatusBadRequest, "Semua field wajib diisi")
	case errors.Is(err, ErrDuplicate):
		return response.Error(c, http.StatusConflict, "Email sudah terdaftar")
	case errors.Is(err, ErrInvalidCredentials):
		return response.Error(c, http.StatusUnauthorized, "Email atau password salah")
	case errors.Is(err, ErrGoogleOnlyAccount):
		return response.Error(c, http.StatusConflict, err.Error())
	case errors.Is(err, ErrEmailNotVerified):
		return response.Error(c, http.StatusForbidden, err.Error())
	case errors.Is(err, ErrInvalidOTP):
		return response.Error(c, http.StatusBadRequest, err.Error())
	default:
		return response.Error(c, http.StatusBadRequest, err.Error())
	}
}

func RegisterRoutes(e *echo.Echo, h *Handler, jwtSecret string) {
	e.POST("/api/member-auth/register", h.Register)
	e.POST("/api/member-auth/login", h.Login)
	e.GET("/api/member-auth/me", h.Me, appMiddleware.RequireMemberAuth(jwtSecret))
	e.POST("/api/member-auth/verify-otp", h.VerifyOTP)
	e.POST("/api/member-auth/google", h.LoginWithGoogle)
	e.POST("/api/member-auth/forgot-password", h.ForgotPassword)
	e.POST("/api/member-auth/reset-password", h.ResetPassword)
	e.PUT("/api/member-auth/profile", h.UpdateProfile, appMiddleware.RequireMemberAuth(jwtSecret))
	e.POST("/api/member-auth/avatar", h.UpdateAvatar, appMiddleware.RequireMemberAuth(jwtSecret))
	e.DELETE("/api/member-auth/avatar", h.DeleteAvatar, appMiddleware.RequireMemberAuth(jwtSecret))
	e.POST("/api/member-auth/change-password", h.ChangePassword, appMiddleware.RequireMemberAuth(jwtSecret))
}

type verifyOTPRequest struct {
	MemberID string `json:"member_id"`
	Code     string `json:"code"`
}

func (h *Handler) VerifyOTP(c echo.Context) error {
	var req verifyOTPRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "Format request tidak valid")
	}

	if err := h.service.VerifyOTP(c.Request().Context(), req.MemberID, req.Code); err != nil {
		return handleError(c, err)
	}
	return response.Success(c, http.StatusOK, "Email berhasil diverifikasi", nil)
}
