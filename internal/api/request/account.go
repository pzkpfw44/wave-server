package request

// RecoverAccountRequest is the request body for account recovery
type RecoverAccountRequest struct {
	Username            string                 `json:"username" validate:"required,min=3,max=50"`
	PublicKey           string                 `json:"public_key" validate:"required"`
	EncryptedPrivateKey map[string]string      `json:"encrypted_private_key" validate:"required"`
	Contacts            map[string]interface{} `json:"contacts,omitempty"`
	Messages            []interface{}          `json:"messages,omitempty"`
}

// DeleteAccountRequest is an empty request for deleting an account
// No request body is needed as the user is identified by their token
type DeleteAccountRequest struct {
	// Empty struct
}
