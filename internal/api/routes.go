package api

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/yourusername/wave-server/internal/api/handlers"
	"github.com/yourusername/wave-server/internal/api/middleware"
	"github.com/yourusername/wave-server/internal/config"
	"github.com/yourusername/wave-server/internal/service"
	"github.com/yourusername/wave-server/pkg/health"
)

// SetupRoutes configures all API routes
func SetupRoutes(e *echo.Echo, h *handlers.Handler, cfg *config.Config, authService *service.AuthService, healthChecker *health.Checker, logger *zap.Logger) {
	// Health check routes
	if healthChecker != nil {
		healthChecker.RegisterHandlers(e)
	}

	// Public routes
	e.GET("/", func(c echo.Context) error {
		return c.JSON(200, map[string]string{
			"message": "Wave API with public-key-based routing",
			"version": "0.1.0",
		})
	})

	// API versioning
	v1 := e.Group("/api/v1")

	// Authentication routes (no auth required)
	auth := v1.Group("/auth")
	auth.POST("/register", h.Auth.Register)
	auth.POST("/login", h.Auth.Login)
	auth.POST("/refresh", h.Auth.RefreshToken)
	auth.POST("/logout", h.Auth.Logout)

	// Account recovery route (no auth required)
	account := v1.Group("/account")
	account.POST("/recover", h.Account.RecoverAccount)

	// Routes requiring authentication
	// Create middleware for authenticated routes
	authMiddleware := middleware.AuthOnly(authService, logger)

	// User routes
	v1.GET("/keys/public", h.Key.GetPublicKey) // This endpoint works with or without auth
	privateKeys := v1.Group("/keys/private", authMiddleware)
	privateKeys.GET("", h.Key.GetEncryptedPrivateKey)

	// Message routes
	messages := v1.Group("/messages", authMiddleware)
	messages.POST("/send", h.Message.SendMessage)
	messages.GET("", h.Message.GetMessages)
	messages.GET("/conversation/:pubkey", h.Message.GetConversation)
	messages.PATCH("/:message_id/status", h.Message.UpdateMessageStatus)

	// Contact routes
	contacts := v1.Group("/contacts", authMiddleware)
	contacts.POST("", h.Contact.AddContact)
	contacts.GET("", h.Contact.GetContacts)
	contacts.GET("/:pubkey", h.Contact.GetContact)
	contacts.PUT("/:pubkey", h.Contact.UpdateContact)
	contacts.DELETE("/:pubkey", h.Contact.DeleteContact)

	// Account management routes
	accountAuth := account.Group("", authMiddleware)
	accountAuth.GET("/backup", h.Account.BackupAccount)
	accountAuth.DELETE("", h.Account.DeleteAccount)

	// Auth routes that require authentication
	authLogoutAll := auth.Group("/logout-all", authMiddleware)
	authLogoutAll.POST("", h.Auth.LogoutAll)

	logger.Info("API routes configured")
}
