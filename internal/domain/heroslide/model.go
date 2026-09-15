package heroslide

import "time"

type HeroSlide struct {
	ID                  string     `db:"id" json:"id"`
	Headline            string     `db:"headline" json:"headline"`
	Description         string     `db:"description" json:"description"`
	PrimaryButtonText   string     `db:"primary_button_text" json:"primary_button_text"`
	PrimaryButtonLink   string     `db:"primary_button_link" json:"primary_button_link"`
	SecondaryButtonText *string    `db:"secondary_button_text" json:"secondary_button_text"`
	SecondaryButtonLink *string    `db:"secondary_button_link" json:"secondary_button_link"`
	ImageURL            string     `db:"image_url" json:"image_url"`
	BgColor             string     `db:"bg_color" json:"bg_color"`
	Position            int        `db:"position" json:"position"`
	Status              string     `db:"status" json:"status"`
	PublishAt           *time.Time `db:"publish_at" json:"publish_at"`
	ExpireAt            *time.Time `db:"expire_at" json:"expire_at"`
	CreatedAt           time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt           time.Time  `db:"updated_at" json:"updated_at"`
}

const (
	StatusDraft     = "draft"
	StatusPublished = "published"
)
