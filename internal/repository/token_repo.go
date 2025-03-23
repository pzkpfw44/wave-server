package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"

	"github.com/yourusername/wave-server/internal/domain"
	"github.com/yourusername/wave-server/internal/errors"
	"github.com/yourusername/wave-server/internal/security"
)

// TokenRepository handles token data storage operations
type TokenRepository struct {
	db     *Database
	logger *zap.Logger
}

// NewTokenRepository creates a new TokenRepository
func NewTokenRepository(db *Database) *TokenRepository {
	return &TokenRepository{
		db:     db,
		logger: db.Logger.With(zap.String("repository", "token")),
	}
}

// Create creates a new token
func (r *TokenRepository) Create(ctx context.Context, token *domain.Token) error {
	query := `
	INSERT INTO tokens (token_id, user_id, token_hash, created_at, expires_at, last_used)
	VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.Pool.Exec(ctx, query,
		token.TokenID,
		token.UserID,
		token.TokenHash,
		token.CreatedAt,
		token.ExpiresAt,
		token.LastUsed,
	)

	if err != nil {
		r.logger.Error("Failed to create token",
			zap.Error(err),
			zap.String("user_id", token.UserID),
			zap.String("token_id", token.TokenID.String()))

		// Check for unique constraint violation
		if pgErr, ok := err.(*pgx.PgError); ok && pgErr.Code == "23505" {
			return errors.NewConflictError("Token hash already exists")
		}

		return errors.NewInternalError("Failed to create token", err)
	}

	return nil
}

// GetByTokenHash gets a token by its hash
func (r *TokenRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*domain.Token, error) {
	query := `
	SELECT token_id, user_id, token_hash, created_at, expires_at, last_used
	FROM tokens
	WHERE token_hash = $1
	`

	row := r.db.Pool.QueryRow(ctx, query, tokenHash)

	token := &domain.Token{}
	err := row.Scan(
		&token.TokenID,
		&token.UserID,
		&token.TokenHash,
		&token.CreatedAt,
		&token.ExpiresAt,
		&token.LastUsed,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.NewNotFoundError("Token")
		}
		r.logger.Error("Failed to get token by hash", zap.Error(err))
		return nil, errors.NewInternalError("Failed to get token", err)
	}

	return token, nil
}

// GetByUserID gets all tokens for a user
func (r *TokenRepository) GetByUserID(ctx context.Context, userID string) ([]*domain.Token, error) {
	query := `
	SELECT token_id, user_id, token_hash, created_at, expires_at, last_used
	FROM tokens
	WHERE user_id = $1
	ORDER BY created_at DESC
	`

	rows, err := r.db.Pool.Query(ctx, query, userID)
	if err != nil {
		r.logger.Error("Failed to get tokens by user ID", zap.Error(err), zap.String("user_id", userID))
		return nil, errors.NewInternalError("Failed to get tokens", err)
	}
	defer rows.Close()

	var tokens []*domain.Token
	for rows.Next() {
		token := &domain.Token{}
		err := rows.Scan(
			&token.TokenID,
			&token.UserID,
			&token.TokenHash,
			&token.CreatedAt,
			&token.ExpiresAt,
			&token.LastUsed,
		)
		if err != nil {
			r.logger.Error("Failed to scan token row", zap.Error(err))
			return nil, errors.NewInternalError("Failed to read token data", err)
		}
		tokens = append(tokens, token)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("Error iterating token rows", zap.Error(err))
		return nil, errors.NewInternalError("Failed to read token data", err)
	}

	return tokens, nil
}

// UpdateLastUsed updates a token's last_used timestamp
func (r *TokenRepository) UpdateLastUsed(ctx context.Context, tokenID uuid.UUID) error {
	query := `
	UPDATE tokens
	SET last_used = $1
	WHERE token_id = $2
	`

	_, err := r.db.Pool.Exec(ctx, query, time.Now(), tokenID)
	if err != nil {
		r.logger.Error("Failed to update token's last_used timestamp",
			zap.Error(err),
			zap.String("token_id", tokenID.String()))
		return errors.NewInternalError("Failed to update token", err)
	}

	return nil
}

// Delete deletes a token
func (r *TokenRepository) Delete(ctx context.Context, tokenHash string) error {
	query := `
	DELETE FROM tokens
	WHERE token_hash = $1
	`

	result, err := r.db.Pool.Exec(ctx, query, tokenHash)
	if err != nil {
		r.logger.Error("Failed to delete token", zap.Error(err))
		return errors.NewInternalError("Failed to delete token", err)
	}

	if result.RowsAffected() == 0 {
		return errors.NewNotFoundError("Token")
	}

	return nil
}

// DeleteUserTokens deletes all tokens for a user
func (r *TokenRepository) DeleteUserTokens(ctx context.Context, userID string) (int64, error) {
	query := `
	DELETE FROM tokens
	WHERE user_id = $1
	`

	result, err := r.db.Pool.Exec(ctx, query, userID)
	if err != nil {
		r.logger.Error("Failed to delete user tokens", zap.Error(err), zap.String("user_id", userID))
		return 0, errors.NewInternalError("Failed to delete tokens", err)
	}

	return result.RowsAffected(), nil
}

// CleanupExpired deletes all expired tokens
func (r *TokenRepository) CleanupExpired(ctx context.Context) (int64, error) {
	query := `
	DELETE FROM tokens
	WHERE expires_at < $1
	`

	result, err := r.db.Pool.Exec(ctx, query, time.Now())
	if err != nil {
		r.logger.Error("Failed to cleanup expired tokens", zap.Error(err))
		return 0, errors.NewInternalError("Failed to cleanup tokens", err)
	}

	count := result.RowsAffected()
	if count > 0 {
		r.logger.Info("Cleaned up expired tokens", zap.Int64("count", count))
	}

	return count, nil
}

// CreateForUser generates a new token for a user
func (r *TokenRepository) CreateForUser(ctx context.Context, userID string, expiryDuration time.Duration) (string, error) {
	// Generate a random token
	tokenStr, err := security.GenerateRandomToken(32) // 32 bytes = 64 hex chars
	if err != nil {
		r.logger.Error("Failed to generate random token", zap.Error(err))
		return "", errors.NewInternalError("Failed to create token", err)
	}

	// Hash the token for storage
	tokenHash := security.HashToken(tokenStr)

	// Create the token
	token := domain.NewToken(userID, tokenHash, time.Now().Add(expiryDuration))

	// Store the token
	err = r.Create(ctx, token)
	if err != nil {
		return "", fmt.Errorf("failed to store token: %w", err)
	}

	// Return the unhashed token to the caller
	return tokenStr, nil
}

// ValidateToken validates a token and returns the user ID
func (r *TokenRepository) ValidateToken(ctx context.Context, tokenStr string) (string, error) {
	// Hash the token for lookup
	tokenHash := security.HashToken(tokenStr)

	// Get the token
	token, err := r.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return "", errors.NewUnauthenticatedError("Invalid or expired token")
	}

	// Check if the token is expired
	if token.IsExpired() {
		// Try to delete the expired token
		_ = r.Delete(ctx, tokenHash)
		return "", errors.NewUnauthenticatedError("Token expired")
	}

	// Update last used timestamp
	_ = r.UpdateLastUsed(ctx, token.TokenID)

	return token.UserID, nil
}
