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

	// Collect statistics
	var stats struct {
		UsersExtracted    int
		UsersCreated      int
		ContactsExtracted int
		ContactsCreated   int
		MessagesExtracted int
		MessagesCreated   int
	}

	// Extract users
	users, err := extractor.ExtractUsers()
	if err != nil {
		return fmt.Errorf("failed to extract users: %w", err)
	}
	stats.UsersExtracted = len(users)

	log.Info("Extracted users", zap.Int("count", stats.UsersExtracted))

	// Process users
	for _, user := range users {
		if !dryRun {
			err = userRepo.Create(ctx, user)
			if err != nil {
				log.Warn("Failed to create user",
					zap.Error(err),
					zap.String("username", user.Username),
					zap.String("user_id", user.UserID))
				continue
			}
			stats.UsersCreated++
		}
	}

	// Extract contacts
	contacts, err := extractor.ExtractContacts()
	if err != nil {
		return fmt.Errorf("failed to extract contacts: %w", err)
	}
	stats.ContactsExtracted = len(contacts)

	log.Info("Extracted contacts", zap.Int("count", stats.ContactsExtracted))

	// Process contacts
	for _, contact := range contacts {
		if !dryRun {
			err = contactRepo.Create(ctx, contact)
			if err != nil {
				log.Warn("Failed to create contact",
					zap.Error(err),
					zap.String("user_id", contact.UserID),
					zap.String("contact_pubkey", contact.ContactPubKey))
				continue
			}
			stats.ContactsCreated++
		}
	}

	// Extract messages
	messages, err := extractor.ExtractMessages()
	if err != nil {
		return fmt.Errorf("failed to extract messages: %w", err)
	}
	stats.MessagesExtracted = len(messages)

	log.Info("Extracted messages", zap.Int("count", stats.MessagesExtracted))

	// Process messages
	for _, message := range messages {
		if !dryRun {
			err = messageRepo.Create(ctx, message)
			if err != nil {
				log.Warn("Failed to create message",
					zap.Error(err),
					zap.String("message_id", message.MessageID.String()))
				continue
			}
			stats.MessagesCreated++
		}
	}

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
