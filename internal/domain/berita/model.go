package berita

import (
	"time"

	"github.com/lib/pq"
)

type Berita struct {
	ID                string        `db:"id" json:"id"`
	Title             string        `db:"title" json:"title"`
	Slug              string        `db:"slug" json:"slug"`
	Description       string        `db:"description" json:"description"`
	ImageURL          string        `db:"image_url" json:"image_url"`
	Visibility        string        `db:"visibility" json:"visibility"`
	Status            string        `db:"status" json:"status"`
	IsHighlight       bool          `db:"is_highlight" json:"is_highlight"`
	AuthorID          *string       `db:"author_id" json:"author_id"`
	AuthorName        string        `db:"author_name" json:"author_name"`
	EventDate         time.Time     `db:"event_date" json:"event_date"`
	TargetGenerations pq.Int64Array `db:"target_generations" json:"target_generations"`
	CreatedAt         time.Time     `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time     `db:"updated_at" json:"updated_at"`
	PublishAt         *time.Time    `db:"publish_at" json:"publish_at"`
	ExpireAt          *time.Time    `db:"expire_at" json:"expire_at"`
}

const (
	VisibilityPublic   = "public"
	VisibilityInternal = "internal"

	StatusDraft     = "draft"
	StatusPublished = "published"
)
