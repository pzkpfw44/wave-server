package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"github.com/pzkpfw44/wave-server/internal/config"
	"github.com/pzkpfw44/wave-server/pkg/logger"
	"github.com/pzkpfw44/wave-server/pkg/yugabytedb"
)

var (
	backupDir = flag.String("backup-dir", "./backups", "Directory to store backups")
)

func main() {
	flag.Parse()

	// Load .env file if it exists
	_ = godotenv.Load()

	// Load config
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Setup logger
	log, err := logger.New(cfg.LogLevel, cfg.IsDevelopment())
	if err != nil {
		fmt.Printf("Failed to create logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync(log)

	// Create context
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	// Connect to database
	connString := cfg.GetDSN()
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		log.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer pool.Close()

	// Create backup directory if it doesn't exist
	if _, err := os.Stat(*backupDir); os.IsNotExist(err) {
		if err := os.MkdirAll(*backupDir, 0755); err != nil {
			log.Fatal("Failed to create backup directory", zap.Error(err))
		}
	}

	// Create backup filename with timestamp
	timestamp := time.Now().Format("20060102_150405")
	backupFile := fmt.Sprintf("%s/wave_backup_%s.ybak", *backupDir, timestamp)

	// Perform backup
	if err := yugabytedb.BackupDatabase(ctx, pool, backupFile, log); err != nil {
		log.Fatal("Backup failed", zap.Error(err))
	}

	log.Info("Backup completed successfully", zap.String("backup_file", backupFile))
}
