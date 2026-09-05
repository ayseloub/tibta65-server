package member

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"mime/multipart"
	"time"

	"github.com/oklog/ulid/v2"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/api/idtoken"

	"github.com/Tibta65web/tibta65-server/pkg/email"
	"github.com/Tibta65web/tibta65-server/pkg/jwt"

	"github.com/Tibta65web/tibta65-server/pkg/storage"
	"github.com/rs/zerolog/log"
)

var (
	ErrValidation         = errors.New("data tidak valid")
	ErrInvalidCredentials = errors.New("email atau password salah")
	ErrGoogleOnlyAccount  = errors.New("akun ini terdaftar via Google, silakan login dengan Google")
	ErrEmailNotVerified   = errors.New("email belum diverifikasi, silakan cek kode OTP yang dikirim ke email kamu")
	ErrInvalidOTP         = errors.New("kode OTP salah atau sudah kedaluwarsa")

	otpExpiry = 10 * time.Minute
)

type RegisterInput struct {
	FullName string
	Email    string
	KordaID  string
	Password string
}

type LoginResult struct {
	Token string `json:"token"`
}

type Service interface {
	Register(ctx context.Context, in RegisterInput) (*Member, error)
	VerifyOTP(ctx context.Context, memberID, code string) error
	Login(ctx context.Context, email, password string) (*LoginResult, error)
	Me(ctx context.Context, id string) (*Member, error)
	LoginWithGoogle(ctx context.Context, idTokenString string) (*LoginResult, error)
	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, email, code, newPassword string) error
	UpdateProfile(ctx context.Context, memberID string, in UpdateProfileInput) (*Member, error)
	UpdateAvatar(ctx context.Context, memberID string, file *multipart.FileHeader) (*Member, error)
	DeleteAvatar(ctx context.Context, memberID string) error
	ChangePassword(ctx context.Context, memberID, currentPassword, newPassword string) error
}

type service struct {
	repo           Repository
	otpRepo        OTPRepository
	emailSender    email.EmailSender
	storage        storage.Storage
	jwtSecret      string
	jwtExpiry      time.Duration
	googleClientID string
}

type UpdateProfileInput struct {
	FullName string
	Phone    string
	KordaID  string
	Address  string
}

func (s *service) UpdateProfile(ctx context.Context, memberID string, in UpdateProfileInput) (*Member, error) {
	if in.FullName == "" {
		return nil, ErrValidation
	}

	var phonePtr, addressPtr, kordaPtr *string
	if in.Phone != "" {
		phonePtr = &in.Phone
	}
	if in.Address != "" {
		addressPtr = &in.Address
	}
	if in.KordaID != "" {
		kordaPtr = &in.KordaID
	}

	m := &Member{ID: memberID, FullName: in.FullName, Phone: phonePtr, Address: addressPtr, KordaID: kordaPtr}
	if err := s.repo.UpdateProfile(ctx, m); err != nil {
		return nil, err
	}
	return s.repo.FindByID(ctx, memberID)
}

func (s *service) UpdateAvatar(ctx context.Context, memberID string, file *multipart.FileHeader) (*Member, error) {
	url, err := s.storage.Upload(ctx, file, "member-avatar")
	if err != nil {
		return nil, err
	}

	existing, _ := s.repo.FindByID(ctx, memberID)

	if err := s.repo.UpdateAvatar(ctx, memberID, &url); err != nil {
		_ = s.storage.Delete(ctx, url)
		return nil, err
	}

	if existing != nil && existing.AvatarURL != nil {
		_ = s.storage.Delete(ctx, *existing.AvatarURL)
	}

	return s.repo.FindByID(ctx, memberID)
}

func (s *service) DeleteAvatar(ctx context.Context, memberID string) error {
	existing, err := s.repo.FindByID(ctx, memberID)
	if err != nil {
		return err
	}

	if err := s.repo.UpdateAvatar(ctx, memberID, nil); err != nil {
		return err
	}
	if existing.AvatarURL != nil {
		_ = s.storage.Delete(ctx, *existing.AvatarURL)
	}
	return nil
}

func (s *service) ChangePassword(ctx context.Context, memberID, currentPassword, newPassword string) error {
	m, err := s.repo.FindByID(ctx, memberID)
	if err != nil {
		return err
	}

	if m.PasswordHash == nil {
		return ErrGoogleOnlyAccount
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*m.PasswordHash), []byte(currentPassword)); err != nil {
		return ErrInvalidCredentials
	}
	if len(newPassword) < 8 {
		return errors.New("password minimal 8 karakter")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.repo.UpdatePassword(ctx, memberID, string(hash))
}

func NewService(repo Repository, otpRepo OTPRepository, emailSender email.EmailSender, storage storage.Storage, jwtSecret string, jwtExpiry time.Duration, googleClientID string) Service {
	return &service{
		repo: repo, otpRepo: otpRepo, emailSender: emailSender, storage: storage,
		jwtSecret: jwtSecret, jwtExpiry: jwtExpiry, googleClientID: googleClientID,
	}
}

