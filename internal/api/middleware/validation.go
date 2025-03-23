package middleware

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"

	"github.com/pzkpfw44/wave-server/internal/api/request"
)

// ValidationMiddleware handles request validation
type ValidationMiddleware struct {
	validator *validator.Validate
	logger    *zap.Logger
}

// NewValidationMiddleware creates a new validation middleware
func NewValidationMiddleware(logger *zap.Logger) *ValidationMiddleware {
	v := validator.New()

	// Use JSON tag names in validation errors
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return &ValidationMiddleware{
		validator: v,
		logger:    logger.With(zap.String("middleware", "validation")),
	}
}

// GetValidator returns the validator for use with Echo
func (m *ValidationMiddleware) GetValidator() *request.CustomValidator {
	return request.NewValidator(m.logger)
}
