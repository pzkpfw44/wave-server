package response

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
