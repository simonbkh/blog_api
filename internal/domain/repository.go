package domain

import (
	"context"
	"time"
)

type PostRepository interface {
	List(ctx context.Context) ([]Post, error)
	ListByAuthor(ctx context.Context, authorID uint64) ([]Post, error)
	GetByID(ctx context.Context, id uint64) (*Post, error)
	Create(ctx context.Context, post *Post) error
	Update(ctx context.Context, post *Post) error
	Delete(ctx context.Context, id uint64) error
}

type UserRepository interface {
	GetByID(ctx context.Context, id uint64) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Create(ctx context.Context, user *User) error
}

type RefreshTokenRepository interface {
	Create(ctx context.Context, token *RefreshTokenSession) error
	GetByTokenHash(ctx context.Context, tokenHash string) (*RefreshTokenSession, error)
	Rotate(ctx context.Context, oldTokenID string, newToken *RefreshTokenSession, revokedAt time.Time) error
	RevokeByTokenID(ctx context.Context, tokenID string, revokedAt time.Time) error
}
