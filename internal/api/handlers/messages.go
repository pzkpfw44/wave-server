package handlers

import (
	"encoding/base64"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/yourusername/wave-server/internal/api/middleware"
	"github.com/yourusername/wave-server/internal/api/request"
	"github.com/yourusername/wave-server/internal/api/response"
	"github.com/yourusername/wave-server/internal/domain"
	"github.com/yourusername/wave-server/internal/errors"
	"github.com/yourusername/wave-server/internal/service"
)

// MessageHandler handles message-related requests
type MessageHandler struct {
	messageService *service.MessageService
	userService    *service.UserService
	logger         *zap.Logger
}

// NewMessageHandler creates a new message handler
func NewMessageHandler(
	messageService *service.MessageService,
	userService *service.UserService,
	logger *zap.Logger,
) *MessageHandler {
	return &MessageHandler{
		messageService: messageService,
		userService:    userService,
		logger:         logger.With(zap.String("handler", "message")),
	}
}

// SendMessage handles sending a message
func (h *MessageHandler) SendMessage(c echo.Context) error {
	// Get user ID from context
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return err
	}

	// Validate request
	var req request.SendMessageRequest
	if err := request.ValidateRequest(c, &req); err != nil {
		return err
	}

	// Send message
	msg, err := h.messageService.SendMessage(
		c.Request().Context(),
		userID,
		req.RecipientPubKey,
		req.CiphertextKEM,
		req.CiphertextMsg,
		req.Nonce,
		req.SenderCiphertextKEM,
		req.SenderCiphertextMsg,
		req.SenderNonce,
	)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.JSON(appErr.Status, response.NewErrorResponse(appErr.Message, appErr.Code))
		}
		h.logger.Error("Send message failed", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, response.NewErrorResponse("Failed to send message", "INTERNAL"))
	}

	// Construct response
	msgResponse := response.MessageResponse{
		MessageID:           msg.MessageID.String(),
		SenderPubKey:        msg.SenderPubKey,
		RecipientPubKey:     msg.RecipientPubKey,
		CiphertextKEM:       base64.URLEncoding.EncodeToString(msg.CiphertextKEM),
		CiphertextMsg:       base64.URLEncoding.EncodeToString(msg.CiphertextMsg),
		Nonce:               base64.URLEncoding.EncodeToString(msg.Nonce),
		SenderCiphertextKEM: base64.URLEncoding.EncodeToString(msg.SenderCiphertextKEM),
		SenderCiphertextMsg: base64.URLEncoding.EncodeToString(msg.SenderCiphertextMsg),
		SenderNonce:         base64.URLEncoding.EncodeToString(msg.SenderNonce),
		Timestamp:           msg.Timestamp.Format(time.RFC3339),
		Status:              string(msg.Status),
	}

	return c.JSON(http.StatusCreated, response.NewSuccessResponse(msgResponse))
}

// GetMessages gets messages for the current user
func (h *MessageHandler) GetMessages(c echo.Context) error {
	// Get user ID from context
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return err
	}

	// Parse query parameters
	var req request.GetMessagesRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.NewErrorResponse("Invalid query parameters", "BAD_REQUEST"))
	}

	// Set defaults
	if req.Limit <= 0 {
		req.Limit = 100
	}
	if req.Limit > 1000 {
		req.Limit = 1000
	}

	// Get user to get public key
	user, err := h.userService.GetByID(c.Request().Context(), userID)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.JSON(appErr.Status, response.NewErrorResponse(appErr.Message, appErr.Code))
		}
		h.logger.Error("Get user failed", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, response.NewErrorResponse("Failed to get user information", "INTERNAL"))
	}

	// Get messages
	userPubKey := base64.URLEncoding.EncodeToString(user.PublicKey)
	messages, err := h.messageService.GetMessagesForUser(c.Request().Context(), userPubKey, req.Limit, req.Offset)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.JSON(appErr.Status, response.NewErrorResponse(appErr.Message, appErr.Code))
		}
		h.logger.Error("Get messages failed", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, response.NewErrorResponse("Failed to get messages", "INTERNAL"))
	}

	// Format messages for response
	messageResponses := make([]response.MessageResponse, len(messages))
	for i, msg := range messages {
		// Determine if all fields should be included
		includeAllFields := msg.SenderPubKey == userPubKey

		// Convert message to response
		msgResp := msg.ToResponse(includeAllFields)

		// Map domain response to API response
		messageResponses[i] = response.MessageResponse{
			MessageID:           msgResp.MessageID,
			SenderPubKey:        msgResp.SenderPubKey,
			RecipientPubKey:     msgResp.RecipientPubKey,
			CiphertextKEM:       msgResp.CiphertextKEM,
			CiphertextMsg:       msgResp.CiphertextMsg,
			Nonce:               msgResp.Nonce,
			SenderCiphertextKEM: msgResp.SenderCiphertextKEM,
			SenderCiphertextMsg: msgResp.SenderCiphertextMsg,
			SenderNonce:         msgResp.SenderNonce,
			Timestamp:           msgResp.Timestamp.Format(time.RFC3339),
			Status:              string(msgResp.Status),
		}
	}

	// Return messages
	messagesResponse := response.MessagesResponse{
		Messages: messageResponses,
		Total:    len(messageResponses),
		Limit:    req.Limit,
		Offset:   req.Offset,
	}

	return c.JSON(http.StatusOK, response.NewSuccessResponse(messagesResponse))
}

