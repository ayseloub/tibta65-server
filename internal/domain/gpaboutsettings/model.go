package gpaboutsettings

import "time"

type GPAboutSettings struct {
	ID          string    `db:"id" json:"id"`
	Title       string    `db:"title" json:"title"`
	Description string    `db:"description" json:"description"`
	Stat1       string    `db:"stat_1" json:"stat_1"`
	Stat2       string    `db:"stat_2" json:"stat_2"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}
