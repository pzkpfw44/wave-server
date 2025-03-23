package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"

	"github.com/pzkpfw44/wave-server/internal/domain"
	"github.com/pzkpfw44/wave-server/internal/errors"
)

// UserRepository handles user data storage operations
type UserRepository struct {
	db     *Database
	logger *zap.Logger
}

// NewUserRepository creates a new UserRepository
func NewUserRepository(db *Database) *UserRepository {
	return &UserRepository{
		db:     db,
		logger: db.Logger.With(zap.String("repository", "user")),
	}
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
	INSERT INTO users (user_id, username, public_key, encrypted_private_key, salt, created_at, last_active)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.Pool.Exec(ctx, query,
		user.UserID,
		user.Username,
		user.PublicKey,
		user.EncryptedPrivateKey,
		user.Salt,
		user.CreatedAt,
		user.LastActive,
	)

	if err != nil {
		r.logger.Error("Failed to create user", zap.Error(err), zap.String("username", user.Username))

		// Check for unique constraint violation
		if pgErr, ok := err.(*pgx.PgError); ok && pgErr.Code == "23505" {
			if pgErr.ConstraintName == "users_username_key" {
				return errors.NewConflictError(fmt.Sprintf("User with username '%s' already exists", user.Username))
			}
		}

		return errors.NewInternalError("Failed to create user", err)
	}

	return nil
}

// GetByUsername gets a user by username
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	query := `
	SELECT user_id, username, public_key, encrypted_private_key, salt, created_at, last_active
	FROM users
	WHERE username = $1
	`

	row := r.db.Pool.QueryRow(ctx, query, username)

	user := &domain.User{}
	err := row.Scan(
		&user.UserID,
		&user.Username,
		&user.PublicKey,
		&user.EncryptedPrivateKey,
		&user.Salt,
		&user.CreatedAt,
		&user.LastActive,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.NewNotFoundError(fmt.Sprintf("User with username '%s'", username))
		}
		r.logger.Error("Failed to get user by username", zap.Error(err), zap.String("username", username))
		return nil, errors.NewInternalError("Failed to get user", err)
	}

	return user, nil
}

// GetByID gets a user by ID
func (r *UserRepository) GetByID(ctx context.Context, userID string) (*domain.User, error) {
	query := `
	SELECT user_id, username, public_key, encrypted_private_key, salt, created_at, last_active
	FROM users
	WHERE user_id = $1
	`

	row := r.db.Pool.QueryRow(ctx, query, userID)

	user := &domain.User{}
	err := row.Scan(
		&user.UserID,
		&user.Username,
		&user.PublicKey,
		&user.EncryptedPrivateKey,
		&user.Salt,
		&user.CreatedAt,
		&user.LastActive,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.NewNotFoundError(fmt.Sprintf("User with ID '%s'", userID))
		}
		r.logger.Error("Failed to get user by ID", zap.Error(err), zap.String("user_id", userID))
		return nil, errors.NewInternalError("Failed to get user", err)
	}

	return user, nil
}

// UpdateLastActive updates a user's last active timestamp
func (r *UserRepository) UpdateLastActive(ctx context.Context, userID string) error {
	query := `
	UPDATE users
	SET last_active = $1
	WHERE user_id = $2
	`

	_, err := r.db.Pool.Exec(ctx, query, time.Now(), userID)
	if err != nil {
		r.logger.Error("Failed to update user's last active timestamp", zap.Error(err), zap.String("user_id", userID))
		return errors.NewInternalError("Failed to update user", err)
	}

	return nil
}

// Delete deletes a user
func (r *UserRepository) Delete(ctx context.Context, userID string) error {
	query := `
	DELETE FROM users
	WHERE user_id = $1
	`

	result, err := r.db.Pool.Exec(ctx, query, userID)
	if err != nil {
		r.logger.Error("Failed to delete user", zap.Error(err), zap.String("user_id", userID))
		return errors.NewInternalError("Failed to delete user", err)
	}

	if result.RowsAffected() == 0 {
		return errors.NewNotFoundError(fmt.Sprintf("User with ID '%s'", userID))
	}

	return nil
}
