package service

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/pzkpfw44/wave-server/internal/config"
	"github.com/pzkpfw44/wave-server/internal/errors"
	"github.com/pzkpfw44/wave-server/internal/repository"
	"github.com/pzkpfw44/wave-server/internal/security"
)

// AuthService provides authentication business logic
type AuthService struct {
	userRepo  *repository.UserRepository
	tokenRepo *repository.TokenRepository
	config    *config.Config
	logger    *zap.Logger
}

// NewAuthService creates a new AuthService
func NewAuthService(
	userRepo *repository.UserRepository,
	tokenRepo *repository.TokenRepository,
	config *config.Config,
	logger *zap.Logger,
) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		config:    config,
		logger:    logger.With(zap.String("service", "auth")),
	}
}

// Login authenticates a user and returns a token
// Note: In our zero-knowledge architecture, we don't verify the password server-side
// Password verification happens client-side by attempting to decrypt the private key
func (s *AuthService) Login(ctx context.Context, username string) (string, error) {
	// Find the user
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return "", errors.NewUnauthenticatedError("Invalid username")
	}

	// Update last active timestamp
	if err := s.userRepo.UpdateLastActive(ctx, user.UserID); err != nil {
		s.logger.Warn("Failed to update last active timestamp",
			zap.Error(err), zap.String("user_id", user.UserID))
		// Non-critical error, continue
	}

	// Create a token
	token, err := s.tokenRepo.CreateForUser(ctx, user.UserID, s.config.Auth.TokenExpiry)
	if err != nil {
		return "", errors.NewInternalError("Failed to create token", err)
	}

	s.logger.Info("User logged in", zap.String("username", username), zap.String("user_id", user.UserID))
	return token, nil
}

// ValidateToken validates a token and returns the user ID
func (s *AuthService) ValidateToken(ctx context.Context, tokenStr string) (string, error) {
	return s.tokenRepo.ValidateToken(ctx, tokenStr)
}

// RefreshToken validates a token and issues a new one
func (s *AuthService) RefreshToken(ctx context.Context, tokenStr string) (string, error) {
	// Validate the current token
	userID, err := s.tokenRepo.ValidateToken(ctx, tokenStr)
	if err != nil {
		return "", err
	}

	// Delete the old token
	oldTokenHash := security.HashToken(tokenStr)
	if err := s.tokenRepo.Delete(ctx, oldTokenHash); err != nil {
		s.logger.Warn("Failed to delete old token", zap.Error(err))
		// Non-critical error, continue
	}

	// Create a new token
	newToken, err := s.tokenRepo.CreateForUser(ctx, userID, s.config.Auth.TokenExpiry)
	if err != nil {
		return "", errors.NewInternalError("Failed to create token", err)
	}

	s.logger.Info("Token refreshed", zap.String("user_id", userID))
	return newToken, nil
}

// Logout invalidates a token
func (s *AuthService) Logout(ctx context.Context, tokenStr string) error {
	tokenHash := security.HashToken(tokenStr)
	if err := s.tokenRepo.Delete(ctx, tokenHash); err != nil {
		// Don't return an error if the token was not found
		if appErr, ok := errors.IsAppError(err); ok && appErr.Code == errors.ErrCodeNotFound {
			s.logger.Warn("Token not found during logout", zap.String("token_hash", tokenHash))
			return nil
		}
		return err
	}
	return nil
}

// LogoutAll invalidates all tokens for a user
func (s *AuthService) LogoutAll(ctx context.Context, userID string) error {
	count, err := s.tokenRepo.DeleteUserTokens(ctx, userID)
	if err != nil {
		return err
	}
	s.logger.Info("All tokens invalidated for user", zap.String("user_id", userID), zap.Int64("count", count))
	return nil
}

// CleanupExpiredTokens removes all expired tokens
func (s *AuthService) CleanupExpiredTokens(ctx context.Context) error {
	count, err := s.tokenRepo.CleanupExpired(ctx)
	if err != nil {
		return err
	}
	if count > 0 {
		s.logger.Info("Cleaned up expired tokens", zap.Int64("count", count))
	}
	return nil
}

// ScheduleTokenCleanup starts a goroutine to periodically clean up expired tokens
func (s *AuthService) ScheduleTokenCleanup(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		for {
			select {
			case <-ticker.C:
				if err := s.CleanupExpiredTokens(ctx); err != nil {
					s.logger.Error("Failed to clean up expired tokens", zap.Error(err))
				}
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
	s.logger.Info("Scheduled token cleanup")
}

// UpdateUserActivity updates a user's last active timestamp
func (s *AuthService) UpdateUserActivity(ctx context.Context, userID string) error {
	return s.userRepo.UpdateLastActive(ctx, userID)
}
