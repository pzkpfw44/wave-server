package domain

import (
	"time"

	"github.com/google/uuid"
)

// Token represents an authentication token
type Token struct {
	TokenID   uuid.UUID `json:"token_id"`
	UserID    string    `json:"user_id"`    // The user who owns this token
	TokenHash string    `json:"token_hash"` // Hash of the token, not the token itself
	CreatedAt time.Time `json:"created_at"` // When the token was created
	ExpiresAt time.Time `json:"expires_at"` // When the token expires
	LastUsed  time.Time `json:"last_used"`  // Last time the token was used
}

// IsExpired checks if the token is expired
func (t *Token) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// NewToken creates a new Token
func NewToken(userID, tokenHash string, expiresAt time.Time) *Token {
	now := time.Now()
	return &Token{
		TokenID:   uuid.New(),
		UserID:    userID,
		TokenHash: tokenHash,
		CreatedAt: now,
		ExpiresAt: expiresAt,
		LastUsed:  now,
	}
}
