package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"crypto/rand"
	"encoding/hex"
)

// LocalStorage stores files on the local filesystem.
type LocalStorage struct {
	baseDir string
}

// NewLocalStorage creates a LocalStorage that writes into baseDir.
// The directory is created if it does not exist.
func NewLocalStorage(baseDir string) (*LocalStorage, error) {
	clean := filepath.Clean(baseDir)
	if err := os.MkdirAll(clean, 0o755); err != nil {
		return nil, fmt.Errorf("create upload dir: %w", err)
	}
	return &LocalStorage{baseDir: clean}, nil
}

// Save writes the contents of r to a unique file inside the base directory.
// It returns the generated filename (not the full path) so the caller can
// build a URL from it.
func (s *LocalStorage) Save(_ context.Context, originalName string, r io.Reader) (string, error) {
	ext := filepath.Ext(originalName)
	stored := uniqueName() + ext

	// Prevent path traversal.
	dest := filepath.Join(s.baseDir, stored)
	if !strings.HasPrefix(filepath.Clean(dest), s.baseDir) {
		return "", fmt.Errorf("invalid file path")
	}

	f, err := os.Create(dest)
	if err != nil {
		return "", fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, r); err != nil {
		os.Remove(dest)
		return "", fmt.Errorf("write file: %w", err)
	}
	return stored, nil
}

// Delete removes a previously stored file.
func (s *LocalStorage) Delete(_ context.Context, storedName string) error {
	dest := filepath.Join(s.baseDir, storedName)
	if !strings.HasPrefix(filepath.Clean(dest), s.baseDir) {
		return fmt.Errorf("invalid file path")
	}
	return os.Remove(dest)
}

func uniqueName() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}
