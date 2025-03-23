package response

// TokenResponse is the response for token requests
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"` // Seconds
}

// SessionResponse contains the user session status
type SessionResponse struct {
	LoggedIn bool   `json:"logged_in"`
	Username string `json:"username,omitempty"`
}
