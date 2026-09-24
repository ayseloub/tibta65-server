package gpheroslide

import "time"

type GPHeroSlide struct {
	ID                  string    `db:"id" json:"id"`
	Headline            string    `db:"headline" json:"headline"`
	Description         *string   `db:"description" json:"description"`
	PrimaryButtonText   *string   `db:"primary_button_text" json:"primary_button_text"`
	PrimaryButtonLink   *string   `db:"primary_button_link" json:"primary_button_link"`
	SecondaryButtonText *string   `db:"secondary_button_text" json:"secondary_button_text"`
	SecondaryButtonLink *string   `db:"secondary_button_link" json:"secondary_button_link"`
	ImageURL            string    `db:"image_url" json:"image_url"`
	Position            int       `db:"position" json:"position"`
	Status              string    `db:"status" json:"status"`
	CreatedAt           time.Time `db:"created_at" json:"created_at"`
	UpdatedAt           time.Time `db:"updated_at" json:"updated_at"`
}

const (
	StatusDraft     = "draft"
	StatusPublished = "published"
)
