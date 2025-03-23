package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"

	"github.com/pzkpfw44/wave-server/internal/domain"
	"github.com/pzkpfw44/wave-server/internal/errors"
)

// ContactRepository handles contact data storage operations
type ContactRepository struct {
	db     *Database
	logger *zap.Logger
}

// NewContactRepository creates a new ContactRepository
func NewContactRepository(db *Database) *ContactRepository {
	return &ContactRepository{
		db:     db,
		logger: db.Logger.With(zap.String("repository", "contact")),
	}
}

// Create creates a new contact
func (r *ContactRepository) Create(ctx context.Context, contact *domain.Contact) error {
	query := `
	INSERT INTO contacts (user_id, contact_pubkey, nickname, created_at)
	VALUES ($1, $2, $3, $4)
	`

	_, err := r.db.Pool.Exec(ctx, query,
		contact.UserID,
		contact.ContactPubKey,
		contact.Nickname,
		contact.CreatedAt,
	)

	if err != nil {
		r.logger.Error("Failed to create contact",
			zap.Error(err),
			zap.String("user_id", contact.UserID),
			zap.String("contact_pubkey", contact.ContactPubKey))

		// Check for unique constraint violation
		if pgErr, ok := err.(*pgx.PgError); ok && pgErr.Code == "23505" {
			return errors.NewConflictError(fmt.Sprintf("Contact with public key '%s' already exists for this user", contact.ContactPubKey))
		}

		return errors.NewInternalError("Failed to create contact", err)
	}

	return nil
}

// GetByUserID gets all contacts for a user
func (r *ContactRepository) GetByUserID(ctx context.Context, userID string) ([]*domain.Contact, error) {
	query := `
	SELECT user_id, contact_pubkey, nickname, created_at
	FROM contacts
	WHERE user_id = $1
	ORDER BY nickname ASC
	`

	rows, err := r.db.Pool.Query(ctx, query, userID)
	if err != nil {
		r.logger.Error("Failed to get contacts by user ID", zap.Error(err), zap.String("user_id", userID))
		return nil, errors.NewInternalError("Failed to get contacts", err)
	}
	defer rows.Close()

	var contacts []*domain.Contact
	for rows.Next() {
		contact := &domain.Contact{}
		err := rows.Scan(
			&contact.UserID,
			&contact.ContactPubKey,
			&contact.Nickname,
			&contact.CreatedAt,
		)
		if err != nil {
			r.logger.Error("Failed to scan contact row", zap.Error(err))
			return nil, errors.NewInternalError("Failed to read contact data", err)
		}
		contacts = append(contacts, contact)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("Error iterating contact rows", zap.Error(err))
		return nil, errors.NewInternalError("Failed to read contact data", err)
	}

	return contacts, nil
}

// GetByContactPubKey gets a specific contact
func (r *ContactRepository) GetByContactPubKey(ctx context.Context, userID, contactPubKey string) (*domain.Contact, error) {
	query := `
	SELECT user_id, contact_pubkey, nickname, created_at
	FROM contacts
	WHERE user_id = $1 AND contact_pubkey = $2
	`

	row := r.db.Pool.QueryRow(ctx, query, userID, contactPubKey)

	contact := &domain.Contact{}
	err := row.Scan(
		&contact.UserID,
		&contact.ContactPubKey,
		&contact.Nickname,
		&contact.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.NewNotFoundError(fmt.Sprintf("Contact with public key '%s'", contactPubKey))
		}
		r.logger.Error("Failed to get contact by public key",
			zap.Error(err),
			zap.String("user_id", userID),
			zap.String("contact_pubkey", contactPubKey))
		return nil, errors.NewInternalError("Failed to get contact", err)
	}

	return contact, nil
}

// Update updates a contact
func (r *ContactRepository) Update(ctx context.Context, contact *domain.Contact) error {
	query := `
	UPDATE contacts
	SET nickname = $3
	WHERE user_id = $1 AND contact_pubkey = $2
	`

	result, err := r.db.Pool.Exec(ctx, query, contact.UserID, contact.ContactPubKey, contact.Nickname)
	if err != nil {
		r.logger.Error("Failed to update contact",
			zap.Error(err),
			zap.String("user_id", contact.UserID),
			zap.String("contact_pubkey", contact.ContactPubKey))
		return errors.NewInternalError("Failed to update contact", err)
	}

	if result.RowsAffected() == 0 {
		return errors.NewNotFoundError(fmt.Sprintf("Contact with public key '%s'", contact.ContactPubKey))
	}

	return nil
}

// Delete deletes a contact
func (r *ContactRepository) Delete(ctx context.Context, userID, contactPubKey string) error {
	query := `
	DELETE FROM contacts
	WHERE user_id = $1 AND contact_pubkey = $2
	`

	result, err := r.db.Pool.Exec(ctx, query, userID, contactPubKey)
	if err != nil {
		r.logger.Error("Failed to delete contact",
			zap.Error(err),
			zap.String("user_id", userID),
			zap.String("contact_pubkey", contactPubKey))
		return errors.NewInternalError("Failed to delete contact", err)
	}

	if result.RowsAffected() == 0 {
		return errors.NewNotFoundError(fmt.Sprintf("Contact with public key '%s'", contactPubKey))
	}

	return nil
}

// DeleteUserContacts deletes all contacts for a user
func (r *ContactRepository) DeleteUserContacts(ctx context.Context, userID string) (int64, error) {
	query := `
	DELETE FROM contacts
	WHERE user_id = $1
	`

	result, err := r.db.Pool.Exec(ctx, query, userID)
	if err != nil {
		r.logger.Error("Failed to delete user contacts", zap.Error(err), zap.String("user_id", userID))
		return 0, errors.NewInternalError("Failed to delete contacts", err)
	}

	return result.RowsAffected(), nil
}
