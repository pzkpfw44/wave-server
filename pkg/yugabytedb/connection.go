package yugabytedb

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// Connection represents a connection to YugabyteDB
type Connection struct {
	Pool   *pgxpool.Pool
	Logger *zap.Logger
}

// ConnectOptions defines options for connecting to YugabyteDB
type ConnectOptions struct {
	Host           string
	Port           int
	User           string
	Password       string
	Database       string
	PoolSize       int
	ConnectTimeout time.Duration
	SSLMode        string
}

// Connect creates a new connection to YugabyteDB
func Connect(ctx context.Context, opts ConnectOptions, logger *zap.Logger) (*Connection, error) {
	// Build connection string
	connString := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s&pool_max_conns=%d",
		opts.User,
		opts.Password,
		opts.Host,
		opts.Port,
		opts.Database,
		opts.SSLMode,
		opts.PoolSize,
	)

	if opts.ConnectTimeout > 0 {
		connString = fmt.Sprintf("%s&connect_timeout=%d", connString, int(opts.ConnectTimeout.Seconds()))
	}

	// Create connection pool
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	// YugabyteDB-specific optimizations
	config.MaxConns = int32(opts.PoolSize)
	config.MinConns = 1
	config.MaxConnLifetime = 1 * time.Hour
	config.MaxConnIdleTime = 30 * time.Minute
	config.HealthCheckPeriod = 30 * time.Second // Increased for YugabyteDB

	// Create connection pool
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Connected to YugabyteDB",
		zap.String("host", opts.Host),
		zap.Int("port", opts.Port),
		zap.String("database", opts.Database),
		zap.Int("pool_size", opts.PoolSize),
	)

	return &Connection{
		Pool:   pool,
		Logger: logger.With(zap.String("component", "yugabytedb")),
	}, nil
}

// CheckIsYugabyteDB checks if the database is a YugabyteDB instance
func (c *Connection) CheckIsYugabyteDB(ctx context.Context) (bool, error) {
	// Try to query YugabyteDB-specific system tables
	var count int
	err := c.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM pg_catalog.pg_class WHERE relname = 'yb_servers'").Scan(&count)
	if err != nil {
		c.Logger.Warn("Failed to check if database is YugabyteDB", zap.Error(err))
		return false, nil
	}

	isYugabyte := count > 0
	if isYugabyte {
		c.Logger.Info("Connected to YugabyteDB instance")
	} else {
		c.Logger.Info("Connected to PostgreSQL instance (not YugabyteDB)")
	}

	return isYugabyte, nil
}

// Close closes the connection pool
func (c *Connection) Close() {
	if c.Pool != nil {
		c.Pool.Close()
		c.Logger.Info("Closed YugabyteDB connection pool")
	}
}
