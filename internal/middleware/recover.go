// Package middleware — recover.go
//
// Recover catches panics anywhere in a subsequent handler and converts them
// into a structured JSON error response via the central ErrorHandler.
// This prevents the server from crashing on unexpected panics.
//
// Placement: must be the FIRST middleware registered on the app so that
// panics in all subsequent middleware and handlers are captured.
package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/recover"
)

// Recover returns the configured panic recovery middleware.
// EnableStackTrace is disabled in production to avoid leaking internals.
func Recover() fiber.Handler {
	return recover.New(recover.Config{
		// EnableStackTrace: logging only — the stack trace is NOT sent to the client.
		// In production the error handler will return a generic 500 JSON response.
		EnableStackTrace: true,
	})
}
