// rename the file from migrate-data.go to migrate_data.go to resolve package conflicts

package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"github.com/pzkpfw44/wave-server/internal/config"
	"github.com/pzkpfw44/wave-server/internal/domain"
	"github.com/pzkpfw44/wave-server/internal/repository"
	"github.com/pzkpfw44/wave-server/internal/security"
	"github.com/pzkpfw44/wave-server/pkg/logger"
)

var (
	sourceDir = getEnvOrDefault("SOURCE_DIR", "./old_data")
)

func main() {
	// Load .env file if it exists
	_ = godotenv.Load()

	// Parse command line flags
	if len(os.Args) > 1 {
		sourceDir = os.Args[1]
	}

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
	extractor := NewExtractor(sourceDir, log)

	// Verify that source directory exists
	if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
		log.Fatal("Source directory does not exist", zap.String("directory", sourceDir))
	}

	// Run migration
	dryRun := false
	err = RunMigration(ctx, extractor, userRepo, messageRepo, contactRepo, log, dryRun)
	if err != nil {
		log.Fatal("Migration failed", zap.Error(err))
	}

	log.Info("Migration completed successfully")
}

// Extractor extracts data from the old file-based storage system
type Extractor struct {
	sourceDir string
	logger    *zap.Logger
}

// NewExtractor creates a new extractor
func NewExtractor(sourceDir string, logger *zap.Logger) *Extractor {
	return &Extractor{
		sourceDir: sourceDir,
		logger:    logger.With(zap.String("component", "extractor")),
	}
}

// ExtractUsers extracts users from the old file-based storage
func (e *Extractor) ExtractUsers() ([]*domain.User, error) {
	userDir := filepath.Join(e.sourceDir, "extension_wave_keys")
	e.logger.Info("Extracting users", zap.String("directory", userDir))

	files, err := ioutil.ReadDir(userDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read user directory: %w", err)
	}

	var users []*domain.User

	// Process public key files
	for _, file := range files {
		if strings.HasSuffix(file.Name(), "_public.key") {
			username := strings.TrimSuffix(file.Name(), "_public.key")

			// Check if private key file also exists
			privateKeyFile := username + "_private.json"
			privateKeyPath := filepath.Join(userDir, privateKeyFile)
			if _, err := os.Stat(privateKeyPath); os.IsNotExist(err) {
				e.logger.Warn("No private key file found for user", zap.String("username", username))
				continue
			}

			// Read public key
			publicKeyPath := filepath.Join(userDir, file.Name())
			publicKey, err := ioutil.ReadFile(publicKeyPath)
			if err != nil {
				e.logger.Error("Failed to read public key", zap.String("file", publicKeyPath), zap.Error(err))
				continue
			}

			// Read private key
			privateKeyData, err := ioutil.ReadFile(privateKeyPath)
			if err != nil {
				e.logger.Error("Failed to read private key", zap.String("file", privateKeyPath), zap.Error(err))
				continue
			}

			var privateKeyJSON struct {
				Salt         string `json:"salt"`
				EncryptedKey string `json:"encrypted_key"`
			}

			if err := json.Unmarshal(privateKeyData, &privateKeyJSON); err != nil {
				e.logger.Error("Failed to parse private key JSON", zap.String("file", privateKeyPath), zap.Error(err))
				continue
			}

			// Decode base64 fields
			salt, err := base64.URLEncoding.DecodeString(privateKeyJSON.Salt)
			if err != nil {
				e.logger.Error("Failed to decode salt", zap.String("username", username), zap.Error(err))
				continue
			}

			encryptedPrivateKey, err := base64.URLEncoding.DecodeString(privateKeyJSON.EncryptedKey)
			if err != nil {
				e.logger.Error("Failed to decode encrypted private key", zap.String("username", username), zap.Error(err))
				continue
			}

			// Create user object
			userID := security.HashUsername(username)
			now := time.Now()
			user := &domain.User{
				UserID:              userID,
				Username:            username,
				PublicKey:           publicKey,
				EncryptedPrivateKey: encryptedPrivateKey,
				Salt:                salt,
				CreatedAt:           now,
				LastActive:          now,
			}

			users = append(users, user)
			e.logger.Info("Extracted user", zap.String("username", username))
		}
	}

	e.logger.Info("User extraction completed", zap.Int("count", len(users)))
	return users, nil
}

// ExtractContacts extracts contacts from the old file-based storage
func (e *Extractor) ExtractContacts() ([]*domain.Contact, error) {
	contactDir := filepath.Join(e.sourceDir, "extension_wave_contacts")
	e.logger.Info("Extracting contacts", zap.String("directory", contactDir))

	files, err := ioutil.ReadDir(contactDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read contact directory: %w", err)
	}

	var contacts []*domain.Contact

	// Process contact files
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".json") {
			username := strings.TrimSuffix(file.Name(), ".json")

			// Read contact file
			contactPath := filepath.Join(contactDir, file.Name())
			contactData, err := ioutil.ReadFile(contactPath)
			if err != nil {
				e.logger.Error("Failed to read contact file", zap.String("file", contactPath), zap.Error(err))
				continue
			}

			var contactMap map[string]struct {
				Nickname string `json:"nickname"`
			}

			if err := json.Unmarshal(contactData, &contactMap); err != nil {
				e.logger.Error("Failed to parse contact JSON", zap.String("file", contactPath), zap.Error(err))
				continue
			}

			// Create user ID
			userID := security.HashUsername(username)

			// Create contacts
			for pubKey, contactInfo := range contactMap {
				nickname := contactInfo.Nickname
				if nickname == "" {
					nickname = "Contact " + pubKey[:8]
				}

				contact := &domain.Contact{
					UserID:        userID,
					ContactPubKey: pubKey,
					Nickname:      nickname,
					CreatedAt:     time.Now(),
				}

				contacts = append(contacts, contact)
			}

			e.logger.Info("Extracted contacts for user",
				zap.String("username", username),
				zap.Int("count", len(contactMap)))
		}
	}

	e.logger.Info("Contact extraction completed", zap.Int("count", len(contacts)))
	return contacts, nil
}

