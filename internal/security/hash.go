package security

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
)

// HashUsername creates a deterministic hash of a username for use as user_id
func HashUsername(username string) string {
	// Simple SHA-256 hash of username (lowercased for consistency)
	hash := sha256.Sum256([]byte(username))
	return hex.EncodeToString(hash[:])
}

// HashToken creates a secure hash of a token for storage
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// GenerateRandomToken generates a random token string
func GenerateRandomToken(byteLength int) (string, error) {
	randomBytes := make([]byte, byteLength)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(randomBytes), nil
}

// SecureCompare performs a constant-time comparison of two strings
// to prevent timing attacks when comparing sensitive data
func SecureCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// ValidateBytes checks if a byte slice has a valid length for a key
func ValidateBytes(data []byte, expectedMinLength, expectedMaxLength int) error {
	if len(data) < expectedMinLength {
		return fmt.Errorf("data too short, expected at least %d bytes", expectedMinLength)
	}
	if expectedMaxLength > 0 && len(data) > expectedMaxLength {
		return fmt.Errorf("data too long, expected at most %d bytes", expectedMaxLength)
	}
	return nil
}
