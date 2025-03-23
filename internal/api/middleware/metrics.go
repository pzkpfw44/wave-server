package middleware

import (
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/pzkpfw44/wave-server/pkg/metrics"
)

// MetricsMiddleware handles metrics collection
type MetricsMiddleware struct {
	logger *zap.Logger
}

// NewMetricsMiddleware creates a new metrics middleware
func NewMetricsMiddleware(logger *zap.Logger) *MetricsMiddleware {
	return &MetricsMiddleware{
		logger: logger.With(zap.String("middleware", "metrics")),
	}
}

// Metrics middleware collects metrics for HTTP requests
func (m *MetricsMiddleware) Metrics() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			// Process the request
			err := next(c)

			// Skip metrics for health check endpoints to reduce noise
			path := c.Path()
			if path == "/health" || path == "/health/liveness" || path == "/health/readiness" {
				return err
			}

			// Record metrics
			duration := time.Since(start).Seconds()
			status := c.Response().Status
			method := c.Request().Method
			responseSize := c.Response().Size

			// Record HTTP metrics
			metrics.RecordRequestMetrics(method, path, status, duration, int(responseSize))

			// If this is an error, record it
			if status >= 400 {
				errorType := "client_error"
				if status >= 500 {
					errorType = "server_error"
				}
				metrics.RecordError(errorType)
			}

			return err
		}
	}
}

// SetupMetricsEndpoint registers metrics endpoint with Echo
func (m *MetricsMiddleware) SetupMetricsEndpoint(e *echo.Echo) {
	metrics.RegisterMetricsHandler(e)
	m.logger.Info("Metrics endpoint registered at /metrics")
}
