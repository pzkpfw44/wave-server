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
	"github.com/pzkpfw44/wave-server/internal/repository"
	"github.com/pzkpfw44/wave-server/internal/security"
	"github.com/pzkpfw44/wave-server/pkg/logger"
)

// FilesystemUser represents a user from the file-based storage
type FilesystemUser struct {
	Username            string `json:"username"`
	PublicKey           string `json:"public_key"`
	EncryptedPrivateKey struct {
		Salt         string `json:"salt"`
		EncryptedKey string `json:"encrypted_key"`
	} `json:"encrypted_private_key"`
}

// FilesystemMessage represents a message from the file-based storage
type FilesystemMessage struct {
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

// FilesystemContact represents a contact from the file-based storage
type FilesystemContact struct {
	Nickname string `json:"nickname"`
}

func main() {
	// Load .env file if it exists
	_ = godotenv.Load()

	// Parse command line flags
	sourceDir := getEnvOrDefault("SOURCE_DIR", "./old_data")

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

	// Verify that source directory exists
	if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
		log.Fatal("Source directory does not exist", zap.String("directory", sourceDir))
	}

	// Migrate users
	userDir := filepath.Join(sourceDir, "extension_wave_keys")
	migrateUsers(ctx, userDir, userRepo, log)

	// Migrate messages
	messageDir := filepath.Join(sourceDir, "extension_wave_messages")
	migrateMessages(ctx, messageDir, messageRepo, log)

	// Migrate contacts
	contactDir := filepath.Join(sourceDir, "extension_wave_contacts")
	migrateContacts(ctx, contactDir, contactRepo, log)

	log.Info("Migration completed")
}

func migrateUsers(ctx context.Context, userDir string, userRepo *repository.UserRepository, log *zap.Logger) {
	log.Info("Migrating users", zap.String("directory", userDir))

	files, err := ioutil.ReadDir(userDir)
	if err != nil {
		log.Fatal("Failed to read user directory", zap.Error(err))
	}

	userCount := 0

	// Process public key files
	for _, file := range files {
		if strings.HasSuffix(file.Name(), "_public.key") {
			username := strings.TrimSuffix(file.Name(), "_public.key")

			// Check if we also have the private key file
			privateKeyFile := username + "_private.json"
			if _, err := os.Stat(filepath.Join(userDir, privateKeyFile)); os.IsNotExist(err) {
				log.Warn("No private key file found for user", zap.String("username", username))
				continue
			}

			// Read public key
			publicKeyPath := filepath.Join(userDir, file.Name())
			publicKey, err := ioutil.ReadFile(publicKeyPath)
			if err != nil {
				log.Error("Failed to read public key", zap.String("file", publicKeyPath), zap.Error(err))
				continue
			}

			// Read private key
			privateKeyPath := filepath.Join(userDir, privateKeyFile)
			privateKeyData, err := ioutil.ReadFile(privateKeyPath)
			if err != nil {
				log.Error("Failed to read private key", zap.String("file", privateKeyPath), zap.Error(err))
				continue
			}

			var privateKeyJSON struct {
				Salt         string `json:"salt"`
				EncryptedKey string `json:"encrypted_key"`
			}

			if err := json.Unmarshal(privateKeyData, &privateKeyJSON); err != nil {
				log.Error("Failed to parse private key JSON", zap.String("file", privateKeyPath), zap.Error(err))
				continue
			}

			// Decode base64 fields
			salt, err := base64.URLEncoding.DecodeString(privateKeyJSON.Salt)
			if err != nil {
				log.Error("Failed to decode salt", zap.String("username", username), zap.Error(err))
				continue
			}

			encryptedPrivateKey, err := base64.URLEncoding.DecodeString(privateKeyJSON.EncryptedKey)
			if err != nil {
				log.Error("Failed to decode encrypted private key", zap.String("username", username), zap.Error(err))
				continue
			}

			// Create user in database
			userID := security.HashUsername(username)
			now := time.Now()
			user := struct {
				UserID              string
				Username            string
				PublicKey           []byte
				EncryptedPrivateKey []byte
				Salt                []byte
				CreatedAt           time.Time
				LastActive          time.Time
			}{
				UserID:              userID,
				Username:            username,
				PublicKey:           publicKey,
				EncryptedPrivateKey: encryptedPrivateKey,
				Salt:                salt,
				CreatedAt:           now,
				LastActive:          now,
			}

			if err := userRepo.Create(ctx, &user); err != nil {
				log.Error("Failed to create user in database", zap.String("username", username), zap.Error(err))
				continue
			}

			userCount++
			log.Info("Migrated user", zap.String("username", username))
		}
	}

	log.Info("User migration completed", zap.Int("migrated", userCount))
}

