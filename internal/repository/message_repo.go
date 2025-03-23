package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"

	"github.com/pzkpfw44/wave-server/internal/domain"
	"github.com/pzkpfw44/wave-server/internal/errors"
)

// MessageRepository handles message data storage operations
type MessageRepository struct {
	db     *Database
	logger *zap.Logger
}

// NewMessageRepository creates a new MessageRepository
func NewMessageRepository(db *Database) *MessageRepository {
	return &MessageRepository{
		db:     db,
		logger: db.Logger.With(zap.String("repository", "message")),
	}
}

// Create creates a new message
func (r *MessageRepository) Create(ctx context.Context, message *domain.Message) error {
	query := `
	INSERT INTO messages (
		message_id, sender_pubkey, recipient_pubkey,
		ciphertext_kem, ciphertext_msg, nonce,
		sender_ciphertext_kem, sender_ciphertext_msg, sender_nonce,
		timestamp, status
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := r.db.Pool.Exec(ctx, query,
		message.MessageID,
		message.SenderPubKey,
		message.RecipientPubKey,
		message.CiphertextKEM,
		message.CiphertextMsg,
		message.Nonce,
		message.SenderCiphertextKEM,
		message.SenderCiphertextMsg,
		message.SenderNonce,
		message.Timestamp,
		message.Status,
	)

	if err != nil {
		r.logger.Error("Failed to create message", zap.Error(err), zap.String("message_id", message.MessageID.String()))
		return errors.NewInternalError("Failed to create message", err)
	}

	return nil
}

// GetByID gets a message by ID
func (r *MessageRepository) GetByID(ctx context.Context, messageID uuid.UUID) (*domain.Message, error) {
	query := `
	SELECT
		message_id, sender_pubkey, recipient_pubkey,
		ciphertext_kem, ciphertext_msg, nonce,
		sender_ciphertext_kem, sender_ciphertext_msg, sender_nonce,
		timestamp, status
	FROM messages
	WHERE message_id = $1
	`

	row := r.db.Pool.QueryRow(ctx, query, messageID)

	message := &domain.Message{}
	err := row.Scan(
		&message.MessageID,
		&message.SenderPubKey,
		&message.RecipientPubKey,
		&message.CiphertextKEM,
		&message.CiphertextMsg,
		&message.Nonce,
		&message.SenderCiphertextKEM,
		&message.SenderCiphertextMsg,
		&message.SenderNonce,
		&message.Timestamp,
		&message.Status,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.NewNotFoundError(fmt.Sprintf("Message with ID '%s'", messageID))
		}
		r.logger.Error("Failed to get message by ID", zap.Error(err), zap.String("message_id", messageID.String()))
		return nil, errors.NewInternalError("Failed to get message", err)
	}

	return message, nil
}

// GetByRecipient gets messages for a recipient with pagination
func (r *MessageRepository) GetByRecipient(ctx context.Context, pubKey string, limit, offset int) ([]*domain.Message, error) {
	query := `
	SELECT
		message_id, sender_pubkey, recipient_pubkey,
		ciphertext_kem, ciphertext_msg, nonce,
		sender_ciphertext_kem, sender_ciphertext_msg, sender_nonce,
		timestamp, status
	FROM messages
	WHERE recipient_pubkey = $1
	ORDER BY timestamp DESC
	LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Pool.Query(ctx, query, pubKey, limit, offset)
	if err != nil {
		r.logger.Error("Failed to get messages by recipient", zap.Error(err), zap.String("recipient_pubkey", pubKey))
		return nil, errors.NewInternalError("Failed to get messages", err)
	}
	defer rows.Close()

	var messages []*domain.Message
	for rows.Next() {
		message := &domain.Message{}
		err := rows.Scan(
			&message.MessageID,
			&message.SenderPubKey,
			&message.RecipientPubKey,
			&message.CiphertextKEM,
			&message.CiphertextMsg,
			&message.Nonce,
			&message.SenderCiphertextKEM,
			&message.SenderCiphertextMsg,
			&message.SenderNonce,
			&message.Timestamp,
			&message.Status,
		)
		if err != nil {
			r.logger.Error("Failed to scan message row", zap.Error(err))
			return nil, errors.NewInternalError("Failed to read message data", err)
		}
		messages = append(messages, message)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("Error iterating message rows", zap.Error(err))
		return nil, errors.NewInternalError("Failed to read message data", err)
	}

	return messages, nil
}

