package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/pzkpfw44/wave-server/internal/service"
)

// AuthMiddleware handles JWT authentication
type AuthMiddleware struct {
	authService *service.AuthService
	logger      *zap.Logger
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(authService *service.AuthService, logger *zap.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
		logger:      logger.With(zap.String("middleware", "auth")),
	}
}

// Authenticate middleware handles token authentication
func (m *AuthMiddleware) Authenticate() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Extract token from Authorization header
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "Missing authorization header")
			}

			// Support both "Bearer token" and just "token" formats
			token := authHeader
			if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
				token = authHeader[7:] // Remove "Bearer " prefix
			}

			// Create a request-specific context with a longer timeout
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// Validate token
			userID, err := m.authService.ValidateToken(ctx, token)
			if err != nil {
				m.logger.Debug("Authentication failed", zap.Error(err))
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid or expired token")
			}

			// Set user ID in context
			c.Set("user_id", userID)

			// Update activity in a separate goroutine with its own context
			go func() {
				bgCtx, bgCancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer bgCancel()

				if err := m.authService.UpdateUserActivity(bgCtx, userID); err != nil {
					m.logger.Warn("Failed to update user activity", zap.Error(err), zap.String("user_id", userID))
				}
			}()

			return next(c)
		}
	}
}

// GetUserID extracts the authenticated user ID from the context
func GetUserID(c echo.Context) (string, error) {
	userID, ok := c.Get("user_id").(string)
	if !ok || userID == "" {
		return "", echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated")
	}
	return userID, nil
}
