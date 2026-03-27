package auth

import (
	"time"

	"blog_api/internal/domain"
)

type Identity struct {
	UserID uint64
	Role   domain.Role
	Email  string
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}
