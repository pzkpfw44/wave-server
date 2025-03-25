package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/pzkpfw44/wave-server/internal/config"
)

// Database represents a connection to the database
type Database struct {
	Pool   *pgxpool.Pool
	Logger *zap.Logger
	Config *config.Config
}

// New creates a new database connection
func New(ctx context.Context, cfg *config.Config, logger *zap.Logger) (*Database, error) {
	// Create connection pool config
	poolConfig, err := pgxpool.ParseConfig(cfg.GetDSN())
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	// Set connection pool size
	poolConfig.MaxConns = int32(cfg.Database.PoolSize)

	// Increase health check timeout for YugabyteDB
	poolConfig.HealthCheckPeriod = 30 * time.Second

	// Create connection pool
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Connected to database",
		zap.String("host", cfg.Database.Host),
		zap.Int("port", cfg.Database.Port),
		zap.String("database", cfg.Database.Name),
		zap.Int("pool_size", cfg.Database.PoolSize))

	return &Database{
		Pool:   pool,
		Logger: logger,
		Config: cfg,
	}, nil
}

// Close closes the database connection
func (db *Database) Close() {
	if db.Pool != nil {
		db.Pool.Close()
		db.Logger.Info("Closed database connection")
	}
}

// RunMigrations runs database migrations
func (db *Database) RunMigrations(ctx context.Context) error {
	// Initial schema
	initialSchema := `
CREATE TABLE IF NOT EXISTS users (
    user_id VARCHAR(64) PRIMARY KEY,
    username VARCHAR(64) UNIQUE NOT NULL,
    public_key BYTEA NOT NULL,
    encrypted_private_key BYTEA NOT NULL,
    salt BYTEA NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_active TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_last_active ON users(last_active);

CREATE TABLE IF NOT EXISTS messages (
    message_id UUID PRIMARY KEY,
    sender_pubkey VARCHAR(1200) NOT NULL,
    recipient_pubkey VARCHAR(1200) NOT NULL,
    ciphertext_kem BYTEA NOT NULL,
    ciphertext_msg BYTEA NOT NULL,
    nonce BYTEA NOT NULL,
    sender_ciphertext_kem BYTEA NOT NULL,
    sender_ciphertext_msg BYTEA NOT NULL,
    sender_nonce BYTEA NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(16) DEFAULT 'sent'
);

CREATE INDEX IF NOT EXISTS idx_messages_recipient ON messages(recipient_pubkey);
CREATE INDEX IF NOT EXISTS idx_messages_sender ON messages(sender_pubkey);
CREATE INDEX IF NOT EXISTS idx_messages_timestamp ON messages(timestamp);
CREATE INDEX IF NOT EXISTS idx_messages_conversation ON messages((
    CASE WHEN sender_pubkey < recipient_pubkey
        THEN sender_pubkey || recipient_pubkey
        ELSE recipient_pubkey || sender_pubkey
    END
));

CREATE TABLE IF NOT EXISTS contacts (
    user_id VARCHAR(64) NOT NULL,
    contact_pubkey VARCHAR(1200) NOT NULL,
    nickname VARCHAR(64) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, contact_pubkey)
);

CREATE INDEX IF NOT EXISTS idx_contacts_user_id ON contacts(user_id);

CREATE TABLE IF NOT EXISTS tokens (
    token_id UUID PRIMARY KEY,
    user_id VARCHAR(64) NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    token_hash VARCHAR(128) UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    last_used TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_tokens_user_id ON tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_tokens_expires_at ON tokens(expires_at);
    `

	// Execute the migration
	_, err := db.Pool.Exec(ctx, initialSchema)
	if err != nil {
		return fmt.Errorf("failed to run initial migration: %w", err)
	}

	db.Logger.Info("Ran database migrations successfully")
	return nil
}
