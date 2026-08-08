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
// Stack traces are only logged in non-production environments; in production
// the error handler returns a generic 500 JSON response without internals.
func Recover(isProduction bool) fiber.Handler {
	return recover.New(recover.Config{
		EnableStackTrace: !isProduction,
	})
}
