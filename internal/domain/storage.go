package domain

import (
	"context"
	"io"
)

// FileStorage abstracts file persistence so the implementation
// (local disk, S3, GCS, …) can be swapped without touching business logic.
type FileStorage interface {
	Save(ctx context.Context, name string, r io.Reader) (storedPath string, err error)
	Delete(ctx context.Context, storedPath string) error
}
