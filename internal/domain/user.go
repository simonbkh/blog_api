package domain

import "time"

type User struct {
	ID           uint64
	FullName     string
	Email        string
	PasswordHash string
	Role         Role
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
