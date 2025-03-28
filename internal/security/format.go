package security

import (
	"errors"
	"fmt"
)

const (
	// Kyber512 constants - adjusted to be more flexible with encrypted sizes
	Kyber512PublicKeyMinSize  = 800  // Approximate minimum size in bytes
	Kyber512PublicKeyMaxSize  = 1000 // Approximate maximum size in bytes
	Kyber512PrivateKeyMinSize = 1200 // Approximate minimum size in bytes
	Kyber512PrivateKeyMaxSize = 2000 // Increased maximum size to account for encryption overhead
)

var (
	ErrInvalidPublicKeyFormat  = errors.New("invalid public key format")
	ErrInvalidPrivateKeyFormat = errors.New("invalid encrypted private key format")
	ErrInvalidSaltFormat       = errors.New("invalid salt format")
)

// ValidatePublicKeyFormat validates that a public key has the expected format
func ValidatePublicKeyFormat(publicKey []byte) error {
	if err := ValidateBytes(publicKey, Kyber512PublicKeyMinSize, Kyber512PublicKeyMaxSize); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidPublicKeyFormat, err)
	}
	return nil
}

// ValidateEncryptedPrivateKeyFormat validates that an encrypted private key has the expected format
func ValidateEncryptedPrivateKeyFormat(encryptedPrivateKey []byte) error {
	if err := ValidateBytes(encryptedPrivateKey, Kyber512PrivateKeyMinSize, Kyber512PrivateKeyMaxSize); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidPrivateKeyFormat, err)
	}
	return nil
}

// ValidateSaltFormat validates that a salt has the expected format
func ValidateSaltFormat(salt []byte) error {
	// Salt should typically be at least 16 bytes for security
	if err := ValidateBytes(salt, 16, 32); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidSaltFormat, err)
	}
	return nil
}
