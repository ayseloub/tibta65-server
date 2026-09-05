package pemilu

import "time"

type Kandidat struct {
	ID        string    `db:"id" json:"id"`
	FullName  string    `db:"full_name" json:"full_name"`
	Visi      string    `db:"visi" json:"visi"`
	Misi      string    `db:"misi" json:"misi"`
	Pangkat   string    `db:"pangkat" json:"pangkat"`
	VoteCount int       `db:"vote_count" json:"vote_count"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type Settings struct {
	ID            string     `db:"id" json:"id"`
	StartAt       time.Time  `db:"start_at" json:"start_at"`
	EndAt         time.Time  `db:"end_at" json:"end_at"`
	ClosedEarlyAt *time.Time `db:"closed_early_at" json:"closed_early_at"`
	CreatedAt     time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at" json:"updated_at"`
}