// ExtractMessages extracts messages from the old file-based storage
func (e *Extractor) ExtractMessages() ([]*domain.Message, error) {
	messageDir := filepath.Join(e.sourceDir, "extension_wave_messages")
	e.logger.Info("Extracting messages", zap.String("directory", messageDir))

	folders, err := ioutil.ReadDir(messageDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read message directory: %w", err)
	}

	var messages []*domain.Message

	// Process each user folder
	for _, folder := range folders {
		if !folder.IsDir() {
			continue
		}

		userFolder := filepath.Join(messageDir, folder.Name())
		messageFiles, err := ioutil.ReadDir(userFolder)
		if err != nil {
			e.logger.Error("Failed to read user message folder", zap.String("folder", userFolder), zap.Error(err))
			continue
		}

		// Process each message file
		for _, messageFile := range messageFiles {
			if messageFile.IsDir() {
				continue
			}

			// Read message file
			messagePath := filepath.Join(userFolder, messageFile.Name())
			messageData, err := ioutil.ReadFile(messagePath)
			if err != nil {
				e.logger.Error("Failed to read message file", zap.String("file", messagePath), zap.Error(err))
				continue
			}

			var msgData struct {
				MessageID           string  `json:"message_id"`
				SenderPubKey        string  `json:"sender_pubkey_b64"`
				RecipientPubKey     string  `json:"recipient_pubkey_b64"`
				CiphertextKEM       string  `json:"ciphertext_kem"`
				CiphertextMsg       string  `json:"ciphertext_msg"`
				Nonce               string  `json:"nonce"`
				SenderCiphertextKEM string  `json:"sender_ciphertext_kem,omitempty"`
				SenderCiphertextMsg string  `json:"sender_ciphertext_msg,omitempty"`
				SenderNonce         string  `json:"sender_nonce,omitempty"`
				Timestamp           float64 `json:"timestamp"`
				Status              string  `json:"status,omitempty"`
			}

			if err := json.Unmarshal(messageData, &msgData); err != nil {
				e.logger.Error("Failed to parse message JSON", zap.String("file", messagePath), zap.Error(err))
				continue
			}

			// Decode base64 fields
			ciphertextKEM, err := base64.URLEncoding.DecodeString(msgData.CiphertextKEM)
			if err != nil {
				e.logger.Error("Failed to decode ciphertext KEM", zap.Error(err))
				continue
			}

			ciphertextMsg, err := base64.URLEncoding.DecodeString(msgData.CiphertextMsg)
			if err != nil {
				e.logger.Error("Failed to decode ciphertext message", zap.Error(err))
				continue
			}

			nonce, err := base64.URLEncoding.DecodeString(msgData.Nonce)
			if err != nil {
				e.logger.Error("Failed to decode nonce", zap.Error(err))
				continue
			}

			// Decode sender fields if present
			var senderCiphertextKEM, senderCiphertextMsg, senderNonce []byte

			if msgData.SenderCiphertextKEM != "" {
				senderCiphertextKEM, err = base64.URLEncoding.DecodeString(msgData.SenderCiphertextKEM)
				if err != nil {
					e.logger.Error("Failed to decode sender ciphertext KEM", zap.Error(err))
					continue
				}
			}

			if msgData.SenderCiphertextMsg != "" {
				senderCiphertextMsg, err = base64.URLEncoding.DecodeString(msgData.SenderCiphertextMsg)
				if err != nil {
					e.logger.Error("Failed to decode sender ciphertext message", zap.Error(err))
					continue
				}
			}

			if msgData.SenderNonce != "" {
				senderNonce, err = base64.URLEncoding.DecodeString(msgData.SenderNonce)
				if err != nil {
					e.logger.Error("Failed to decode sender nonce", zap.Error(err))
					continue
				}
			}

			// Parse message ID or generate a new one
			var messageID uuid.UUID
			if msgData.MessageID != "" {
				messageID, err = uuid.Parse(msgData.MessageID)
				if err != nil {
					e.logger.Warn("Invalid message ID, generating new one", zap.String("invalid_id", msgData.MessageID))
					messageID = uuid.New()
				}
			} else {
				messageID = uuid.New()
			}

			// Use "sent" as default status if not specified
			status := domain.MessageStatusSent
			if msgData.Status != "" {
				status = domain.MessageStatus(msgData.Status)
			}

			// Create message object
			message := &domain.Message{
				MessageID:           messageID,
				SenderPubKey:        msgData.SenderPubKey,
				RecipientPubKey:     msgData.RecipientPubKey,
				CiphertextKEM:       ciphertextKEM,
				CiphertextMsg:       ciphertextMsg,
				Nonce:               nonce,
				SenderCiphertextKEM: senderCiphertextKEM,
				SenderCiphertextMsg: senderCiphertextMsg,
				SenderNonce:         senderNonce,
				Timestamp:           time.Unix(int64(msgData.Timestamp), 0),
				Status:              status,
			}

			messages = append(messages, message)
		}
	}

	e.logger.Info("Message extraction completed", zap.Int("count", len(messages)))
	return messages, nil
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

func getEnvOrDefault(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}