// GetBySender gets messages sent by a sender with pagination
func (r *MessageRepository) GetBySender(ctx context.Context, pubKey string, limit, offset int) ([]*domain.Message, error) {
	query := `
	SELECT
		message_id, sender_pubkey, recipient_pubkey,
		ciphertext_kem, ciphertext_msg, nonce,
		sender_ciphertext_kem, sender_ciphertext_msg, sender_nonce,
		timestamp, status
	FROM messages
	WHERE sender_pubkey = $1
	ORDER BY timestamp DESC
	LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Pool.Query(ctx, query, pubKey, limit, offset)
	if err != nil {
		r.logger.Error("Failed to get messages by sender", zap.Error(err), zap.String("sender_pubkey", pubKey))
		return nil, errors.NewInternalError("Failed to get messages", err)
	}
	defer rows.Close()

	var messages []*domain.Message
	for rows.Next() {
		message := &domain.Message{}
		err := rows.Scan(
			&message.MessageID,
			&message.SenderPubKey,
			&message.RecipientPubKey,
			&message.CiphertextKEM,
			&message.CiphertextMsg,
			&message.Nonce,
			&message.SenderCiphertextKEM,
			&message.SenderCiphertextMsg,
			&message.SenderNonce,
			&message.Timestamp,
			&message.Status,
		)
		if err != nil {
			r.logger.Error("Failed to scan message row", zap.Error(err))
			return nil, errors.NewInternalError("Failed to read message data", err)
		}
		messages = append(messages, message)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("Error iterating message rows", zap.Error(err))
		return nil, errors.NewInternalError("Failed to read message data", err)
	}

	return messages, nil
}

// GetConversation gets messages between two users with pagination
func (r *MessageRepository) GetConversation(ctx context.Context, userPubKey, contactPubKey string, limit, offset int) ([]*domain.Message, error) {
	query := `
	SELECT
		message_id, sender_pubkey, recipient_pubkey,
		ciphertext_kem, ciphertext_msg, nonce,
		sender_ciphertext_kem, sender_ciphertext_msg, sender_nonce,
		timestamp, status
	FROM messages
	WHERE (sender_pubkey = $1 AND recipient_pubkey = $2)
	   OR (sender_pubkey = $2 AND recipient_pubkey = $1)
	ORDER BY timestamp DESC
	LIMIT $3 OFFSET $4
	`

	rows, err := r.db.Pool.Query(ctx, query, userPubKey, contactPubKey, limit, offset)
	if err != nil {
		r.logger.Error("Failed to get conversation messages",
			zap.Error(err),
			zap.String("user_pubkey", userPubKey),
			zap.String("contact_pubkey", contactPubKey))
		return nil, errors.NewInternalError("Failed to get messages", err)
	}
	defer rows.Close()

	var messages []*domain.Message
	for rows.Next() {
		message := &domain.Message{}
		err := rows.Scan(
			&message.MessageID,
			&message.SenderPubKey,
			&message.RecipientPubKey,
			&message.CiphertextKEM,
			&message.CiphertextMsg,
			&message.Nonce,
			&message.SenderCiphertextKEM,
			&message.SenderCiphertextMsg,
			&message.SenderNonce,
			&message.Timestamp,
			&message.Status,
		)
		if err != nil {
			r.logger.Error("Failed to scan message row", zap.Error(err))
			return nil, errors.NewInternalError("Failed to read message data", err)
		}
		messages = append(messages, message)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("Error iterating message rows", zap.Error(err))
		return nil, errors.NewInternalError("Failed to read message data", err)
	}

	return messages, nil
}

// UpdateStatus updates a message's status
func (r *MessageRepository) UpdateStatus(ctx context.Context, messageID uuid.UUID, status domain.MessageStatus) error {
	query := `
	UPDATE messages
	SET status = $1
	WHERE message_id = $2
	`

	result, err := r.db.Pool.Exec(ctx, query, status, messageID)
	if err != nil {
		r.logger.Error("Failed to update message status",
			zap.Error(err),
			zap.String("message_id", messageID.String()),
			zap.String("status", string(status)))
		return errors.NewInternalError("Failed to update message status", err)
	}

	if result.RowsAffected() == 0 {
		return errors.NewNotFoundError(fmt.Sprintf("Message with ID '%s'", messageID))
	}

	return nil
}

// DeleteUserMessages deletes all messages where a user is sender or recipient
func (r *MessageRepository) DeleteUserMessages(ctx context.Context, pubKey string) (int64, error) {
	query := `
	DELETE FROM messages
	WHERE sender_pubkey = $1 OR recipient_pubkey = $1
	`

	result, err := r.db.Pool.Exec(ctx, query, pubKey)
	if err != nil {
		r.logger.Error("Failed to delete user messages", zap.Error(err), zap.String("pubkey", pubKey))
		return 0, errors.NewInternalError("Failed to delete messages", err)
	}

	return result.RowsAffected(), nil
}
