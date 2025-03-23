package request

import (
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/yourusername/wave-server/internal/errors"
)

// CustomValidator is a custom validator for Echo
type CustomValidator struct {
	validator *validator.Validate
	logger    *zap.Logger
}

// NewValidator creates a new custom validator
func NewValidator(logger *zap.Logger) *CustomValidator {
	v := validator.New()

	// Register custom validation tags here if needed
	// e.g., v.RegisterValidation("custom_tag", customValidationFunc)

	// Use JSON tag names in validation errors
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return &CustomValidator{
		validator: v,
		logger:    logger.With(zap.String("component", "validator")),
	}
}

// Validate validates a struct
func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		// Convert validation errors to user-friendly format
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			fieldsErrors := make(map[string]string)
			for _, e := range validationErrors {
				fieldsErrors[e.Field()] = formatValidationError(e)
			}

			// Log validation errors at debug level
			cv.logger.Debug("Validation failed", zap.Any("errors", fieldsErrors))

			return errors.NewValidationError("Validation failed", err)
		}

		return errors.NewValidationError("Validation failed", err)
	}
	return nil
}

// formatValidationError returns a user-friendly error message for a validation error
func formatValidationError(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "This field is required"
	case "min":
		return "Value is too short"
	case "max":
		return "Value is too long"
	case "email":
		return "Invalid email format"
	case "oneof":
		return "Value must be one of: " + e.Param()
	default:
		return "Invalid value"
	}
}

// ValidateRequest validates a request and binds it to the given struct
func ValidateRequest(c echo.Context, req interface{}) error {
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	if err := c.Validate(req); err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return echo.NewHTTPError(appErr.Status, appErr.Message)
		}
		return echo.NewHTTPError(http.StatusBadRequest, "Validation failed")
	}

	return nil
}
