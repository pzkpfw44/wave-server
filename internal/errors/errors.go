package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// Common error codes
const (
	ErrCodeUnauthenticated = "UNAUTHENTICATED"
	ErrCodeUnauthorized    = "UNAUTHORIZED"
	ErrCodeNotFound        = "NOT_FOUND"
	ErrCodeBadRequest      = "BAD_REQUEST"
	ErrCodeConflict        = "CONFLICT"
	ErrCodeInternal        = "INTERNAL"
	ErrCodeValidation      = "VALIDATION"
)

// AppError represents an application-specific error
type AppError struct {
	Code    string // Error code
	Message string // User-facing message
	Err     error  // Original error (not exposed to users)
	Status  int    // HTTP status code
}

// Error returns the error message
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the wrapped error
func (e *AppError) Unwrap() error {
	return e.Err
}

// NewUnauthenticatedError creates a new unauthenticated error
func NewUnauthenticatedError(msg string) *AppError {
	return &AppError{
		Code:    ErrCodeUnauthenticated,
		Message: msg,
		Status:  http.StatusUnauthorized,
	}
}

// NewUnauthorizedError creates a new unauthorized error
func NewUnauthorizedError(msg string) *AppError {
	return &AppError{
		Code:    ErrCodeUnauthorized,
		Message: msg,
		Status:  http.StatusForbidden,
	}
}

// NewNotFoundError creates a new not found error
func NewNotFoundError(resource string) *AppError {
	return &AppError{
		Code:    ErrCodeNotFound,
		Message: fmt.Sprintf("%s not found", resource),
		Status:  http.StatusNotFound,
	}
}

// NewBadRequestError creates a new bad request error
func NewBadRequestError(msg string, err error) *AppError {
	return &AppError{
		Code:    ErrCodeBadRequest,
		Message: msg,
		Err:     err,
		Status:  http.StatusBadRequest,
	}
}

// NewConflictError creates a new conflict error
func NewConflictError(msg string) *AppError {
	return &AppError{
		Code:    ErrCodeConflict,
		Message: msg,
		Status:  http.StatusConflict,
	}
}

// NewInternalError creates a new internal error
func NewInternalError(msg string, err error) *AppError {
	return &AppError{
		Code:    ErrCodeInternal,
		Message: msg,
		Err:     err,
		Status:  http.StatusInternalServerError,
	}
}

// NewValidationError creates a new validation error
func NewValidationError(msg string, err error) *AppError {
	return &AppError{
		Code:    ErrCodeValidation,
		Message: msg,
		Err:     err,
		Status:  http.StatusUnprocessableEntity,
	}
}

// IsAppError checks if an error is an AppError
func IsAppError(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}
