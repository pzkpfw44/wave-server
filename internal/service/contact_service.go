package service

import (
	"context"

	"go.uber.org/zap"

	"github.com/pzkpfw44/wave-server/internal/domain"
	"github.com/pzkpfw44/wave-server/internal/errors"
	"github.com/pzkpfw44/wave-server/internal/repository"
)

// ContactService provides contact business logic
type ContactService struct {
	contactRepo *repository.ContactRepository
	logger      *zap.Logger
}

// NewContactService creates a new ContactService
func NewContactService(
	contactRepo *repository.ContactRepository,
	logger *zap.Logger,
) *ContactService {
	return &ContactService{
		contactRepo: contactRepo,
		logger:      logger.With(zap.String("service", "contact")),
	}
}

// AddContact adds a new contact for a user
func (s *ContactService) AddContact(ctx context.Context, userID, contactPubKey, nickname string) (*domain.Contact, error) {
	// Validate inputs
	if contactPubKey == "" {
		return nil, errors.NewValidationError("Contact public key is required", nil)
	}

	if nickname == "" {
		return nil, errors.NewValidationError("Nickname is required", nil)
	}

	if len(nickname) > 50 {
		return nil, errors.NewValidationError("Nickname must be at most 50 characters", nil)
	}

	// Create the contact
	contact := domain.NewContact(userID, contactPubKey, nickname)

	// Store the contact
	if err := s.contactRepo.Create(ctx, contact); err != nil {
		return nil, err
	}

	s.logger.Debug("Contact added",
		zap.String("user_id", userID),
		zap.String("contact_pubkey", contactPubKey),
		zap.String("nickname", nickname),
	)

	return contact, nil
}

// GetContacts gets all contacts for a user
func (s *ContactService) GetContacts(ctx context.Context, userID string) ([]*domain.Contact, error) {
	return s.contactRepo.GetByUserID(ctx, userID)
}

// GetContact gets a specific contact
func (s *ContactService) GetContact(ctx context.Context, userID, contactPubKey string) (*domain.Contact, error) {
	if contactPubKey == "" {
		return nil, errors.NewValidationError("Contact public key is required", nil)
	}

	return s.contactRepo.GetByContactPubKey(ctx, userID, contactPubKey)
}

// UpdateContact updates a contact's nickname
func (s *ContactService) UpdateContact(ctx context.Context, userID, contactPubKey, nickname string) (*domain.Contact, error) {
	// Validate inputs
	if contactPubKey == "" {
		return nil, errors.NewValidationError("Contact public key is required", nil)
	}

	if nickname == "" {
		return nil, errors.NewValidationError("Nickname is required", nil)
	}

	if len(nickname) > 50 {
		return nil, errors.NewValidationError("Nickname must be at most 50 characters", nil)
	}

	// Get the current contact
	contact, err := s.contactRepo.GetByContactPubKey(ctx, userID, contactPubKey)
	if err != nil {
		return nil, err
	}

	// Update the nickname
	contact.Nickname = nickname

	// Store the updated contact
	if err := s.contactRepo.Update(ctx, contact); err != nil {
		return nil, err
	}

	s.logger.Debug("Contact updated",
		zap.String("user_id", userID),
		zap.String("contact_pubkey", contactPubKey),
		zap.String("nickname", nickname),
	)

	return contact, nil
}

// DeleteContact deletes a contact
func (s *ContactService) DeleteContact(ctx context.Context, userID, contactPubKey string) error {
	if contactPubKey == "" {
		return errors.NewValidationError("Contact public key is required", nil)
	}

	if err := s.contactRepo.Delete(ctx, userID, contactPubKey); err != nil {
		return err
	}

	s.logger.Debug("Contact deleted",
		zap.String("user_id", userID),
		zap.String("contact_pubkey", contactPubKey),
	)

	return nil
}

// DeleteUserContacts deletes all contacts for a user
func (s *ContactService) DeleteUserContacts(ctx context.Context, userID string) (int64, error) {
	count, err := s.contactRepo.DeleteUserContacts(ctx, userID)
	if err != nil {
		return 0, err
	}

	s.logger.Info("Deleted user contacts",
		zap.String("user_id", userID),
		zap.Int64("count", count),
	)

	return count, nil
}
