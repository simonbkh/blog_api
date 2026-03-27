package domain

import "time"

type Media struct {
	ID           uint64
	OriginalName string
	StoredName   string
	ContentType  string
	SizeBytes    int64
	UploaderID   uint64
	CreatedAt    time.Time
}
