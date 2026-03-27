package postgres

import "time"

type UserModel struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement"`
	FullName     string    `gorm:"size:120;not null"`
	Email        string    `gorm:"size:255;uniqueIndex;not null"`
	PasswordHash string    `gorm:"size:255;not null"`
	Role         string    `gorm:"size:32;not null;index"`
	CreatedAt    time.Time `gorm:"not null"`
	UpdatedAt    time.Time `gorm:"not null"`
}

func (UserModel) TableName() string {
	return "users"
}

type PostModel struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement"`
	Title     string    `gorm:"size:255;not null"`
	Content   string    `gorm:"type:text;not null"`
	AuthorID  uint64    `gorm:"not null;index"`
	Status    string    `gorm:"size:16;not null;index"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (PostModel) TableName() string {
	return "posts"
}

type RefreshTokenModel struct {
	ID         string     `gorm:"primaryKey;size:128"`
	UserID     uint64     `gorm:"not null;index"`
	TokenHash  string     `gorm:"size:128;uniqueIndex;not null"`
	ExpiresAt  time.Time  `gorm:"not null;index"`
	RevokedAt  *time.Time `gorm:"index"`
	ReplacedBy *string    `gorm:"size:128"`
	CreatedAt  time.Time  `gorm:"not null"`
}

func (RefreshTokenModel) TableName() string {
	return "refresh_tokens"
}

type MediaModel struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement"`
	OriginalName string    `gorm:"size:255;not null"`
	StoredName   string    `gorm:"size:255;uniqueIndex;not null"`
	ContentType  string    `gorm:"size:64;not null"`
	SizeBytes    int64     `gorm:"not null"`
	UploaderID   uint64    `gorm:"not null;index"`
	CreatedAt    time.Time `gorm:"not null"`
}

func (MediaModel) TableName() string {
	return "media"
}

type PostImageModel struct {
	PostID   uint64 `gorm:"primaryKey"`
	MediaID  uint64 `gorm:"primaryKey"`
	Position int    `gorm:"not null;default:0"`
}

func (PostImageModel) TableName() string {
	return "post_images"
}
