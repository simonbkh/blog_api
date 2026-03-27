package domain

import "time"

type PostStatus string

const (
	PostStatusDraft     PostStatus = "draft"
	PostStatusPublished PostStatus = "published"
)

type Post struct {
	ID        uint64
	Title     string
	Content   string
	AuthorID  uint64
	Images    []uint64
	Status    PostStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (p Post) IsValidStatus() bool {
	return p.Status == PostStatusDraft || p.Status == PostStatusPublished
}
