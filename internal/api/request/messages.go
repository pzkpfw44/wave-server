package request

// SendMessageRequest is the request body for sending a message
type SendMessageRequest struct {
	RecipientPubKey     string `json:"recipient_pubkey" validate:"required"`
	CiphertextKEM       string `json:"ciphertext_kem" validate:"required"`
	CiphertextMsg       string `json:"ciphertext_msg" validate:"required"`
	Nonce               string `json:"nonce" validate:"required"`
	SenderCiphertextKEM string `json:"sender_ciphertext_kem" validate:"required"`
	SenderCiphertextMsg string `json:"sender_ciphertext_msg" validate:"required"`
	SenderNonce         string `json:"sender_nonce" validate:"required"`
}

// GetMessagesRequest is the query parameters for getting messages
type GetMessagesRequest struct {
	Limit  int `query:"limit" validate:"omitempty,min=1,max=1000"`
	Offset int `query:"offset" validate:"omitempty,min=0"`
}

// GetConversationRequest is the query parameters for getting a conversation
type GetConversationRequest struct {
	ContactPubKey string `param:"pubkey" validate:"required"`
	Limit         int    `query:"limit" validate:"omitempty,min=1,max=1000"`
	Offset        int    `query:"offset" validate:"omitempty,min=0"`
}

// UpdateMessageStatusRequest is the request body for updating a message's status
type UpdateMessageStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=sent delivered read"`
}
