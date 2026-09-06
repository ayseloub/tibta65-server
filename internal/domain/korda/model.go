package korda

import "time"

type Korda struct {
	ID        string    `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	Phone     *string   `db:"phone" json:"phone"`
	AdminName *string   `db:"admin_name" json:"admin_name"`
	Address   *string   `db:"address" json:"address"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}
