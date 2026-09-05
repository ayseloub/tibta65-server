package member

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/oklog/ulid/v2"
)

const (
	OTPPurposeRegister      = "register"
	OTPPurposeResetPassword = "reset_password"
)

type OTPRepository interface {
	Create(ctx context.Context, memberID, code, purpose string, expiresAt time.Time) error
	FindValidByMemberID(ctx context.Context, memberID, purpose string) (*MemberOTP, error)
	DeleteByMemberID(ctx context.Context, memberID, purpose string) error
}

type MemberOTP struct {
	ID        string    `db:"id"`
	MemberID  string    `db:"member_id"`
	OTPCode   string    `db:"otp_code"`
	Purpose   string    `db:"purpose"`
	ExpiresAt time.Time `db:"expires_at"`
	CreatedAt time.Time `db:"created_at"`
}

type otpRepository struct {
	db *sqlx.DB
}

func NewOTPRepository(db *sqlx.DB) OTPRepository {
	return &otpRepository{db: db}
}

func (r *otpRepository) Create(ctx context.Context, memberID, code, purpose string, expiresAt time.Time) error {
	if err := r.DeleteByMemberID(ctx, memberID, purpose); err != nil {
		return err
	}

	query := `INSERT INTO member_otps (id, member_id, otp_code, purpose, expires_at) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.ExecContext(ctx, query, ulid.Make().String(), memberID, code, purpose, expiresAt)
	return err
}

func (r *otpRepository) FindValidByMemberID(ctx context.Context, memberID, purpose string) (*MemberOTP, error) {
	var otp MemberOTP
	query := `
		SELECT id, member_id, otp_code, purpose, expires_at, created_at
		FROM member_otps
		WHERE member_id = $1 AND purpose = $2 AND expires_at > now()
		ORDER BY created_at DESC
		LIMIT 1
	`
	err := r.db.GetContext(ctx, &otp, query, memberID, purpose)
	if err != nil {
		return nil, err
	}
	return &otp, nil
}

func (r *otpRepository) DeleteByMemberID(ctx context.Context, memberID, purpose string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM member_otps WHERE member_id = $1 AND purpose = $2", memberID, purpose)
	return err
}
