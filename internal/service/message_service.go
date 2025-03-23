package service

import (
	"context"
	"encoding/base64"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/yourusername/wave-server/internal/domain"
	"github.com/yourusername/wave-server/internal/errors"
	"github.com/yourusername/wave-server/internal/repository"
)

// MessageService provides message business logic
type MessageService struct {
	messageRepo *repository.MessageRepository
	userRepo    *repository.UserRepository
	logger      *zap.Logger
}

// NewMessageService creates a new MessageService
func NewMessageService(
	messageRepo *repository.MessageRepository,
	userRepo *repository.UserRepository,
	logger *zap.Logger,
) *MessageService {
	return &MessageService{
		messageRepo: messageRepo,
		userRepo:    userRepo,
		logger:      logger.With(zap.String("service", "message")),
	}
}

// SendMessage sends a new message
// Note: In zero-knowledge architecture, message is encrypted client-side
func (s *MessageService) SendMessage(ctx context.Context, userID, recipientPubKey string,
	ciphertextKEMB64, ciphertextMsgB64, nonceB64 string,
	senderCiphertextKEMB64, senderCiphertextMsgB64, senderNonceB64 string) (*domain.Message, error) {

	// Validate inputs
	if recipientPubKey == "" {
		return nil, errors.NewValidationError("Recipient public key is required", nil)
	}

	if ciphertextKEMB64 == "" || ciphertextMsgB64 == "" || nonceB64 == "" {
		return nil, errors.NewValidationError("Ciphertext KEM, ciphertext message, and nonce are required", nil)
	}

	if senderCiphertextKEMB64 == "" || senderCiphertextMsgB64 == "" || senderNonceB64 == "" {
		return nil, errors.NewValidationError("Sender ciphertext KEM, sender ciphertext message, and sender nonce are required", nil)
	}

	// Decode base64 values
	ciphertextKEM, err := base64.URLEncoding.DecodeString(ciphertextKEMB64)
	if err != nil {
		return nil, errors.NewValidationError("Invalid ciphertext KEM format", err)
	}

	ciphertextMsg, err := base64.URLEncoding.DecodeString(ciphertextMsgB64)
	if err != nil {
		return nil, errors.NewValidationError("Invalid ciphertext message format", err)
	}

	nonce, err := base64.URLEncoding.DecodeString(nonceB64)
	if err != nil {
		return nil, errors.NewValidationError("Invalid nonce format", err)
	}

	senderCiphertextKEM, err := base64.URLEncoding.DecodeString(senderCiphertextKEMB64)
	if err != nil {
		return nil, errors.NewValidationError("Invalid sender ciphertext KEM format", err)
	}

	senderCiphertextMsg, err := base64.URLEncoding.DecodeString(senderCiphertextMsgB64)
	if err != nil {
		return nil, errors.NewValidationError("Invalid sender ciphertext message format", err)
	}

	senderNonce, err := base64.URLEncoding.DecodeString(senderNonceB64)
	if err != nil {
		return nil, errors.NewValidationError("Invalid sender nonce format", err)
	}

	// Get the sender's public key
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.NewInternalError("Failed to get sender information", err)
	}

	// Create the message
	message := domain.NewMessage(
		base64.URLEncoding.EncodeToString(user.PublicKey),
		recipientPubKey,
		ciphertextKEM,
		ciphertextMsg,
		nonce,
		senderCiphertextKEM,
		senderCiphertextMsg,
		senderNonce,
	)

	// Store the message
	if err := s.messageRepo.Create(ctx, message); err != nil {
		return nil, errors.NewInternalError("Failed to store message", err)
	}

	s.logger.Debug("Message sent",
		zap.String("message_id", message.MessageID.String()),
		zap.String("sender", userID),
		zap.String("recipient", recipientPubKey),
	)

	return message, nil
}

// GetMessageByID gets a message by its ID
func (s *MessageService) GetMessageByID(ctx context.Context, messageID uuid.UUID) (*domain.Message, error) {
	return s.messageRepo.GetByID(ctx, messageID)
}

// GetMessagesForUser gets all messages for a user (as recipient) with pagination
func (s *MessageService) GetMessagesForUser(ctx context.Context, userPubKey string, limit, offset int) ([]*domain.Message, error) {
	if limit <= 0 {
		limit = 100 // Default limit
	}
	if limit > 1000 {
		limit = 1000 // Max limit
	}

	return s.messageRepo.GetByRecipient(ctx, userPubKey, limit, offset)
}

// GetMessagesSentByUser gets all messages sent by a user with pagination
func (s *MessageService) GetMessagesSentByUser(ctx context.Context, userPubKey string, limit, offset int) ([]*domain.Message, error) {
	if limit <= 0 {
		limit = 100 // Default limit
	}
	if limit > 1000 {
		limit = 1000 // Max limit
	}

	return s.messageRepo.GetBySender(ctx, userPubKey, limit, offset)
}

// GetConversation gets messages between two users with pagination
func (s *MessageService) GetConversation(ctx context.Context, userPubKey, contactPubKey string, limit, offset int) ([]*domain.Message, error) {
	if limit <= 0 {
		limit = 100 // Default limit
	}
	if limit > 1000 {
		limit = 1000 // Max limit
	}

	return s.messageRepo.GetConversation(ctx, userPubKey, contactPubKey, limit, offset)
}

// UpdateMessageStatus updates a message's status
func (s *MessageService) UpdateMessageStatus(ctx context.Context, messageID uuid.UUID, status domain.MessageStatus) error {
	// Validate status
	if status != domain.MessageStatusSent &&
		status != domain.MessageStatusDelivered &&
		status != domain.MessageStatusRead {
		return errors.NewValidationError("Invalid message status", nil)
	}

	return s.messageRepo.UpdateStatus(ctx, messageID, status)
}

// DeleteUserMessages deletes all messages where a user is sender or recipient
func (s *MessageService) DeleteUserMessages(ctx context.Context, userPubKey string) (int64, error) {
	count, err := s.messageRepo.DeleteUserMessages(ctx, userPubKey)
	if err != nil {
		return 0, err
	}

	s.logger.Info("Deleted user messages",
		zap.String("user_pubkey", userPubKey),
		zap.Int64("count", count),
	)

	return count, nil
}
