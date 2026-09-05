package member

import "time"

type Member struct {
	ID                 string     `db:"id" json:"id"`
	FullName           string     `db:"full_name" json:"full_name"`
	Email              string     `db:"email" json:"email"`
	Phone              *string    `db:"phone" json:"phone"`
	Address            *string    `db:"address" json:"address"`
	MemberNumber       string     `db:"member_number" json:"member_number"`
	PasswordHash       *string    `db:"password_hash" json:"-"`
	GoogleID           *string    `db:"google_id" json:"-"`
	AvatarURL          *string    `db:"avatar_url" json:"avatar_url"`
	KordaID            *string    `db:"korda_id" json:"korda_id"`
	MustChangePassword bool       `db:"must_change_password" json:"must_change_password"`
	EmailVerifiedAt    *time.Time `db:"email_verified_at" json:"email_verified_at"`
	CreatedAt          time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time  `db:"updated_at" json:"updated_at"`
	KordaName          *string    `db:"korda_name" json:"korda_name"`
}
