package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/yourusername/wave-server/internal/api/middleware"
	"github.com/yourusername/wave-server/internal/api/request"
	"github.com/yourusername/wave-server/internal/api/response"
	"github.com/yourusername/wave-server/internal/errors"
	"github.com/yourusername/wave-server/internal/service"
)

// AccountHandler handles account-related requests
type AccountHandler struct {
	accountService *service.AccountService
	authService    *service.AuthService
	logger         *zap.Logger
}

// NewAccountHandler creates a new account handler
func NewAccountHandler(
	accountService *service.AccountService,
	authService *service.AuthService,
	logger *zap.Logger,
) *AccountHandler {
	return &AccountHandler{
		accountService: accountService,
		authService:    authService,
		logger:         logger.With(zap.String("handler", "account")),
	}
}

// BackupAccount handles account backup
func (h *AccountHandler) BackupAccount(c echo.Context) error {
	// Get user ID from context
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return err
	}

	// Get account backup
	backup, err := h.accountService.BackupAccount(c.Request().Context(), userID)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.JSON(appErr.Status, response.NewErrorResponse(appErr.Message, appErr.Code))
		}
		h.logger.Error("Backup account failed", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, response.NewErrorResponse("Failed to backup account", "INTERNAL"))
	}

	// Return backup data
	backupResponse := response.BackupResponse{
		PublicKey:           backup.PublicKey,
		EncryptedPrivateKey: backup.EncryptedPrivateKey,
		Contacts:            backup.Contacts,
		Messages:            backup.Messages,
	}

	return c.JSON(http.StatusOK, response.NewSuccessResponse(backupResponse))
}

// RecoverAccount handles account recovery
func (h *AccountHandler) RecoverAccount(c echo.Context) error {
	// Validate request
	var req request.RecoverAccountRequest
	if err := request.ValidateRequest(c, &req); err != nil {
		return err
	}

	// Recover account
	user, err := h.accountService.RecoverAccount(
		c.Request().Context(),
		req.Username,
		req.PublicKey,
		req.EncryptedPrivateKey,
		req.Contacts,
		req.Messages,
	)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.JSON(appErr.Status, response.NewErrorResponse(appErr.Message, appErr.Code))
		}
		h.logger.Error("Recover account failed", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, response.NewErrorResponse("Failed to recover account", "INTERNAL"))
	}

	// Return success
	userResponse := response.UserPublicResponse{
		UserID:     user.UserID,
		Username:   user.Username,
		PublicKey:  req.PublicKey, // Use the provided public key to avoid base64 re-encoding
		CreatedAt:  user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		LastActive: user.LastActive.Format("2006-01-02T15:04:05Z07:00"),
	}

	return c.JSON(http.StatusOK, response.NewSuccessResponse(userResponse))
}

// DeleteAccount handles account deletion
func (h *AccountHandler) DeleteAccount(c echo.Context) error {
	// Get user ID from context
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return err
	}

	// Delete account
	if err := h.accountService.DeleteAccount(c.Request().Context(), userID); err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.JSON(appErr.Status, response.NewErrorResponse(appErr.Message, appErr.Code))
		}
		h.logger.Error("Delete account failed", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, response.NewErrorResponse("Failed to delete account", "INTERNAL"))
	}

	// Return success
	return c.JSON(http.StatusOK, response.NewSuccessResponse(map[string]bool{"deleted": true}))
}
