package request

// RegisterRequest is the request body for user registration
type RegisterRequest struct {
	Username            string `json:"username" validate:"required,min=3,max=50"`
	PublicKey           string `json:"public_key" validate:"required"`
	EncryptedPrivateKey string `json:"encrypted_private_key" validate:"required"`
	Salt                string `json:"salt" validate:"required"`
}

// LoginRequest is the request body for user login
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
}

// RefreshTokenRequest is the request body for token refresh
// The token itself is sent in the Authorization header
type RefreshTokenRequest struct {
	// Empty struct as the token is in the header
}
