package service

import (
	"context"
	"encoding/base64"
	"time"

	"go.uber.org/zap"

	"github.com/yourusername/wave-server/internal/domain"
	"github.com/yourusername/wave-server/internal/errors"
	"github.com/yourusername/wave-server/internal/repository"
	"github.com/yourusername/wave-server/internal/security"
)

// UserService provides user business logic
type UserService struct {
	userRepo *repository.UserRepository
	logger   *zap.Logger
}

// NewUserService creates a new UserService
func NewUserService(userRepo *repository.UserRepository, logger *zap.Logger) *UserService {
	return &UserService{
		userRepo: userRepo,
		logger:   logger.With(zap.String("service", "user")),
	}
}

// Register registers a new user
// Note: Keys are already generated and encrypted client-side in zero-knowledge architecture
func (s *UserService) Register(ctx context.Context, username string, publicKeyB64, encPrivateKeyB64, saltB64 string) (*domain.User, error) {
	// Validate input
	if username == "" {
		return nil, errors.NewValidationError("Username is required", nil)
	}
	if len(username) < 3 || len(username) > 50 {
		return nil, errors.NewValidationError("Username must be between 3 and 50 characters", nil)
	}
	if publicKeyB64 == "" || encPrivateKeyB64 == "" || saltB64 == "" {
		return nil, errors.NewValidationError("Public key, encrypted private key, and salt are required", nil)
	}

	// Decode base64 inputs
	publicKey, err := base64.URLEncoding.DecodeString(publicKeyB64)
	if err != nil {
		return nil, errors.NewValidationError("Invalid public key format", err)
	}

	encryptedPrivateKey, err := base64.URLEncoding.DecodeString(encPrivateKeyB64)
	if err != nil {
		return nil, errors.NewValidationError("Invalid encrypted private key format", err)
	}

	salt, err := base64.URLEncoding.DecodeString(saltB64)
	if err != nil {
		return nil, errors.NewValidationError("Invalid salt format", err)
	}

	// Validate key formats
	if err := security.ValidatePublicKeyFormat(publicKey); err != nil {
		return nil, errors.NewValidationError("Invalid public key", err)
	}
	if err := security.ValidateEncryptedPrivateKeyFormat(encryptedPrivateKey); err != nil {
		return nil, errors.NewValidationError("Invalid encrypted private key", err)
	}
	if err := security.ValidateSaltFormat(salt); err != nil {
		return nil, errors.NewValidationError("Invalid salt", err)
	}

	// Check if user already exists
	userID := security.HashUsername(username)
	_, err = s.userRepo.GetByID(ctx, userID)
	if err == nil {
		return nil, errors.NewConflictError("User already exists")
	}

	// Only proceed if we got a "not found" error
	if appErr, ok := errors.IsAppError(err); !ok || appErr.Code != errors.ErrCodeNotFound {
		return nil, errors.NewInternalError("Error checking user existence", err)
	}

	// Create new user
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

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	s.logger.Info("User registered", zap.String("username", username), zap.String("user_id", userID))
	return user, nil
}

// GetByID gets a user by ID
func (s *UserService) GetByID(ctx context.Context, userID string) (*domain.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

// GetByUsername gets a user by username
func (s *UserService) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	return s.userRepo.GetByUsername(ctx, username)
}

// GetPublicKey gets a user's public key
func (s *UserService) GetPublicKey(ctx context.Context, username string) ([]byte, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	return user.PublicKey, nil
}

// GetEncryptedPrivateKey gets a user's encrypted private key and salt
func (s *UserService) GetEncryptedPrivateKey(ctx context.Context, userID string) (*domain.PrivateKeyResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	response := user.GetPrivateKeyResponse()
	return &response, nil
}

// UpdateLastActive updates a user's last active timestamp
func (s *UserService) UpdateLastActive(ctx context.Context, userID string) error {
	return s.userRepo.UpdateLastActive(ctx, userID)
}

// DeleteUser deletes a user
func (s *UserService) DeleteUser(ctx context.Context, userID string) error {
	return s.userRepo.Delete(ctx, userID)
}
