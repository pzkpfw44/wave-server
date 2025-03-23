package repository

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"path"
	"strings"

	"go.uber.org/zap"
)

//go:embed ../../migrations/*.sql
var migrationsFs embed.FS

// RunMigrations runs database migrations from the embedded migrations directory
func (db *Database) RunMigrations(ctx context.Context) error {
	// Create migrations table if it doesn't exist
	_, err := db.Pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS migrations (
			id SERIAL PRIMARY KEY,
			migration_name VARCHAR(255) NOT NULL UNIQUE,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get already applied migrations
	rows, err := db.Pool.Query(ctx, `SELECT migration_name FROM migrations`)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}
	defer rows.Close()

	appliedMigrations := make(map[string]bool)
	for rows.Next() {
		var migrationName string
		if err := rows.Scan(&migrationName); err != nil {
			return fmt.Errorf("failed to scan migration name: %w", err)
		}
		appliedMigrations[migrationName] = true
	}

	// List migration files
	files, err := fs.ReadDir(migrationsFs, "migrations")
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	// Sort migrations by name
	// Since we follow the naming convention 000001_name.up.sql, this will sort them correctly
	migrationFiles := make([]string, 0, len(files))
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".up.sql") {
			migrationFiles = append(migrationFiles, file.Name())
		}
	}

	// Apply missing migrations
	for _, migrationFile := range migrationFiles {
		migrationName := strings.TrimSuffix(migrationFile, ".up.sql")
		if appliedMigrations[migrationName] {
			db.Logger.Info("Migration already applied", zap.String("migration", migrationName))
			continue
		}

		db.Logger.Info("Applying migration", zap.String("migration", migrationName))

		// Read migration file
		migrationPath := path.Join("migrations", migrationFile)
		migrationContent, err := migrationsFs.ReadFile(migrationPath)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", migrationFile, err)
		}

		// Execute migration
		tx, err := db.Pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}

		_, err = tx.Exec(ctx, string(migrationContent))
		if err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("failed to execute migration %s: %w", migrationName, err)
		}

		// Record migration
		_, err = tx.Exec(ctx, `INSERT INTO migrations (migration_name) VALUES ($1)`, migrationName)
		if err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("failed to record migration %s: %w", migrationName, err)
		}

		// Commit transaction
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", migrationName, err)
		}

		db.Logger.Info("Migration applied successfully", zap.String("migration", migrationName))
	}

	return nil
}
