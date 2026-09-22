package member

import "time"

type Member struct {
	ID                  string     `db:"id" json:"id"`
	FullName            string     `db:"full_name" json:"full_name"`
	Email               string     `db:"email" json:"email"`
	Phone               *string    `db:"phone" json:"phone"`
	Address             *string    `db:"address" json:"address"`
	LegacyMemberNumber  *string    `db:"legacy_member_number" json:"legacy_member_number"`
	MemberNumber        string     `db:"member_number" json:"member_number"`
	Username            *string    `db:"username" json:"username"`
	Generation          int        `db:"generation" json:"generation"`
	ParentMemberID      *string    `db:"parent_member_id" json:"parent_member_id"`
	Status              string     `db:"status" json:"status"`
	NamaSuci            *string    `db:"nama_suci" json:"nama_suci"`
	Agama               *string    `db:"agama" json:"agama"`
	NRP                 *string    `db:"nrp" json:"nrp"`
	NoAK                *string    `db:"no_ak" json:"no_ak"`
	PangkatTerakhir     *string    `db:"pangkat_terakhir" json:"pangkat_terakhir"`
	LegacyIdentifierRaw *string    `db:"legacy_identifier_raw" json:"legacy_identifier_raw"`
	PasswordHash        *string    `db:"password_hash" json:"-"`
	GoogleID            *string    `db:"google_id" json:"-"`
	AvatarURL           *string    `db:"avatar_url" json:"avatar_url"`
	KordaID             *string    `db:"korda_id" json:"korda_id"`
	MustChangePassword  bool       `db:"must_change_password" json:"must_change_password"`
	EmailVerifiedAt     *time.Time `db:"email_verified_at" json:"email_verified_at"`
	CreatedAt           time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt           time.Time  `db:"updated_at" json:"updated_at"`
	ProfileCompleted    bool       `db:"profile_completed" json:"profile_completed"`
	KordaName           *string    `db:"korda_name" json:"korda_name"`
	ApprovedAt          *time.Time `db:"approved_at" json:"approved_at"`
}

const (
	StatusUnclaimed     = "unclaimed"
	StatusPendingReview = "pending_review"
	StatusActive        = "active"
	StatusDeceased      = "deceased"
	StatusRejected      = "rejected"
)
