package mocks

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/pzkpfw44/wave-server/internal/domain"
)

// MockUserRepository is a mock implementation of the UserRepository
type MockUserRepository struct {
	mock.Mock
}

// Create mocks the Create method
func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

// GetByUsername mocks the GetByUsername method
func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

// GetByID mocks the GetByID method
func (m *MockUserRepository) GetByID(ctx context.Context, userID string) (*domain.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

// UpdateLastActive mocks the UpdateLastActive method
func (m *MockUserRepository) UpdateLastActive(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

// Delete mocks the Delete method
func (m *MockUserRepository) Delete(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

// MockMessageRepository is a mock implementation of the MessageRepository
type MockMessageRepository struct {
	mock.Mock
}

// Create mocks the Create method
func (m *MockMessageRepository) Create(ctx context.Context, message *domain.Message) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

// GetByID mocks the GetByID method
func (m *MockMessageRepository) GetByID(ctx context.Context, messageID uuid.UUID) (*domain.Message, error) {
	args := m.Called(ctx, messageID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Message), args.Error(1)
}

// GetByRecipient mocks the GetByRecipient method
func (m *MockMessageRepository) GetByRecipient(ctx context.Context, pubKey string, limit, offset int) ([]*domain.Message, error) {
	args := m.Called(ctx, pubKey, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Message), args.Error(1)
}

// GetBySender mocks the GetBySender method
func (m *MockMessageRepository) GetBySender(ctx context.Context, pubKey string, limit, offset int) ([]*domain.Message, error) {
	args := m.Called(ctx, pubKey, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Message), args.Error(1)
}

// GetConversation mocks the GetConversation method
func (m *MockMessageRepository) GetConversation(ctx context.Context, userPubKey, contactPubKey string, limit, offset int) ([]*domain.Message, error) {
	args := m.Called(ctx, userPubKey, contactPubKey, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Message), args.Error(1)
}

// UpdateStatus mocks the UpdateStatus method
func (m *MockMessageRepository) UpdateStatus(ctx context.Context, messageID uuid.UUID, status domain.MessageStatus) error {
	args := m.Called(ctx, messageID, status)
	return args.Error(0)
}

// DeleteUserMessages mocks the DeleteUserMessages method
func (m *MockMessageRepository) DeleteUserMessages(ctx context.Context, pubKey string) (int64, error) {
	args := m.Called(ctx, pubKey)
	return args.Get(0).(int64), args.Error(1)
}

// MockContactRepository is a mock implementation of the ContactRepository
type MockContactRepository struct {
	mock.Mock
}

// Create mocks the Create method
func (m *MockContactRepository) Create(ctx context.Context, contact *domain.Contact) error {
	args := m.Called(ctx, contact)
	return args.Error(0)
}

// GetByUserID mocks the GetByUserID method
func (m *MockContactRepository) GetByUserID(ctx context.Context, userID string) ([]*domain.Contact, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Contact), args.Error(1)
}

// GetByContactPubKey mocks the GetByContactPubKey method
func (m *MockContactRepository) GetByContactPubKey(ctx context.Context, userID, contactPubKey string) (*domain.Contact, error) {
	args := m.Called(ctx, userID, contactPubKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Contact), args.Error(1)
}

// Update mocks the Update method
func (m *MockContactRepository) Update(ctx context.Context, contact *domain.Contact) error {
	args := m.Called(ctx, contact)
	return args.Error(0)
}

// Delete mocks the Delete method
func (m *MockContactRepository) Delete(ctx context.Context, userID, contactPubKey string) error {
	args := m.Called(ctx, userID, contactPubKey)
	return args.Error(0)
}

// DeleteUserContacts mocks the DeleteUserContacts method
func (m *MockContactRepository) DeleteUserContacts(ctx context.Context, userID string) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

// MockTokenRepository is a mock implementation of the TokenRepository
type MockTokenRepository struct {
	mock.Mock
}

// Create mocks the Create method
func (m *MockTokenRepository) Create(ctx context.Context, token *domain.Token) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

// GetByTokenHash mocks the GetByTokenHash method
func (m *MockTokenRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*domain.Token, error) {
	args := m.Called(ctx, tokenHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Token), args.Error(1)
}

// GetByUserID mocks the GetByUserID method
func (m *MockTokenRepository) GetByUserID(ctx context.Context, userID string) ([]*domain.Token, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Token), args.Error(1)
}

// UpdateLastUsed mocks the UpdateLastUsed method
func (m *MockTokenRepository) UpdateLastUsed(ctx context.Context, tokenID uuid.UUID) error {
	args := m.Called(ctx, tokenID)
	return args.Error(0)
}

// Delete mocks the Delete method
func (m *MockTokenRepository) Delete(ctx context.Context, tokenHash string) error {
	args := m.Called(ctx, tokenHash)
	return args.Error(0)
}

// DeleteUserTokens mocks the DeleteUserTokens method
func (m *MockTokenRepository) DeleteUserTokens(ctx context.Context, userID string) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

// CleanupExpired mocks the CleanupExpired method
func (m *MockTokenRepository) CleanupExpired(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

// CreateForUser mocks the CreateForUser method
func (m *MockTokenRepository) CreateForUser(ctx context.Context, userID string, expiryDuration time.Duration) (string, error) {
	args := m.Called(ctx, userID, expiryDuration)
	return args.String(0), args.Error(1)
}

// ValidateToken mocks the ValidateToken method
func (m *MockTokenRepository) ValidateToken(ctx context.Context, tokenStr string) (string, error) {
	args := m.Called(ctx, tokenStr)
	return args.String(0), args.Error(1)
}
