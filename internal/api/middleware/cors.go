package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"

	"github.com/pzkpfw44/wave-server/internal/config"
)

// CORSMiddleware handles CORS configuration
type CORSMiddleware struct {
	logger *zap.Logger
	config *config.Config
}

// NewCORSMiddleware creates a new CORS middleware
func NewCORSMiddleware(logger *zap.Logger, config *config.Config) *CORSMiddleware {
	return &CORSMiddleware{
		logger: logger.With(zap.String("middleware", "cors")),
		config: config,
	}
}

// CORS configures CORS middleware
func (m *CORSMiddleware) CORS() echo.MiddlewareFunc {
	return middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     m.config.Server.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type", "X-Requested-With"},
		AllowCredentials: true,
		MaxAge:           86400, // 24 hours
	})
}