// GetConversation gets messages between the current user and another user
func (h *MessageHandler) GetConversation(c echo.Context) error {
	// Get user ID from context
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return err
	}

	// Parse path and query parameters
	contactPubKey := c.Param("pubkey")
	if contactPubKey == "" {
		return c.JSON(http.StatusBadRequest, response.NewErrorResponse("Contact public key is required", "BAD_REQUEST"))
	}

	var queryParams request.GetMessagesRequest
	if err := c.Bind(&queryParams); err != nil {
		return c.JSON(http.StatusBadRequest, response.NewErrorResponse("Invalid query parameters", "BAD_REQUEST"))
	}

	// Set defaults
	if queryParams.Limit <= 0 {
		queryParams.Limit = 100
	}
	if queryParams.Limit > 1000 {
		queryParams.Limit = 1000
	}

	// Get user to get public key
	user, err := h.userService.GetByID(c.Request().Context(), userID)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.JSON(appErr.Status, response.NewErrorResponse(appErr.Message, appErr.Code))
		}
		h.logger.Error("Get user failed", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, response.NewErrorResponse("Failed to get user information", "INTERNAL"))
	}

	// Get conversation
	userPubKey := base64.URLEncoding.EncodeToString(user.PublicKey)
	messages, err := h.messageService.GetConversation(
		c.Request().Context(),
		userPubKey,
		contactPubKey,
		queryParams.Limit,
		queryParams.Offset,
	)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.JSON(appErr.Status, response.NewErrorResponse(appErr.Message, appErr.Code))
		}
		h.logger.Error("Get conversation failed", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, response.NewErrorResponse("Failed to get conversation", "INTERNAL"))
	}

	// Format messages for response
	messageResponses := make([]response.MessageResponse, len(messages))
	for i, msg := range messages {
		// Determine if all fields should be included
		includeAllFields := msg.SenderPubKey == userPubKey

		// Convert message to response
		msgResp := msg.ToResponse(includeAllFields)

		// Map domain response to API response
		messageResponses[i] = response.MessageResponse{
			MessageID:           msgResp.MessageID,
			SenderPubKey:        msgResp.SenderPubKey,
			RecipientPubKey:     msgResp.RecipientPubKey,
			CiphertextKEM:       msgResp.CiphertextKEM,
			CiphertextMsg:       msgResp.CiphertextMsg,
			Nonce:               msgResp.Nonce,
			SenderCiphertextKEM: msgResp.SenderCiphertextKEM,
			SenderCiphertextMsg: msgResp.SenderCiphertextMsg,
			SenderNonce:         msgResp.SenderNonce,
			Timestamp:           msgResp.Timestamp.Format(time.RFC3339),
			Status:              string(msgResp.Status),
		}
	}

	// Return messages
	messagesResponse := response.MessagesResponse{
		Messages: messageResponses,
		Total:    len(messageResponses),
		Limit:    queryParams.Limit,
		Offset:   queryParams.Offset,
	}

	return c.JSON(http.StatusOK, response.NewSuccessResponse(messagesResponse))
}

// UpdateMessageStatus updates a message's status
func (h *MessageHandler) UpdateMessageStatus(c echo.Context) error {
	// Get user ID from context
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return err
	}

	// Parse message ID
	messageIDStr := c.Param("message_id")
	if messageIDStr == "" {
		return c.JSON(http.StatusBadRequest, response.NewErrorResponse("Message ID is required", "BAD_REQUEST"))
	}

	messageID, err := uuid.Parse(messageIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.NewErrorResponse("Invalid message ID format", "BAD_REQUEST"))
	}

	// Parse request body
	var req request.UpdateMessageStatusRequest
	if err := request.ValidateRequest(c, &req); err != nil {
		return err
	}

	// Update message status
	err = h.messageService.UpdateMessageStatus(c.Request().Context(), messageID, domain.MessageStatus(req.Status))
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.JSON(appErr.Status, response.NewErrorResponse(appErr.Message, appErr.Code))
		}
		h.logger.Error("Update message status failed", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, response.NewErrorResponse("Failed to update message status", "INTERNAL"))
	}

	return c.JSON(http.StatusOK, response.NewSuccessResponse(map[string]string{"status": "updated"}))
}
