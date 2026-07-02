// Package errs defines the error types used across the application.
//
// Architectural decision: all application errors are wrapped in AppError,
// which carries an HTTP status code alongside the message. This allows the
// centralised Fiber error handler in bootstrap/app.go to render a consistent
// JSON response without each handler needing to know about HTTP semantics.
//
// Usage in a service:
//
//	if user == nil {
//	    return nil, errs.NewNotFound("user not found")
//	}
//
// Usage in a repository:
//
//	if errors.Is(err, gorm.ErrRecordNotFound) {
//	    return nil, errs.NewNotFound("record not found")
//	}
package errs

import (
	"errors"
	"fmt"
	"net/http"
)

// AppError is a structured application error that carries an HTTP status code.
// It satisfies the standard error interface.
type AppError struct {
	Code    int    // HTTP status code
	Message string // Human-readable message (exposed to API callers)
	Err     error  // Underlying cause (internal, NOT exposed to callers)
}

// Error implements the error interface.
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// Unwrap allows errors.Is / errors.As to inspect the wrapped error.
func (e *AppError) Unwrap() error {
	return e.Err
}

// HTTPCode returns the HTTP status code associated with the error.
func (e *AppError) HTTPCode() int {
	return e.Code
}

// ── Constructors ──────────────────────────────────────────────────────────────

// New creates a generic AppError with a given HTTP status code.
func New(code int, message string, cause ...error) *AppError {
	var err error
	if len(cause) > 0 {
		err = cause[0]
	}
	return &AppError{Code: code, Message: message, Err: err}
}

// NewBadRequest creates a 400 Bad Request error.
func NewBadRequest(message string, cause ...error) *AppError {
	return New(http.StatusBadRequest, message, cause...)
}

// NewUnauthorized creates a 401 Unauthorized error.
func NewUnauthorized(message string, cause ...error) *AppError {
	return New(http.StatusUnauthorized, message, cause...)
}

// NewForbidden creates a 403 Forbidden error.
func NewForbidden(message string, cause ...error) *AppError {
	return New(http.StatusForbidden, message, cause...)
}

// NewNotFound creates a 404 Not Found error.
func NewNotFound(message string, cause ...error) *AppError {
	return New(http.StatusNotFound, message, cause...)
}

// NewConflict creates a 409 Conflict error.
func NewConflict(message string, cause ...error) *AppError {
	return New(http.StatusConflict, message, cause...)
}

// NewUnprocessable creates a 422 Unprocessable Entity error.
func NewUnprocessable(message string, cause ...error) *AppError {
	return New(http.StatusUnprocessableEntity, message, cause...)
}

// NewInternal creates a 500 Internal Server Error.
func NewInternal(message string, cause ...error) *AppError {
	return New(http.StatusInternalServerError, message, cause...)
}

// NewTooManyRequests creates a 429 Too Many Requests error.
func NewTooManyRequests(message string, cause ...error) *AppError {
	return New(http.StatusTooManyRequests, message, cause...)
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// IsNotFound returns true if err is a 404 AppError.
func IsNotFound(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code == http.StatusNotFound
	}
	return false
}

// IsConflict returns true if err is a 409 AppError.
func IsConflict(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code == http.StatusConflict
	}
	return false
}
