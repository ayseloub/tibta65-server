package ticket

import "time"

type Ticket struct {
	ID               string     `db:"id" json:"id"`
	MemberID         string     `db:"member_id" json:"member_id"`
	MemberName       string     `db:"member_name" json:"member_name"`
	MemberEmail      string     `db:"member_email" json:"member_email"`
	KordaName        *string    `db:"korda_name" json:"korda_name"`
	Subject          string     `db:"subject" json:"subject"`
	Message          string     `db:"message" json:"message"`
	Status           string     `db:"status" json:"status"`
	AdminReply       *string    `db:"admin_reply" json:"admin_reply"`
	RepliedAt        *time.Time `db:"replied_at" json:"replied_at"`
	MemberReadAt     *time.Time `db:"member_read_at" json:"member_read_at"`
	ReportedMemberID *string    `db:"reported_member_id" json:"reported_member_id"`
	ReportedName     *string    `db:"reported_name" json:"reported_name"`
	CreatedAt        time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time  `db:"updated_at" json:"updated_at"`
}

const (
	StatusOpen   = "open"
	StatusClosed = "closed"
)
