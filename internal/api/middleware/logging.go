package middleware

import (
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// LoggingMiddleware handles request logging
type LoggingMiddleware struct {
	logger *zap.Logger
}

// NewLoggingMiddleware creates a new logging middleware
func NewLoggingMiddleware(logger *zap.Logger) *LoggingMiddleware {
	return &LoggingMiddleware{
		logger: logger.With(zap.String("middleware", "logging")),
	}
}

// Logger middleware logs requests
func (m *LoggingMiddleware) Logger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			req := c.Request()

			// Set request ID in context
			requestID := req.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = c.Response().Header().Get(echo.HeaderXRequestID)
			}

			// Process the request
			err := next(c)

			// Log after response
			latency := time.Since(start)
			status := c.Response().Status

			// Skip logging for health check endpoints to reduce noise
			path := c.Path()
			if path == "/health" || path == "/health/liveness" || path == "/health/readiness" {
				return err
			}

			// Get user ID if available
			userID, _ := c.Get("user_id").(string)

			// Log at appropriate level based on status code
			logFunc := m.logger.Info
			if status >= 500 {
				logFunc = m.logger.Error
			} else if status >= 400 {
				logFunc = m.logger.Warn
			}

			// Log the request
			logFunc("HTTP Request",
				zap.String("method", req.Method),
				zap.String("path", path),
				zap.Int("status", status),
				zap.Duration("latency", latency),
				zap.String("ip", c.RealIP()),
				zap.String("user_agent", req.UserAgent()),
				zap.String("request_id", requestID),
				zap.String("user_id", userID), // Empty if not authenticated
			)

			return err
		}
	}
}
