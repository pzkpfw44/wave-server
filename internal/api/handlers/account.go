package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/pzkpfw44/wave-server/internal/api/middleware"
	"github.com/pzkpfw44/wave-server/internal/api/request"
	"github.com/pzkpfw44/wave-server/internal/api/response"
	"github.com/pzkpfw44/wave-server/internal/errors"
	"github.com/pzkpfw44/wave-server/internal/service"
)

// AccountHandler handles account management
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

	// Create backup
	backup, err := h.accountService.BackupAccount(c.Request().Context(), userID)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.JSON(appErr.Status, response.NewErrorResponse(appErr.Message, appErr.Code))
		}
		h.logger.Error("Backup failed", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, response.NewErrorResponse("Failed to create backup", "INTERNAL"))
	}

	return c.JSON(http.StatusOK, response.NewSuccessResponse(backup))
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
		h.logger.Error("Recovery failed", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, response.NewErrorResponse("Failed to recover account", "INTERNAL"))
	}

	// Generate token for the recovered account
	token, err := h.authService.Login(c.Request().Context(), user.Username)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.JSON(appErr.Status, response.NewErrorResponse(appErr.Message, appErr.Code))
		}
		h.logger.Error("Token generation failed", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, response.NewErrorResponse("Authentication failed", "INTERNAL"))
	}

	// Return token
	result := map[string]interface{}{
		"user": user.ToPublic(),
		"token": response.TokenResponse{
			AccessToken: token,
			TokenType:   "Bearer",
			ExpiresIn:   86400, // 24 hours
		},
	}

	return c.JSON(http.StatusOK, response.NewSuccessResponse(result))
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
		h.logger.Error("Account deletion failed", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, response.NewErrorResponse("Failed to delete account", "INTERNAL"))
	}

	return c.JSON(http.StatusOK, response.NewSuccessResponse(map[string]bool{"deleted": true}))
}
