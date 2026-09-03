package auth

import "time"

type Admin struct {
	ID                 string     `db:"id" json:"id"`
	Username           string     `db:"username" json:"username"`
	FullName           string     `db:"full_name" json:"full_name"`
	PasswordHash       string     `db:"password_hash" json:"-"`
	Role               string     `db:"role" json:"role"`
	MustChangePassword bool       `db:"must_change_password" json:"must_change_password"`
	LastLoginAt        *time.Time `db:"last_login_at" json:"last_login_at"`
	CreatedAt          time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time  `db:"updated_at" json:"updated_at"`
}

const (
	RoleSuperAdmin = "superadmin"
	RoleAdmin      = "admin"
)
