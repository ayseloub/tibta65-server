package auth

import "time"

type Admin struct {
	ID                 string    `db:"id"`
	Username           string    `db:"username"`
	PasswordHash       string    `db:"password_hash"`
	Role               string    `db:"role"`
	MustChangePassword bool      `db:"must_change_password"`
	CreatedAt          time.Time `db:"created_at"`
	UpdatedAt          time.Time `db:"updated_at"`
}

const (
	RoleSuperAdmin = "superadmin"
	RoleAdmin      = "admin"
)
