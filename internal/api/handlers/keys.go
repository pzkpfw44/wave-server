package handlers

import (
	"encoding/base64"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/pzkpfw44/wave-server/internal/api/middleware"
	"github.com/pzkpfw44/wave-server/internal/api/response"
	"github.com/pzkpfw44/wave-server/internal/errors"
	"github.com/pzkpfw44/wave-server/internal/service"
)

// KeyHandler handles key-related requests
type KeyHandler struct {
	userService *service.UserService
	logger      *zap.Logger
}

// NewKeyHandler creates a new key handler
func NewKeyHandler(
	userService *service.UserService,
	logger *zap.Logger,
) *KeyHandler {
	return &KeyHandler{
		userService: userService,
		logger:      logger.With(zap.String("handler", "key")),
	}
}

// GetPublicKey handles getting a user's public key
func (h *KeyHandler) GetPublicKey(c echo.Context) error {
	// Get username from query
	username := c.QueryParam("username")
	if username == "" {
		// If no username provided, return current user's public key
		userID, err := middleware.GetUserID(c)
		if err != nil {
			return err
		}

		user, err := h.userService.GetByID(c.Request().Context(), userID)
		if err != nil {
			if appErr, ok := errors.IsAppError(err); ok {
				return c.JSON(appErr.Status, response.NewErrorResponse(appErr.Message, appErr.Code))
			}
			h.logger.Error("Get user failed", zap.Error(err))
			return c.JSON(http.StatusInternalServerError, response.NewErrorResponse("Failed to get user information", "INTERNAL"))
		}

		// Return public key
		publicKey := base64.URLEncoding.EncodeToString(user.PublicKey)
		return c.JSON(http.StatusOK, response.NewSuccessResponse(map[string]string{"public_key": publicKey}))
	}

	// Get public key for specified username
	user, err := h.userService.GetByUsername(c.Request().Context(), username)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.JSON(appErr.Status, response.NewErrorResponse(appErr.Message, appErr.Code))
		}
		h.logger.Error("Get user failed", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, response.NewErrorResponse("Failed to get user information", "INTERNAL"))
	}

	// Return public key
	publicKey := base64.URLEncoding.EncodeToString(user.PublicKey)
	return c.JSON(http.StatusOK, response.NewSuccessResponse(map[string]string{"public_key": publicKey}))
}

// GetEncryptedPrivateKey handles getting a user's encrypted private key
func (h *KeyHandler) GetEncryptedPrivateKey(c echo.Context) error {
	// Get user ID from context
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return err
	}

	// Get encrypted private key
	privateKeyResponse, err := h.userService.GetEncryptedPrivateKey(c.Request().Context(), userID)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.JSON(appErr.Status, response.NewErrorResponse(appErr.Message, appErr.Code))
		}
		h.logger.Error("Get encrypted private key failed", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, response.NewErrorResponse("Failed to get encrypted private key", "INTERNAL"))
	}

	// Return encrypted private key
	return c.JSON(http.StatusOK, response.NewSuccessResponse(privateKeyResponse))
}
