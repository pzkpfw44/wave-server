package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/pzkpfw44/wave-server/internal/api/middleware"
	"github.com/pzkpfw44/wave-server/internal/api/request"
	"github.com/pzkpfw44/wave-server/internal/api/response"
	"github.com/pzkpfw44/wave-server/internal/config"
	"github.com/pzkpfw44/wave-server/internal/errors"
	"github.com/pzkpfw44/wave-server/internal/service"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	authService *service.AuthService
	userService *service.UserService
	config      *config.Config
	logger      *zap.Logger
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(
	authService *service.AuthService,
	userService *service.UserService,
	config *config.Config,
	logger *zap.Logger,
) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		userService: userService,
		config:      config,
		logger:      logger.With(zap.String("handler", "auth")),
	}
}

// Register handles user registration
func (h *AuthHandler) Register(c echo.Context) error {
	var req request.RegisterRequest
	if err := request.ValidateRequest(c, &req); err != nil {
		return err
	}

	// Register user
	_, err := h.userService.Register(
		c.Request().Context(),
		req.Username,
		req.PublicKey,
		req.EncryptedPrivateKey,
		req.Salt,
	)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.JSON(appErr.Status, response.NewErrorResponse(appErr.Message, appErr.Code))
		}
		h.logger.Error("Registration failed", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, response.NewErrorResponse("Registration failed", "INTERNAL"))
	}

	// Generate token
	token, err := h.authService.Login(c.Request().Context(), req.Username)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.JSON(appErr.Status, response.NewErrorResponse(appErr.Message, appErr.Code))
		}
		h.logger.Error("Token generation failed", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, response.NewErrorResponse("Authentication failed", "INTERNAL"))
	}

	// Return token
	tokenResponse := response.TokenResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int(h.config.Auth.TokenExpiry.Seconds()),
	}

	return c.JSON(http.StatusCreated, response.NewSuccessResponse(tokenResponse))
}

// Login handles user login
func (h *AuthHandler) Login(c echo.Context) error {
	var req request.LoginRequest
	if err := request.ValidateRequest(c, &req); err != nil {
		return err
	}

	// Generate token
	token, err := h.authService.Login(c.Request().Context(), req.Username)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.JSON(appErr.Status, response.NewErrorResponse(appErr.Message, appErr.Code))
		}
		h.logger.Error("Login failed", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, response.NewErrorResponse("Authentication failed", "INTERNAL"))
	}

	// Return token
	tokenResponse := response.TokenResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int(h.config.Auth.TokenExpiry.Seconds()),
	}

	return c.JSON(http.StatusOK, response.NewSuccessResponse(tokenResponse))
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(c echo.Context) error {
	// Extract token from Authorization header
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return c.JSON(http.StatusUnauthorized, response.NewErrorResponse("Missing authorization header", "UNAUTHENTICATED"))
	}

	// Support both "Bearer token" and just "token" formats
	token := authHeader
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		token = authHeader[7:]
	}

	// Refresh token
	newToken, err := h.authService.RefreshToken(c.Request().Context(), token)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.JSON(appErr.Status, response.NewErrorResponse(appErr.Message, appErr.Code))
		}
		h.logger.Error("Token refresh failed", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, response.NewErrorResponse("Token refresh failed", "INTERNAL"))
	}

	// Return new token
	tokenResponse := response.TokenResponse{
		AccessToken: newToken,
		TokenType:   "Bearer",
		ExpiresIn:   int(h.config.Auth.TokenExpiry.Seconds()),
	}

	return c.JSON(http.StatusOK, response.NewSuccessResponse(tokenResponse))
}

// Logout handles user logout
func (h *AuthHandler) Logout(c echo.Context) error {
	// Extract token from Authorization header
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return c.JSON(http.StatusUnauthorized, response.NewErrorResponse("Missing authorization header", "UNAUTHENTICATED"))
	}

	// Support both "Bearer token" and just "token" formats
	token := authHeader
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		token = authHeader[7:]
	}

	// Invalidate token
	if err := h.authService.Logout(c.Request().Context(), token); err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.JSON(appErr.Status, response.NewErrorResponse(appErr.Message, appErr.Code))
		}
		h.logger.Error("Logout failed", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, response.NewErrorResponse("Logout failed", "INTERNAL"))
	}

	return c.JSON(http.StatusOK, response.NewSuccessResponse(map[string]bool{"logged_out": true}))
}

// LogoutAll invalidates all tokens for the current user
func (h *AuthHandler) LogoutAll(c echo.Context) error {
	// Get user ID from context
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return err
	}

	// Invalidate all tokens
	if err := h.authService.LogoutAll(c.Request().Context(), userID); err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.JSON(appErr.Status, response.NewErrorResponse(appErr.Message, appErr.Code))
		}
		h.logger.Error("Logout all failed", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, response.NewErrorResponse("Failed to logout from all devices", "INTERNAL"))
	}

	return c.JSON(http.StatusOK, response.NewSuccessResponse(map[string]bool{"logged_out_all": true}))
}
