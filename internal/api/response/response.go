package response

// Response is the base API response format
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

// ErrorInfo contains error details
type ErrorInfo struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}

// NewSuccessResponse creates a new success response
func NewSuccessResponse(data interface{}) Response {
	return Response{
		Success: true,
		Data:    data,
	}
}

// NewErrorResponse creates a new error response
func NewErrorResponse(message, code string) Response {
	return Response{
		Success: false,
		Error: &ErrorInfo{
			Message: message,
			Code:    code,
		},
	}
}

// TokenResponse is the response for token requests
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"` // Seconds
}

// MessageResponse is the response for message operations
type MessageResponse struct {
	MessageID           string `json:"message_id"`
	SenderPubKey        string `json:"sender_pubkey"`
	RecipientPubKey     string `json:"recipient_pubkey"`
	CiphertextKEM       string `json:"ciphertext_kem"`
	CiphertextMsg       string `json:"ciphertext_msg"`
	Nonce               string `json:"nonce"`
	SenderCiphertextKEM string `json:"sender_ciphertext_kem,omitempty"`
	SenderCiphertextMsg string `json:"sender_ciphertext_msg,omitempty"`
	SenderNonce         string `json:"sender_nonce,omitempty"`
	Timestamp           string `json:"timestamp"`
	Status              string `json:"status"`
}

// MessagesResponse is the response for listing messages
type MessagesResponse struct {
	Messages []MessageResponse `json:"messages"`
	Total    int               `json:"total"`
	Limit    int               `json:"limit"`
	Offset   int               `json:"offset"`
}

// ContactResponse is the response for contact operations
type ContactResponse struct {
	ContactPubKey string `json:"contact_pubkey"`
	Nickname      string `json:"nickname"`
	CreatedAt     string `json:"created_at"`
}

// ContactsResponse is the response for listing contacts
type ContactsResponse struct {
	Contacts []ContactResponse `json:"contacts"`
}
