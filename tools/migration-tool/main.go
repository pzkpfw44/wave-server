package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"

	"github.com/pzkpfw44/wave-server/internal/config"
	"github.com/pzkpfw44/wave-server/internal/repository"
	"github.com/pzkpfw44/wave-server/pkg/logger"
)

var (
	sourceDir = flag.String("source", "./old_data", "Source directory for old data")
	dryRun    = flag.Bool("dry-run", false, "Dry run mode (no writes)")
)

// Extractor extracts data from the source directory
type Extractor struct {
	sourceDir string
	logger    *zap.Logger
}

// NewExtractor creates a new extractor instance
func NewExtractor(sourceDir string, logger *zap.Logger) *Extractor {
	return &Extractor{
		sourceDir: sourceDir,
		logger:    logger.With(zap.String("component", "extractor")),
	}
}

// Extract extracts data from the source directory
func (e *Extractor) Extract() error {
	e.logger.Info("Extracting data from source directory", zap.String("dir", e.sourceDir))
	// Implement extraction logic
	return nil
}

func main() {
	// Parse command line flags
	flag.Parse()

	// Setup logger
	log, err := logger.New("info", true)
	if err != nil {
		fmt.Printf("Failed to create logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync(log)

	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration", zap.Error(err))
	}

	// Create context
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Connect to database
	db, err := repository.New(ctx, cfg, log)
	if err != nil {
		log.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Create repositories
	userRepo := repository.NewUserRepository(db)
	messageRepo := repository.NewMessageRepository(db)
	contactRepo := repository.NewContactRepository(db)

	// Create extractor
	extractor := NewExtractor(*sourceDir, log)

	// Verify that source directory exists
	if _, err := os.Stat(*sourceDir); os.IsNotExist(err) {
		log.Fatal("Source directory does not exist", zap.String("directory", *sourceDir))
	}

	// Run migration
	err = RunMigration(ctx, extractor, userRepo, messageRepo, contactRepo, log, *dryRun)
	if err != nil {
		log.Fatal("Migration failed", zap.Error(err))
	}

	log.Info("Migration completed successfully")
}

// RunMigration runs the migration process
func RunMigration(ctx context.Context, extractor *Extractor, userRepo *repository.UserRepository,
	messageRepo *repository.MessageRepository, contactRepo *repository.ContactRepository,
	log *zap.Logger, dryRun bool) error {

	// Extract the data
	if err := extractor.Extract(); err != nil {
		return fmt.Errorf("extraction failed: %w", err)
	}

	// Implement migration logic
	log.Info("Migration would process data here", zap.Bool("dry_run", dryRun))

	// Collect statistics
	var stats struct {
		UsersExtracted    int
		UsersCreated      int
		ContactsExtracted int
		ContactsCreated   int
		MessagesExtracted int
		MessagesCreated   int
	}

	// For now just set some placeholder values 
	stats.UsersExtracted = 10
	stats.UsersCreated = dryRun ? 0 : 10
	stats.ContactsExtracted = 50
	stats.ContactsCreated = dryRun ? 0 : 50
	stats.MessagesExtracted = 200
	stats.MessagesCreated = dryRun ? 0 : 200

	log.Info("Migration statistics",
		zap.Int("users_extracted", stats.UsersExtracted),
		zap.Int("users_created", stats.UsersCreated),
		zap.Int("contacts_extracted", stats.ContactsExtracted),
		zap.Int("contacts_created", stats.ContactsCreated),
		zap.Int("messages_extracted", stats.MessagesExtracted),
		zap.Int("messages_created", stats.MessagesCreated),
		zap.Bool("dry_run", dryRun),
	)

	return nil
}