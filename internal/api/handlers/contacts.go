package handlers

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/pzkpfw44/wave-server/internal/api/middleware"
	"github.com/pzkpfw44/wave-server/internal/api/request"
	"github.com/pzkpfw44/wave-server/internal/api/response"
	"github.com/pzkpfw44/wave-server/internal/errors"
	"github.com/pzkpfw44/wave-server/internal/service"
)

// ContactHandler handles contact-related requests
type ContactHandler struct {
	contactService *service.ContactService
	logger         *zap.Logger
}

// NewContactHandler creates a new contact handler
func NewContactHandler(
	contactService *service.ContactService,
	logger *zap.Logger,
) *ContactHandler {
	return &ContactHandler{
		contactService: contactService,
		logger:         logger.With(zap.String("handler", "contact")),
	}
}

// AddContact handles adding a contact
func (h *ContactHandler) AddContact(c echo.Context) error {
	// Get user ID from context
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return err
	}

	// Validate request
	var req request.AddContactRequest
	if err := request.ValidateRequest(c, &req); err != nil {
		return err
	}

	// Add contact
	contact, err := h.contactService.AddContact(c.Request().Context(), userID, req.ContactPublicKey, req.Nickname)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.JSON(appErr.Status, response.NewErrorResponse(appErr.Message, appErr.Code))
		}
		h.logger.Error("Add contact failed", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, response.NewErrorResponse("Failed to add contact", "INTERNAL"))
	}

	// Format response
	contactResponse := response.ContactResponse{
		ContactPubKey: contact.ContactPubKey,
		Nickname:      contact.Nickname,
		CreatedAt:     contact.CreatedAt.Format(time.RFC3339),
	}

	return c.JSON(http.StatusCreated, response.NewSuccessResponse(contactResponse))
}

// GetContacts gets all contacts for the current user
func (h *ContactHandler) GetContacts(c echo.Context) error {
	// Get user ID from context
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return err
	}

	// Get contacts
	contacts, err := h.contactService.GetContacts(c.Request().Context(), userID)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.JSON(appErr.Status, response.NewErrorResponse(appErr.Message, appErr.Code))
		}
		h.logger.Error("Get contacts failed", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, response.NewErrorResponse("Failed to get contacts", "INTERNAL"))
	}

	// Format contacts for response
	contactResponses := make([]response.ContactResponse, len(contacts))
	for i, contact := range contacts {
		contactResponses[i] = response.ContactResponse{
			ContactPubKey: contact.ContactPubKey,
			Nickname:      contact.Nickname,
			CreatedAt:     contact.CreatedAt.Format(time.RFC3339),
		}
	}

	// Return contacts
	contactsResponse := response.ContactsResponse{
		Contacts: contactResponses,
	}

	return c.JSON(http.StatusOK, response.NewSuccessResponse(contactsResponse))
}

// GetContact gets a specific contact
func (h *ContactHandler) GetContact(c echo.Context) error {
	// Get user ID from context
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return err
	}

	// Get contact public key from path
	contactPubKey := c.Param("pubkey")
	if contactPubKey == "" {
		return c.JSON(http.StatusBadRequest, response.NewErrorResponse("Contact public key is required", "BAD_REQUEST"))
	}

	// Get contact
	contact, err := h.contactService.GetContact(c.Request().Context(), userID, contactPubKey)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.JSON(appErr.Status, response.NewErrorResponse(appErr.Message, appErr.Code))
		}
		h.logger.Error("Get contact failed", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, response.NewErrorResponse("Failed to get contact", "INTERNAL"))
	}

	// Format response
	contactResponse := response.ContactResponse{
		ContactPubKey: contact.ContactPubKey,
		Nickname:      contact.Nickname,
		CreatedAt:     contact.CreatedAt.Format(time.RFC3339),
	}

	return c.JSON(http.StatusOK, response.NewSuccessResponse(contactResponse))
}

// UpdateContact updates a contact
func (h *ContactHandler) UpdateContact(c echo.Context) error {
	// Get user ID from context
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return err
	}

	// Get contact public key from path
	contactPubKey := c.Param("pubkey")
	if contactPubKey == "" {
		return c.JSON(http.StatusBadRequest, response.NewErrorResponse("Contact public key is required", "BAD_REQUEST"))
	}

	// Validate request
	var req request.UpdateContactRequest
	if err := request.ValidateRequest(c, &req); err != nil {
		return err
	}

	// Update contact
	contact, err := h.contactService.UpdateContact(c.Request().Context(), userID, contactPubKey, req.Nickname)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.JSON(appErr.Status, response.NewErrorResponse(appErr.Message, appErr.Code))
		}
		h.logger.Error("Update contact failed", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, response.NewErrorResponse("Failed to update contact", "INTERNAL"))
	}

	// Format response
	contactResponse := response.ContactResponse{
		ContactPubKey: contact.ContactPubKey,
		Nickname:      contact.Nickname,
		CreatedAt:     contact.CreatedAt.Format(time.RFC3339),
	}

	return c.JSON(http.StatusOK, response.NewSuccessResponse(contactResponse))
}

// DeleteContact deletes a contact
func (h *ContactHandler) DeleteContact(c echo.Context) error {
	// Get user ID from context
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return err
	}

	// Get contact public key from path
	contactPubKey := c.Param("pubkey")
	if contactPubKey == "" {
		return c.JSON(http.StatusBadRequest, response.NewErrorResponse("Contact public key is required", "BAD_REQUEST"))
	}

	// Delete contact
	if err := h.contactService.DeleteContact(c.Request().Context(), userID, contactPubKey); err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.JSON(appErr.Status, response.NewErrorResponse(appErr.Message, appErr.Code))
		}
		h.logger.Error("Delete contact failed", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, response.NewErrorResponse("Failed to delete contact", "INTERNAL"))
	}

	return c.JSON(http.StatusOK, response.NewSuccessResponse(map[string]bool{"deleted": true}))
}
