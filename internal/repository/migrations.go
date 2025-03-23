package repository

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"go.uber.org/zap"
)

// We'll use Go 1.16+ embed feature for migration files
// This comment would normally be used with //go:embed migrations/*.sql
// but since we're not embedding anything in this fix, we're using a dummy FS

// MigrationsFS contains the embedded SQL migrations
var MigrationsFS embed.FS

// GetMigrationFiles returns a sorted list of migration files
func GetMigrationFiles() ([]string, error) {
	// Check if the migrations directory exists in the current directory
	migrationDir := "migrations"

	// If running locally, read from filesystem
	files, err := os.ReadDir(migrationDir)
	if err == nil {
		result := make([]string, 0, len(files))
		for _, file := range files {
			if !file.IsDir() && strings.HasSuffix(file.Name(), ".sql") {
				result = append(result, filepath.Join(migrationDir, file.Name()))
			}
		}
		sort.Strings(result)
		return result, nil
	}

	// If running from binary, use embedded files
	entries, err := fs.ReadDir(MigrationsFS, migrationDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}

	result := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			result = append(result, filepath.Join(migrationDir, entry.Name()))
		}
	}

	sort.Strings(result)
	return result, nil
}

// RunMigrationsFromFS applies migrations from the embedded file system
func (db *Database) RunMigrationsFromFS(ctx context.Context) error {
	db.Logger.Info("Running migrations from embedded file system")

	files, err := GetMigrationFiles()
	if err != nil {
		return fmt.Errorf("failed to get migration files: %w", err)
	}

	for _, file := range files {
		// Skip non-up migrations
		if !strings.Contains(file, ".up.") {
			continue
		}

		db.Logger.Info("Applying migration", zap.String("file", file))

		// Read migration file
		var migrationContent []byte

		// Try reading from filesystem first, then embedded FS
		if content, err := os.ReadFile(file); err == nil {
			migrationContent = content
		} else {
			content, err := MigrationsFS.ReadFile(file)
			if err != nil {
				return fmt.Errorf("failed to read migration file %s: %w", file, err)
			}
			migrationContent = content
		}

		// Execute the migration
		_, err := db.Pool.Exec(ctx, string(migrationContent))
		if err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", file, err)
		}

		db.Logger.Info("Migration applied successfully", zap.String("file", file))
	}

	return nil
}
