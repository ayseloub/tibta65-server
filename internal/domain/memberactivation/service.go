package memberactivation

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/oklog/ulid/v2"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/api/idtoken"

	"github.com/Tibta65web/tibta65-server/internal/domain/member"
	"github.com/Tibta65web/tibta65-server/pkg/jwt"
	"github.com/Tibta65web/tibta65-server/pkg/storage"
	"github.com/rs/zerolog/log"
)

var (
	ErrInvalidToken      = errors.New("nomor induk atau token tidak valid, atau sudah kedaluwarsa")
	ErrAlreadyClaimed    = errors.New("akun ini sudah pernah diaktivasi")
	ErrMemberDeceased    = errors.New("akun ini berstatus meninggal dan tidak dapat diaktivasi")
	ErrValidation        = errors.New("data tidak valid")
	ErrGoogleAlreadyUsed = errors.New("akun Google ini sudah terhubung ke anggota lain")
	ErrEmailAlreadyUsed  = errors.New("email ini sudah digunakan akun lain")
	ErrNotClaimedYet     = errors.New("akun ini belum pernah diaktivasi, silakan aktivasi dulu")

	tokenExpiry = 48 * time.Hour
)

type ClaimInput struct {
	MemberNumber    string
	Token           string
	Username        string
	Password        string
	KordaID         string
	Phone           string
	Address         string
	FullName        string
	NamaSuci        string
	Agama           string
	NRP             string
	NoAK            string
	PangkatTerakhir string
}

type ClaimWithGoogleInput struct {
	MemberNumber    string
	Token           string
	IDToken         string
	KordaID         string
	Phone           string
	Address         string
	FullName        string
	NamaSuci        string
	Agama           string
	NRP             string
	NoAK            string
	PangkatTerakhir string
}

type Service interface {
	AdminGenerateToken(ctx context.Context, memberID, adminID, purpose string) (string, error)
	CheckClaimToken(ctx context.Context, memberNumber, token string) (*member.Member, error)
	ClaimWithPassword(ctx context.Context, in ClaimInput) (*member.LoginResult, error)
	ClaimWithGoogle(ctx context.Context, in ClaimWithGoogleInput) (*member.LoginResult, error)
	ResetPasswordWithToken(ctx context.Context, memberNumber, token, newPassword string) error
}

var ErrTooManyAttempts = errors.New("terlalu banyak percobaan gagal, coba lagi dalam 1 jam")

const maxFailedAttempts = 5
const attemptWindow = 1 * time.Hour

type service struct {
	tokenRepo      Repository
	pendingRepo    PendingChangeRepository
	memberRepo     member.Repository
	storage        storage.Storage
	jwtSecret      string
	jwtExpiry      time.Duration
	googleClientID string
}

func NewService(
	tokenRepo Repository,
	pendingRepo PendingChangeRepository,
	memberRepo member.Repository,
	storage storage.Storage,
	jwtSecret string,
	jwtExpiry time.Duration,
	googleClientID string,
) Service {
	return &service{
		tokenRepo: tokenRepo, pendingRepo: pendingRepo, memberRepo: memberRepo, storage: storage,
		jwtSecret: jwtSecret, jwtExpiry: jwtExpiry, googleClientID: googleClientID,
	}
}

