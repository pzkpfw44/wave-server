package domain

import (
	"encoding/base64"
	"time"
)

// User represents a user of the application
type User struct {
	UserID              string    `json:"user_id"`
	Username            string    `json:"username"`
	PublicKey           []byte    `json:"-"` // Don't include binary data in JSON
	EncryptedPrivateKey []byte    `json:"-"` // Don't include binary data in JSON
	Salt                []byte    `json:"-"` // Don't include binary data in JSON
	CreatedAt           time.Time `json:"created_at"`
	LastActive          time.Time `json:"last_active"`
}

// PublicUser is a user safe for public API responses
type PublicUser struct {
	UserID     string    `json:"user_id"`
	Username   string    `json:"username"`
	PublicKey  string    `json:"public_key"` // Base64 encoded
	CreatedAt  time.Time `json:"created_at"`
	LastActive time.Time `json:"last_active"`
}

// ToPublic converts a User to a PublicUser for API responses
func (u *User) ToPublic() PublicUser {
	return PublicUser{
		UserID:     u.UserID,
		Username:   u.Username,
		PublicKey:  base64.URLEncoding.EncodeToString(u.PublicKey),
		CreatedAt:  u.CreatedAt,
		LastActive: u.LastActive,
	}
}

// PrivateKeyResponse provides the encrypted private key and salt
type PrivateKeyResponse struct {
	EncryptedPrivateKey string `json:"encrypted_private_key"` // Base64 encoded
	Salt                string `json:"salt"`                  // Base64 encoded
}

// GetPrivateKeyResponse returns the encrypted private key and salt
func (u *User) GetPrivateKeyResponse() PrivateKeyResponse {
	return PrivateKeyResponse{
		EncryptedPrivateKey: base64.URLEncoding.EncodeToString(u.EncryptedPrivateKey),
		Salt:                base64.URLEncoding.EncodeToString(u.Salt),
	}
}
