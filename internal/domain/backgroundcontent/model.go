package backgroundcontent

import "time"

type BackgroundContent struct {
	ID          string    `db:"id" json:"id"`
	Section     string    `db:"section" json:"section"`
	Title       string    `db:"title" json:"title"`
	Description string    `db:"description" json:"description"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

const (
	SectionAbout   = "about"
	SectionSejarah = "sejarah"
)

func IsValidSection(section string) bool {
	return section == SectionAbout || section == SectionSejarah
}
