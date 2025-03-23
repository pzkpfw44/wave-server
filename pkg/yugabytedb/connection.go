package yugabytedb

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// Config defines YugabyteDB-specific configuration
type Config struct {
	Host         string
	Port         int
	User         string
	Password     string
	Database     string
	MaxPoolSize  int
	MinPoolSize  int
	ConnLifetime int // seconds
}

// Connect establishes a connection to YugabyteDB
func Connect(ctx context.Context, cfg Config, logger *zap.Logger) (*pgxpool.Pool, error) {
	// Build connection string
	connString := fmt.Sprintf(
		"postgresql://%s:%s@%s:%d/%s?application_name=wave-server",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database,
	)

	// Create a pool config
	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	// Configure pool settings
	if cfg.MaxPoolSize > 0 {
		poolConfig.MaxConns = int32(cfg.MaxPoolSize)
	}
	if cfg.MinPoolSize > 0 {
		poolConfig.MinConns = int32(cfg.MinPoolSize)
	}

	// Create the connection pool
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test the connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Check if using YugabyteDB
	var version string
	err = pool.QueryRow(ctx, "SELECT version()").Scan(&version)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to get database version: %w", err)
	}

	if !isYugabyteDB(version) {
		logger.Warn("Not connected to YugabyteDB, some features may not work properly",
			zap.String("version", version))
	} else {
		logger.Info("Connected to YugabyteDB", zap.String("version", version))
	}

	return pool, nil
}

// isYugabyteDB checks if the version string indicates YugabyteDB
func isYugabyteDB(version string) bool {
	return strings.Contains(strings.ToLower(version), "yugabytedb") ||
		strings.Contains(strings.ToLower(version), "yb")
}
