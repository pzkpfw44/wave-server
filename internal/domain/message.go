package domain

import (
	"encoding/base64"
	"time"

	"github.com/google/uuid"
)

// MessageStatus represents the status of a message
type MessageStatus string

const (
	MessageStatusSent      MessageStatus = "sent"
	MessageStatusDelivered MessageStatus = "delivered"
	MessageStatusRead      MessageStatus = "read"
)

// Message represents an encrypted message
type Message struct {
	MessageID           uuid.UUID     `json:"message_id"`
	SenderPubKey        string        `json:"sender_pubkey"`
	RecipientPubKey     string        `json:"recipient_pubkey"`
	CiphertextKEM       []byte        `json:"-"` // Don't include binary data in JSON
	CiphertextMsg       []byte        `json:"-"` // Don't include binary data in JSON
	Nonce               []byte        `json:"-"` // Don't include binary data in JSON
	SenderCiphertextKEM []byte        `json:"-"` // Don't include binary data in JSON
	SenderCiphertextMsg []byte        `json:"-"` // Don't include binary data in JSON
	SenderNonce         []byte        `json:"-"` // Don't include binary data in JSON
	Timestamp           time.Time     `json:"timestamp"`
	Status              MessageStatus `json:"status"`
}

// MessageResponse is the API response format for a message
type MessageResponse struct {
	MessageID           string        `json:"message_id"`
	SenderPubKey        string        `json:"sender_pubkey"`
	RecipientPubKey     string        `json:"recipient_pubkey"`
	CiphertextKEM       string        `json:"ciphertext_kem"`                  // Base64 encoded
	CiphertextMsg       string        `json:"ciphertext_msg"`                  // Base64 encoded
	Nonce               string        `json:"nonce"`                           // Base64 encoded
	SenderCiphertextKEM string        `json:"sender_ciphertext_kem,omitempty"` // Base64 encoded, optional
	SenderCiphertextMsg string        `json:"sender_ciphertext_msg,omitempty"` // Base64 encoded, optional
	SenderNonce         string        `json:"sender_nonce,omitempty"`          // Base64 encoded, optional
	Timestamp           time.Time     `json:"timestamp"`
	Status              MessageStatus `json:"status"`
}

// ToResponse converts a Message to a MessageResponse
func (m *Message) ToResponse(includeAllFields bool) MessageResponse {
	response := MessageResponse{
		MessageID:       m.MessageID.String(),
		SenderPubKey:    m.SenderPubKey,
		RecipientPubKey: m.RecipientPubKey,
		CiphertextKEM:   base64.URLEncoding.EncodeToString(m.CiphertextKEM),
		CiphertextMsg:   base64.URLEncoding.EncodeToString(m.CiphertextMsg),
		Nonce:           base64.URLEncoding.EncodeToString(m.Nonce),
		Timestamp:       m.Timestamp,
		Status:          m.Status,
	}

	// Only include sender fields if specified (typically only for the sender)
	if includeAllFields {
		response.SenderCiphertextKEM = base64.URLEncoding.EncodeToString(m.SenderCiphertextKEM)
		response.SenderCiphertextMsg = base64.URLEncoding.EncodeToString(m.SenderCiphertextMsg)
		response.SenderNonce = base64.URLEncoding.EncodeToString(m.SenderNonce)
	}

	return response
}

// NewMessage creates a new Message
func NewMessage(senderPubKey, recipientPubKey string,
	ciphertextKEM, ciphertextMsg, nonce []byte,
	senderCiphertextKEM, senderCiphertextMsg, senderNonce []byte) *Message {

	return &Message{
		MessageID:           uuid.New(),
		SenderPubKey:        senderPubKey,
		RecipientPubKey:     recipientPubKey,
		CiphertextKEM:       ciphertextKEM,
		CiphertextMsg:       ciphertextMsg,
		Nonce:               nonce,
		SenderCiphertextKEM: senderCiphertextKEM,
		SenderCiphertextMsg: senderCiphertextMsg,
		SenderNonce:         senderNonce,
		Timestamp:           time.Now(),
		Status:              MessageStatusSent,
	}
}
