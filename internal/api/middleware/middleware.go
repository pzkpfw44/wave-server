package middleware

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"

	"github.com/pzkpfw44/wave-server/internal/api/request"
	"github.com/pzkpfw44/wave-server/internal/config"
	"github.com/pzkpfw44/wave-server/internal/service"
)

// SetupMiddleware configures all middleware for the API
func SetupMiddleware(e *echo.Echo, cfg *config.Config, logger *zap.Logger, authService *service.AuthService) {
	// Create middleware instances
	recoveryMiddleware := NewRecoveryMiddleware(logger)
	loggingMiddleware := NewLoggingMiddleware(logger)
	corsMiddleware := NewCORSMiddleware(logger, cfg)
	authMiddleware := NewAuthMiddleware(authService, logger)
	metricsMiddleware := NewMetricsMiddleware(logger)

	// Setup rate limiters
	// General rate limiter: 100 requests per minute
	generalRateLimiter := NewRateLimiter(100, time.Minute, logger)
	// Auth rate limiter: 20 requests per 5 minutes
	authRateLimiter := NewRateLimiter(20, 5*time.Minute, logger)

	// Set custom validator
	e.Validator = request.NewValidator(logger)

	// Apply global middleware
	e.Use(middleware.RequestID())
	e.Use(recoveryMiddleware.Recover())
	e.Use(loggingMiddleware.Logger())
	e.Use(corsMiddleware.CORS())
	e.Use(middleware.Secure())
	e.Use(generalRateLimiter.Limit())
	e.Use(metricsMiddleware.Metrics())

	// Register metrics endpoint
	metricsMiddleware.SetupMetricsEndpoint(e)

	// Create auth group
	authGroup := e.Group("/api/v1/auth")
	authGroup.Use(authRateLimiter.Limit())

	// Create protected group for authenticated endpoints
	protectedGroup := e.Group("/api/v1")
	protectedGroup.Use(authMiddleware.Authenticate())

	// Extra middleware for specific endpoints can be added later
}

// AuthOnly returns the auth middleware for routes that need it
func AuthOnly(authService *service.AuthService, logger *zap.Logger) echo.MiddlewareFunc {
	authMiddleware := NewAuthMiddleware(authService, logger)
	return authMiddleware.Authenticate()
}
