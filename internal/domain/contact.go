package domain

import "time"

// Contact represents a user's contact
type Contact struct {
	UserID        string    `json:"user_id"`        // The user who owns this contact
	ContactPubKey string    `json:"contact_pubkey"` // The contact's public key
	Nickname      string    `json:"nickname"`       // Friendly name for the contact
	CreatedAt     time.Time `json:"created_at"`     // When the contact was added
}

// ContactResponse is the API response format for a contact
type ContactResponse struct {
	ContactPubKey string    `json:"contact_pubkey"`
	Nickname      string    `json:"nickname"`
	CreatedAt     time.Time `json:"created_at"`
}

// ToResponse converts a Contact to a ContactResponse
func (c *Contact) ToResponse() ContactResponse {
	return ContactResponse{
		ContactPubKey: c.ContactPubKey,
		Nickname:      c.Nickname,
		CreatedAt:     c.CreatedAt,
	}
}

// NewContact creates a new Contact
func NewContact(userID, contactPubKey, nickname string) *Contact {
	return &Contact{
		UserID:        userID,
		ContactPubKey: contactPubKey,
		Nickname:      nickname,
		CreatedAt:     time.Now(),
	}
}
