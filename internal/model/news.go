package model

import "time"

type NewsStatus string

const (
	NewsStatusDraft     NewsStatus = "draft"
	NewsStatusPublished NewsStatus = "published"
)

type NewsCategory struct {
	ID        int64     `json:"id"`
	Slug      string    `json:"slug"`
	NameIDN   string    `json:"name_idn"`
	NameEN    string    `json:"name_en"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type News struct {
	ID          int64      `json:"id"`
	CategoryID  int64      `json:"category_id"`
	AuthorID    int64      `json:"author_id"`
	Slug        string     `json:"slug"`
	Thumbnail   *string    `json:"thumbnail"`
	TitleIDN    string     `json:"title_idn"`
	TitleEN     string     `json:"title_en"`
	ContentIDN  string     `json:"content_idn"`
	ContentEN   string     `json:"content_en"`
	Status      NewsStatus `json:"status"`
	PublishedAt *time.Time `json:"published_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}
