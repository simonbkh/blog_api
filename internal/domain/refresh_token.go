package domain

import "time"

type RefreshTokenSession struct {
	ID         string
	UserID     uint64
	TokenHash  string
	ExpiresAt  time.Time
	RevokedAt  *time.Time
	ReplacedBy *string
	CreatedAt  time.Time
}

func (s RefreshTokenSession) IsRevoked() bool {
	return s.RevokedAt != nil
}
