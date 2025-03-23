package middleware

import (
	"fmt"
	"net/http"
	"runtime"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/yourusername/wave-server/internal/api/response"
)

// RecoveryMiddleware handles panic recovery
type RecoveryMiddleware struct {
	logger *zap.Logger
}

// NewRecoveryMiddleware creates a new recovery middleware
func NewRecoveryMiddleware(logger *zap.Logger) *RecoveryMiddleware {
	return &RecoveryMiddleware{
		logger: logger.With(zap.String("middleware", "recovery")),
	}
}

// Recover middleware recovers from panics
func (m *RecoveryMiddleware) Recover() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					err, ok := r.(error)
					if !ok {
						err = fmt.Errorf("%v", r)
					}

					// Get stack trace
					stack := make([]byte, 4096)
					length := runtime.Stack(stack, false)
					stackTrace := string(stack[:length])

					// Log the panic
					m.logger.Error("Panic recovered",
						zap.Error(err),
						zap.String("stack", stackTrace),
						zap.String("method", c.Request().Method),
						zap.String("path", c.Request().URL.Path),
						zap.String("client_ip", c.RealIP()),
					)

					// Return a generic error to the client
					resp := response.NewErrorResponse(
						"Internal server error",
						"INTERNAL",
					)
					_ = c.JSON(http.StatusInternalServerError, resp)
				}
			}()
			return next(c)
		}
	}
}
