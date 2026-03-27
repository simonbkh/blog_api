package media

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"

	"blog_api/internal/application"
	"blog_api/internal/application/auth"
	"blog_api/internal/domain"
)

var allowedTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
}

type Service struct {
	repo    domain.MediaRepository
	storage domain.FileStorage
	maxSize int64 // bytes
}

func NewService(repo domain.MediaRepository, storage domain.FileStorage, maxSizeMB int) *Service {
	return &Service{
		repo:    repo,
		storage: storage,
		maxSize: int64(maxSizeMB) << 20,
	}
}

func (s *Service) Upload(ctx context.Context, identity auth.Identity, file multipart.File, header *multipart.FileHeader) (*domain.Media, error) {
	if header.Size > s.maxSize {
		return nil, fmt.Errorf("file too large (max %d MB): %w", s.maxSize>>20, application.ErrValidation)
	}

	ct := header.Header.Get("Content-Type")
	if ct == "" {
		ct = inferContentType(header.Filename)
	}
	if !allowedTypes[ct] {
		return nil, fmt.Errorf("unsupported file type %q: %w", ct, application.ErrValidation)
	}

	storedName, err := s.storage.Save(ctx, header.Filename, file)
	if err != nil {
		return nil, fmt.Errorf("store file: %w", err)
	}

	media := &domain.Media{
		OriginalName: sanitizeName(header.Filename),
		StoredName:   storedName,
		ContentType:  ct,
		SizeBytes:    header.Size,
		UploaderID:   identity.UserID,
	}
	if err := s.repo.Create(ctx, media); err != nil {
		_ = s.storage.Delete(ctx, storedName)
		return nil, err
	}
	return media, nil
}

func (s *Service) Get(ctx context.Context, id uint64) (*domain.Media, error) {
	m, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, application.ErrNotFound
	}
	return m, nil
}

func sanitizeName(name string) string {
	return filepath.Base(strings.ReplaceAll(name, "..", ""))
}

func inferContentType(filename string) string {
	switch strings.ToLower(filepath.Ext(filename)) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	default:
		return "application/octet-stream"
	}
}

// LimitedReader wraps a reader to enforce a max byte limit.
func LimitedReader(r io.Reader, maxBytes int64) io.Reader {
	return io.LimitReader(r, maxBytes+1)
}