func generateToken() string {
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

func (s *service) AdminGenerateToken(ctx context.Context, memberID, adminID, purpose string) (string, error) {
	m, err := s.memberRepo.FindByID(ctx, memberID)
	if err != nil {
		return "", err
	}
	if m.Status == member.StatusDeceased {
		return "", ErrMemberDeceased
	}

	raw := generateToken()
	hash, err := bcrypt.GenerateFromPassword([]byte(raw), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	t := &ActivationToken{
		ID: ulid.Make().String(), MemberID: memberID, TokenHash: string(hash),
		Purpose: purpose, ExpiresAt: time.Now().Add(tokenExpiry), CreatedByAdminID: adminID,
	}
	if err := s.tokenRepo.Create(ctx, t); err != nil {
		return "", err
	}
	return raw, nil
}

func (s *service) validateToken(ctx context.Context, memberNumber, token, purpose string) (*member.Member, *ActivationToken, error) {
	failedCount, err := s.tokenRepo.CountRecentFailedAttempts(ctx, memberNumber, purpose, time.Now().Add(-attemptWindow))
	if err != nil {
		return nil, nil, err
	}
	if failedCount >= maxFailedAttempts {
		return nil, nil, ErrTooManyAttempts
	}

	m, err := s.memberRepo.FindByMemberNumber(ctx, memberNumber)
	if err != nil {
		_ = s.tokenRepo.RecordFailedAttempt(ctx, memberNumber, purpose)
		return nil, nil, ErrInvalidToken
	}

	t, err := s.tokenRepo.FindLatestValidByMemberID(ctx, m.ID, purpose)
	if err != nil {
		_ = s.tokenRepo.RecordFailedAttempt(ctx, memberNumber, purpose)
		return nil, nil, ErrInvalidToken
	}

	if err := bcrypt.CompareHashAndPassword([]byte(t.TokenHash), []byte(token)); err != nil {
		_ = s.tokenRepo.RecordFailedAttempt(ctx, memberNumber, purpose)
		return nil, nil, ErrInvalidToken
	}

	return m, t, nil
}

func (s *service) CheckClaimToken(ctx context.Context, memberNumber, token string) (*member.Member, error) {
	m, _, err := s.validateToken(ctx, memberNumber, token, PurposeClaim)
	if err != nil {
		return nil, err
	}
	if m.Status == member.StatusDeceased {
		return nil, ErrMemberDeceased
	}
	if m.Status == member.StatusActive {
		return nil, ErrAlreadyClaimed
	}
	return m, nil
}

func applyOrQueue(changes map[string]interface{}, field string, current *string, newVal string) *string {
	if newVal == "" {
		return current
	}
	if current == nil || *current == "" {
		v := newVal
		return &v
	}
	if *current != newVal {
		changes[field] = newVal
	}
	return current
}

func (s *service) ClaimWithPassword(ctx context.Context, in ClaimInput) (*member.LoginResult, error) {
	if in.Username == "" || in.Password == "" || in.FullName == "" || in.KordaID == "" {
		return nil, ErrValidation
	}
	if len(in.Password) < 8 {
		return nil, errors.New("password minimal 8 karakter")
	}

	m, t, err := s.validateToken(ctx, in.MemberNumber, in.Token, PurposeClaim)
	if err != nil {
		return nil, err
	}
	if m.Status == member.StatusDeceased {
		return nil, ErrMemberDeceased
	}
	if m.Status == member.StatusActive {
		return nil, ErrAlreadyClaimed
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	hashStr := string(hash)

	changes := map[string]interface{}{}
	if in.FullName != "" && in.FullName != m.FullName {
		changes["full_name"] = in.FullName
	}
	m.NamaSuci = applyOrQueue(changes, "nama_suci", m.NamaSuci, in.NamaSuci)
	m.Agama = applyOrQueue(changes, "agama", m.Agama, in.Agama)
	m.NRP = applyOrQueue(changes, "nrp", m.NRP, in.NRP)
	m.NoAK = applyOrQueue(changes, "no_ak", m.NoAK, in.NoAK)
	m.PangkatTerakhir = applyOrQueue(changes, "pangkat_terakhir", m.PangkatTerakhir, in.PangkatTerakhir)

	if len(changes) > 0 {
		if err := s.pendingRepo.Create(ctx, m.ID, changes); err != nil {
			return nil, err
		}
	}

	var phonePtr, addressPtr *string
	if in.Phone != "" {
		phonePtr = &in.Phone
	}
	if in.Address != "" {
		addressPtr = &in.Address
	}
	kordaID := in.KordaID
	username := in.Username

	m.Username = &username
	m.PasswordHash = &hashStr
	m.KordaID = &kordaID
	m.Phone = phonePtr
	m.Address = addressPtr
	m.Status = member.StatusActive

	if err := s.memberRepo.ClaimAccount(ctx, m); err != nil {
		return nil, err
	}
	_ = s.tokenRepo.MarkUsed(ctx, t.ID)

	tokenStr, err := jwt.GenerateMemberToken(s.jwtSecret, m.ID, s.jwtExpiry)
	if err != nil {
		return nil, err
	}
	return &member.LoginResult{Token: tokenStr}, nil
}

func (s *service) ClaimWithGoogle(ctx context.Context, in ClaimWithGoogleInput) (*member.LoginResult, error) {
	if in.FullName == "" || in.KordaID == "" || in.IDToken == "" {
		return nil, ErrValidation
	}

	m, t, err := s.validateToken(ctx, in.MemberNumber, in.Token, PurposeClaim)
	if err != nil {
		return nil, err
	}
	if m.Status == member.StatusDeceased {
		return nil, ErrMemberDeceased
	}
	if m.Status == member.StatusActive {
		return nil, ErrAlreadyClaimed
	}

	payload, err := idtoken.Validate(ctx, in.IDToken, s.googleClientID)
	if err != nil {
		return nil, errors.New("token Google tidak valid")
	}
	googleID := payload.Subject
	emailAddr, _ := payload.Claims["email"].(string)
	picture, _ := payload.Claims["picture"].(string)

	if emailAddr == "" {
		return nil, errors.New("email tidak ditemukan pada akun Google")
	}

	if existing, err := s.memberRepo.FindByGoogleID(ctx, googleID); err == nil && existing.ID != m.ID {
		return nil, ErrGoogleAlreadyUsed
	}
	if existing, err := s.memberRepo.FindByEmail(ctx, emailAddr); err == nil && existing.ID != m.ID {
		return nil, ErrEmailAlreadyUsed
	}

	var avatarPtr *string
	if picture != "" {
		if uploadedURL, uploadErr := s.storage.UploadFromURL(ctx, picture, "member-avatar"); uploadErr == nil {
			avatarPtr = &uploadedURL
		} else {
			log.Error().Err(uploadErr).Msg("gagal mengunduh foto profil Google saat klaim akun")
		}
	}

	changes := map[string]interface{}{}
	if in.FullName != "" && in.FullName != m.FullName {
		changes["full_name"] = in.FullName
	}
	m.NamaSuci = applyOrQueue(changes, "nama_suci", m.NamaSuci, in.NamaSuci)
	m.Agama = applyOrQueue(changes, "agama", m.Agama, in.Agama)
	m.NRP = applyOrQueue(changes, "nrp", m.NRP, in.NRP)
	m.NoAK = applyOrQueue(changes, "no_ak", m.NoAK, in.NoAK)
	m.PangkatTerakhir = applyOrQueue(changes, "pangkat_terakhir", m.PangkatTerakhir, in.PangkatTerakhir)

	if len(changes) > 0 {
		if err := s.pendingRepo.Create(ctx, m.ID, changes); err != nil {
			return nil, err
		}
	}

	var phonePtr, addressPtr *string
	if in.Phone != "" {
		phonePtr = &in.Phone
	}
	if in.Address != "" {
		addressPtr = &in.Address
	}
	kordaID := in.KordaID

	m.Email = emailAddr
	m.GoogleID = &googleID
	m.AvatarURL = avatarPtr
	m.KordaID = &kordaID
	m.Phone = phonePtr
	m.Address = addressPtr
	m.Status = member.StatusActive

	if err := s.memberRepo.ClaimAccount(ctx, m); err != nil {
		return nil, err
	}
	_ = s.tokenRepo.MarkUsed(ctx, t.ID)

	tokenStr, err := jwt.GenerateMemberToken(s.jwtSecret, m.ID, s.jwtExpiry)
	if err != nil {
		return nil, err
	}
	return &member.LoginResult{Token: tokenStr}, nil
}

func (s *service) ResetPasswordWithToken(ctx context.Context, memberNumber, token, newPassword string) error {
	if len(newPassword) < 8 {
		return errors.New("password minimal 8 karakter")
	}

	m, t, err := s.validateToken(ctx, memberNumber, token, PurposeResetPassword)
	if err != nil {
		return err
	}
	if m.PasswordHash == nil {
		return ErrNotClaimedYet
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	if err := s.memberRepo.UpdatePassword(ctx, m.ID, string(hash)); err != nil {
		return err
	}
	_ = s.tokenRepo.MarkUsed(ctx, t.ID)
	return nil
}