func generateOTPCode() string {
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

func (s *service) Register(ctx context.Context, in RegisterInput) (*Member, error) {
	if in.FullName == "" || in.Email == "" || in.KordaID == "" || in.Password == "" {
		return nil, ErrValidation
	}
	if len(in.Password) < 8 {
		return nil, errors.New("password minimal 8 karakter")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	hashStr := string(hash)

	kordaID := in.KordaID
	m := &Member{
		ID:           ulid.Make().String(),
		FullName:     in.FullName,
		Email:        in.Email,
		PasswordHash: &hashStr,
		KordaID:      &kordaID,
	}

	if err := s.repo.Create(ctx, m); err != nil {
		return nil, err
	}

	code := generateOTPCode()
	if err := s.otpRepo.Create(ctx, m.ID, code, OTPPurposeRegister, time.Now().Add(otpExpiry)); err != nil {
		log.Error().Err(err).Msg("gagal simpan OTP ke database")
	} else {
		body, err := email.RenderOTPEmail(
			"Verifikasi Akun TIBTA 65",
			"Gunakan kode berikut untuk memverifikasi email kamu dan menyelesaikan pendaftaran.",
			code,
			time.Now().Year(),
		)
		if err != nil {
			log.Error().Err(err).Msg("gagal render template email")
		} else {
			if err := s.emailSender.Send(ctx, m.Email, "Verifikasi Akun TIBTA 65", body); err != nil {
				log.Error().Err(err).Msg("gagal kirim email OTP")
			}
		}
	}
	return m, nil
}

func (s *service) VerifyOTP(ctx context.Context, memberID, code string) error {
	otp, err := s.otpRepo.FindValidByMemberID(ctx, memberID, OTPPurposeRegister)
	if err != nil {
		return ErrInvalidOTP
	}

	if otp.OTPCode != code {
		return ErrInvalidOTP
	}

	if err := s.repo.MarkEmailVerified(ctx, memberID); err != nil {
		return err
	}

	_ = s.otpRepo.DeleteByMemberID(ctx, memberID, OTPPurposeRegister)
	return nil
}

func (s *service) Login(ctx context.Context, emailAddr, password string) (*LoginResult, error) {
	m, err := s.repo.FindByEmail(ctx, emailAddr)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if m.PasswordHash == nil {
		return nil, ErrGoogleOnlyAccount
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*m.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	if m.EmailVerifiedAt == nil {
		return nil, ErrEmailNotVerified
	}

	token, err := jwt.GenerateMemberToken(s.jwtSecret, m.ID, s.jwtExpiry)
	if err != nil {
		return nil, err
	}

	return &LoginResult{Token: token}, nil
}

func (s *service) Me(ctx context.Context, id string) (*Member, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *service) LoginWithGoogle(ctx context.Context, idTokenString string) (*LoginResult, error) {
	payload, err := idtoken.Validate(ctx, idTokenString, s.googleClientID)
	if err != nil {
		return nil, errors.New("token Google tidak valid")
	}

	googleID := payload.Subject
	emailAddr, _ := payload.Claims["email"].(string)
	fullName, _ := payload.Claims["name"].(string)
	picture, _ := payload.Claims["picture"].(string)

	if emailAddr == "" {
		return nil, errors.New("email tidak ditemukan pada akun Google")
	}

	m, err := s.repo.FindByGoogleID(ctx, googleID)
	if err == nil {
		return s.issueTokenFor(m)
	}

	m, err = s.repo.FindByEmail(ctx, emailAddr)
	if err == nil {
		var avatarPtr *string
		if picture != "" {
			avatarPtr = &picture
		}
		if err := s.repo.LinkGoogleID(ctx, m.ID, googleID, avatarPtr); err != nil {
			return nil, err
		}
		return s.issueTokenFor(m)
	}

	newMember := &Member{
		ID:              ulid.Make().String(),
		FullName:        fullName,
		Email:           emailAddr,
		GoogleID:        &googleID,
		KordaID:         nil,
		EmailVerifiedAt: timePtr(time.Now()),
	}
	if picture != "" {
		newMember.AvatarURL = &picture
	}

	if err := s.repo.Create(ctx, newMember); err != nil {
		return nil, err
	}

	return s.issueTokenFor(newMember)
}

func (s *service) issueTokenFor(m *Member) (*LoginResult, error) {
	token, err := jwt.GenerateMemberToken(s.jwtSecret, m.ID, s.jwtExpiry)
	if err != nil {
		return nil, err
	}
	return &LoginResult{Token: token}, nil
}

func timePtr(t time.Time) *time.Time {
	return &t
}

func (s *service) ForgotPassword(ctx context.Context, emailAddr string) error {
	m, err := s.repo.FindByEmail(ctx, emailAddr)
	if err != nil {
		return nil
	}

	if m.PasswordHash == nil {
		return nil
	}

	code := generateOTPCode()
	if err := s.otpRepo.Create(ctx, m.ID, code, OTPPurposeResetPassword, time.Now().Add(otpExpiry)); err != nil {
		return nil
	}

	body, _ := email.RenderOTPEmail(
		"Reset Password TIBTA 65",
		"Gunakan kode berikut untuk membuat kata sandi baru. Abaikan email ini jika kamu tidak meminta reset password.",
		code,
		time.Now().Year(),
	)
	_ = s.emailSender.Send(ctx, m.Email, "Reset Password TIBTA 65", body)

	return nil
}

func (s *service) ResetPassword(ctx context.Context, emailAddr, code, newPassword string) error {
	if len(newPassword) < 8 {
		return errors.New("password minimal 8 karakter")
	}

	m, err := s.repo.FindByEmail(ctx, emailAddr)
	if err != nil {
		return ErrInvalidOTP
	}

	otp, err := s.otpRepo.FindValidByMemberID(ctx, m.ID, OTPPurposeResetPassword)
	if err != nil || otp.OTPCode != code {
		return ErrInvalidOTP
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	if err := s.repo.UpdatePassword(ctx, m.ID, string(hash)); err != nil {
		return err
	}

	_ = s.otpRepo.DeleteByMemberID(ctx, m.ID, OTPPurposeResetPassword)
	return nil
}
