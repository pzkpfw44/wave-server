package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/pzkpfw44/wave-server/internal/domain"
	"github.com/pzkpfw44/wave-server/internal/service"
)

// MockUserService is a mock implementation of the UserService
type MockUserService struct {
	mock.Mock
}

// Register mocks the Register method
func (m *MockUserService) Register(ctx context.Context, username, publicKeyB64, encPrivateKeyB64, saltB64 string) (*domain.User, error) {
	args := m.Called(ctx, username, publicKeyB64, encPrivateKeyB64, saltB64)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

// GetByID mocks the GetByID method
func (m *MockUserService) GetByID(ctx context.Context, userID string) (*domain.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

// GetByUsername mocks the GetByUsername method
func (m *MockUserService) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

// GetPublicKey mocks the GetPublicKey method
func (m *MockUserService) GetPublicKey(ctx context.Context, username string) ([]byte, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

// GetEncryptedPrivateKey mocks the GetEncryptedPrivateKey method
func (m *MockUserService) GetEncryptedPrivateKey(ctx context.Context, userID string) (*domain.PrivateKeyResponse, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PrivateKeyResponse), args.Error(1)
}

// UpdateLastActive mocks the UpdateLastActive method
func (m *MockUserService) UpdateLastActive(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

// DeleteUser mocks the DeleteUser method
func (m *MockUserService) DeleteUser(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

// MockAuthService is a mock implementation of the AuthService
type MockAuthService struct {
	mock.Mock
}

// Login mocks the Login method
func (m *MockAuthService) Login(ctx context.Context, username string) (string, error) {
	args := m.Called(ctx, username)
	return args.String(0), args.Error(1)
}

// ValidateToken mocks the ValidateToken method
func (m *MockAuthService) ValidateToken(ctx context.Context, tokenStr string) (string, error) {
	args := m.Called(ctx, tokenStr)
	return args.String(0), args.Error(1)
}

// RefreshToken mocks the RefreshToken method
func (m *MockAuthService) RefreshToken(ctx context.Context, tokenStr string) (string, error) {
	args := m.Called(ctx, tokenStr)
	return args.String(0), args.Error(1)
}

// Logout mocks the Logout method
func (m *MockAuthService) Logout(ctx context.Context, tokenStr string) error {
	args := m.Called(ctx, tokenStr)
	return args.Error(0)
}

// LogoutAll mocks the LogoutAll method
func (m *MockAuthService) LogoutAll(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

// CleanupExpiredTokens mocks the CleanupExpiredTokens method
func (m *MockAuthService) CleanupExpiredTokens(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// ScheduleTokenCleanup mocks the ScheduleTokenCleanup method
func (m *MockAuthService) ScheduleTokenCleanup(ctx context.Context) {
	m.Called(ctx)
}

// UpdateUserActivity mocks the UpdateUserActivity method
func (m *MockAuthService) UpdateUserActivity(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

// MockMessageService is a mock implementation of the MessageService
type MockMessageService struct {
	mock.Mock
}

// SendMessage mocks the SendMessage method
func (m *MockMessageService) SendMessage(ctx context.Context, userID, recipientPubKey string,
	ciphertextKEMB64, ciphertextMsgB64, nonceB64 string,
	senderCiphertextKEMB64, senderCiphertextMsgB64, senderNonceB64 string) (*domain.Message, error) {
	args := m.Called(ctx, userID, recipientPubKey, ciphertextKEMB64, ciphertextMsgB64, nonceB64,
		senderCiphertextKEMB64, senderCiphertextMsgB64, senderNonceB64)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Message), args.Error(1)
}

// GetMessageByID mocks the GetMessageByID method
func (m *MockMessageService) GetMessageByID(ctx context.Context, messageID uuid.UUID) (*domain.Message, error) {
	args := m.Called(ctx, messageID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Message), args.Error(1)
}

// GetMessagesForUser mocks the GetMessagesForUser method
func (m *MockMessageService) GetMessagesForUser(ctx context.Context, userPubKey string, limit, offset int) ([]*domain.Message, error) {
	args := m.Called(ctx, userPubKey, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Message), args.Error(1)
}

// GetMessagesSentByUser mocks the GetMessagesSentByUser method
func (m *MockMessageService) GetMessagesSentByUser(ctx context.Context, userPubKey string, limit, offset int) ([]*domain.Message, error) {
	args := m.Called(ctx, userPubKey, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Message), args.Error(1)
}

// GetConversation mocks the GetConversation method
func (m *MockMessageService) GetConversation(ctx context.Context, userPubKey, contactPubKey string, limit, offset int) ([]*domain.Message, error) {
	args := m.Called(ctx, userPubKey, contactPubKey, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Message), args.Error(1)
}

// UpdateMessageStatus mocks the UpdateMessageStatus method
func (m *MockMessageService) UpdateMessageStatus(ctx context.Context, messageID uuid.UUID, status domain.MessageStatus) error {
	args := m.Called(ctx, messageID, status)
	return args.Error(0)
}

// DeleteUserMessages mocks the DeleteUserMessages method
func (m *MockMessageService) DeleteUserMessages(ctx context.Context, userPubKey string) (int64, error) {
	args := m.Called(ctx, userPubKey)
	return args.Get(0).(int64), args.Error(1)
}

// MockContactService is a mock implementation of the ContactService
type MockContactService struct {
	mock.Mock
}

// AddContact mocks the AddContact method
func (m *MockContactService) AddContact(ctx context.Context, userID, contactPubKey, nickname string) (*domain.Contact, error) {
	args := m.Called(ctx, userID, contactPubKey, nickname)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Contact), args.Error(1)
}

// GetContacts mocks the GetContacts method
func (m *MockContactService) GetContacts(ctx context.Context, userID string) ([]*domain.Contact, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Contact), args.Error(1)
}

// GetContact mocks the GetContact method
func (m *MockContactService) GetContact(ctx context.Context, userID, contactPubKey string) (*domain.Contact, error) {
	args := m.Called(ctx, userID, contactPubKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Contact), args.Error(1)
}

// UpdateContact mocks the UpdateContact method
func (m *MockContactService) UpdateContact(ctx context.Context, userID, contactPubKey, nickname string) (*domain.Contact, error) {
	args := m.Called(ctx, userID, contactPubKey, nickname)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Contact), args.Error(1)
}

// DeleteContact mocks the DeleteContact method
func (m *MockContactService) DeleteContact(ctx context.Context, userID, contactPubKey string) error {
	args := m.Called(ctx, userID, contactPubKey)
	return args.Error(0)
}

// DeleteUserContacts mocks the DeleteUserContacts method
func (m *MockContactService) DeleteUserContacts(ctx context.Context, userID string) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

// MockAccountService is a mock implementation of the AccountService
type MockAccountService struct {
	mock.Mock
}

// BackupAccount mocks the BackupAccount method
func (m *MockAccountService) BackupAccount(ctx context.Context, userID string) (*service.BackupData, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.BackupData), args.Error(1)
}

// RecoverAccount mocks the RecoverAccount method
func (m *MockAccountService) RecoverAccount(ctx context.Context, username, publicKeyB64 string,
	encryptedPrivateKey map[string]string, contactsData map[string]interface{},
	messagesData []interface{}) (*domain.User, error) {
	args := m.Called(ctx, username, publicKeyB64, encryptedPrivateKey, contactsData, messagesData)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

// DeleteAccount mocks the DeleteAccount method
func (m *MockAccountService) DeleteAccount(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}
