package memberactivation

import "time"

const (
	PurposeClaim         = "claim"
	PurposeResetPassword = "reset_password"
)

type ActivationToken struct {
	ID               string     `db:"id"`
	MemberID         string     `db:"member_id"`
	TokenHash        string     `db:"token_hash"`
	Purpose          string     `db:"purpose"`
	ExpiresAt        time.Time  `db:"expires_at"`
	CreatedByAdminID string     `db:"created_by_admin_id"`
	UsedAt           *time.Time `db:"used_at"`
	CreatedAt        time.Time  `db:"created_at"`
}
