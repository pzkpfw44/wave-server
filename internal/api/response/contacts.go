package response

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
