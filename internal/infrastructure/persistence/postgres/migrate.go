package postgres

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"gorm.io/gorm"
)

func RunMigrations(db *gorm.DB, migrationsDir string) error {
	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
	if err != nil {
		return err
	}
	sort.Strings(files)
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		if _, err := sqlDB.Exec(string(content)); err != nil {
			return fmt.Errorf("migration %s failed: %w", filepath.Base(file), err)
		}
	}
	return nil
}
