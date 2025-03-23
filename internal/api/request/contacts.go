package request

// AddContactRequest is the request body for adding a contact
type AddContactRequest struct {
	ContactPublicKey string `json:"contact_public_key" validate:"required"`
	Nickname         string `json:"nickname" validate:"required,max=50"`
}

// UpdateContactRequest is the request body for updating a contact
type UpdateContactRequest struct {
	Nickname string `json:"nickname" validate:"required,max=50"`
}

// GetContactRequest is the path parameter for getting a contact
type GetContactRequest struct {
	ContactPubKey string `param:"pubkey" validate:"required"`
}

// DeleteContactRequest is the path parameter for deleting a contact
type DeleteContactRequest struct {
	ContactPubKey string `param:"pubkey" validate:"required"`
}