func migrateMessages(ctx context.Context, messageDir string, messageRepo *repository.MessageRepository, log *zap.Logger) {
	log.Info("Migrating messages", zap.String("directory", messageDir))

	// Read message folders (these are user folders)
	folders, err := ioutil.ReadDir(messageDir)
	if err != nil {
		log.Fatal("Failed to read message directory", zap.Error(err))
	}

	messageCount := 0

	// Process each user folder
	for _, folder := range folders {
		if !folder.IsDir() {
			continue
		}

		userFolder := filepath.Join(messageDir, folder.Name())
		messageFiles, err := ioutil.ReadDir(userFolder)
		if err != nil {
			log.Error("Failed to read user message folder", zap.String("folder", userFolder), zap.Error(err))
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
				log.Error("Failed to read message file", zap.String("file", messagePath), zap.Error(err))
				continue
			}

			var message FilesystemMessage
			if err := json.Unmarshal(messageData, &message); err != nil {
				log.Error("Failed to parse message JSON", zap.String("file", messagePath), zap.Error(err))
				continue
			}

			// Decode base64 fields
			ciphertextKEM, err := base64.URLEncoding.DecodeString(message.CiphertextKEM)
			if err != nil {
				log.Error("Failed to decode ciphertext KEM", zap.Error(err))
				continue
			}

			ciphertextMsg, err := base64.URLEncoding.DecodeString(message.CiphertextMsg)
			if err != nil {
				log.Error("Failed to decode ciphertext message", zap.Error(err))
				continue
			}

			nonce, err := base64.URLEncoding.DecodeString(message.Nonce)
			if err != nil {
				log.Error("Failed to decode nonce", zap.Error(err))
				continue
			}

			// Decode sender fields if present
			var senderCiphertextKEM, senderCiphertextMsg, senderNonce []byte

			if message.SenderCiphertextKEM != "" {
				senderCiphertextKEM, err = base64.URLEncoding.DecodeString(message.SenderCiphertextKEM)
				if err != nil {
					log.Error("Failed to decode sender ciphertext KEM", zap.Error(err))
					continue
				}
			}

			if message.SenderCiphertextMsg != "" {
				senderCiphertextMsg, err = base64.URLEncoding.DecodeString(message.SenderCiphertextMsg)
				if err != nil {
					log.Error("Failed to decode sender ciphertext message", zap.Error(err))
					continue
				}
			}

			if message.SenderNonce != "" {
				senderNonce, err = base64.URLEncoding.DecodeString(message.SenderNonce)
				if err != nil {
					log.Error("Failed to decode sender nonce", zap.Error(err))
					continue
				}
			}

			// Create message in database
			status := "sent"
			if message.Status != "" {
				status = message.Status
			}

			newMessage := struct {
				MessageID           uuid.UUID
				SenderPubKey        string
				RecipientPubKey     string
				CiphertextKEM       []byte
				CiphertextMsg       []byte
				Nonce               []byte
				SenderCiphertextKEM []byte
				SenderCiphertextMsg []byte
				SenderNonce         []byte
				Timestamp           time.Time
				Status              string
			}{
				MessageID:           uuid.MustParse(message.MessageID),
				SenderPubKey:        message.SenderPubKey,
				RecipientPubKey:     message.RecipientPubKey,
				CiphertextKEM:       ciphertextKEM,
				CiphertextMsg:       ciphertextMsg,
				Nonce:               nonce,
				SenderCiphertextKEM: senderCiphertextKEM,
				SenderCiphertextMsg: senderCiphertextMsg,
				SenderNonce:         senderNonce,
				Timestamp:           time.Unix(int64(message.Timestamp), 0),
				Status:              status,
			}

			if err := messageRepo.Create(ctx, &newMessage); err != nil {
				log.Error("Failed to create message in database", zap.String("message_id", message.MessageID), zap.Error(err))
				continue
			}

			messageCount++
		}
	}

	log.Info("Message migration completed", zap.Int("migrated", messageCount))
}

func migrateContacts(ctx context.Context, contactDir string, contactRepo *repository.ContactRepository, log *zap.Logger) {
	log.Info("Migrating contacts", zap.String("directory", contactDir))

	files, err := ioutil.ReadDir(contactDir)
	if err != nil {
		log.Fatal("Failed to read contact directory", zap.Error(err))
	}

	contactCount := 0

	// Process contact files
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".json") {
			username := strings.TrimSuffix(file.Name(), ".json")

			// Read contact file
			contactPath := filepath.Join(contactDir, file.Name())
			contactData, err := ioutil.ReadFile(contactPath)
			if err != nil {
				log.Error("Failed to read contact file", zap.String("file", contactPath), zap.Error(err))
				continue
			}

			var contacts map[string]FilesystemContact
			if err := json.Unmarshal(contactData, &contacts); err != nil {
				log.Error("Failed to parse contact JSON", zap.String("file", contactPath), zap.Error(err))
				continue
			}

			// Create user ID
			userID := security.HashUsername(username)

			// Create contacts in database
			for pubKey, contact := range contacts {
				if contact.Nickname == "" {
					contact.Nickname = "Contact " + pubKey[:8]
				}

				newContact := struct {
					UserID        string
					ContactPubKey string
					Nickname      string
					CreatedAt     time.Time
				}{
					UserID:        userID,
					ContactPubKey: pubKey,
					Nickname:      contact.Nickname,
					CreatedAt:     time.Now(),
				}

				if err := contactRepo.Create(ctx, &newContact); err != nil {
					log.Error("Failed to create contact in database",
						zap.String("username", username),
						zap.String("pubkey", pubKey),
						zap.Error(err))
					continue
				}

				contactCount++
			}

			log.Info("Migrated contacts for user",
				zap.String("username", username),
				zap.Int("count", len(contacts)))
		}
	}

	log.Info("Contact migration completed", zap.Int("migrated", contactCount))
}

func getEnvOrDefault(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}
