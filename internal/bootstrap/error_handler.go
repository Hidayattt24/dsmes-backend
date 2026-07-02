// Package bootstrap — error_handler.go
//
// errorHandler is the centralised Fiber ErrorHandler registered on the app.
// Any handler that returns an error (including panics caught by the Recover
// middleware) ends up here.
//
// Resolution order:
//  1. *errs.AppError   → use its Code and Message fields directly.
//  2. *fiber.Error     → use Fiber's built-in error type (e.g. 404 from router).
//  3. Anything else    → 500 Internal Server Error (internal detail is NOT exposed).
package bootstrap

import (
	"errors"
	"net/http"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"

	"github.com/dsmes/dsmes-backend/internal/pkg/errs"
)

// errorEnvelope matches the shape used by internal/pkg/response for consistency.
type errorEnvelope struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Errors  any    `json:"errors,omitempty"`
}

// errorHandler returns a Fiber-compatible ErrorHandler function.
// The logger is captured in the closure to log unexpected internal errors.
func errorHandler(log *zap.Logger) fiber.ErrorHandler {
	return func(c fiber.Ctx, err error) error {
		code := http.StatusInternalServerError
		message := "internal server error"

		var appErr *errs.AppError
		var fiberErr *fiber.Error

		switch {
		case errors.As(err, &appErr):
			// Structured application error — use its HTTP code and message.
			code = appErr.HTTPCode()
			message = appErr.Message
			if appErr.Err != nil {
				log.Warn("application error",
					zap.Int("status", code),
					zap.String("message", message),
					zap.Error(appErr.Err),
				)
			}

		case errors.As(err, &fiberErr):
			// Fiber built-in error (e.g. 404 route not found, 405 method not allowed).
			code = fiberErr.Code
			message = fiberErr.Message

		default:
			// Unexpected error — log it with full detail but hide internals from caller.
			log.Error("unhandled error",
				zap.Error(err),
				zap.String("path", c.Path()),
				zap.String("method", c.Method()),
			)
		}

		return c.Status(code).JSON(errorEnvelope{
			Success: false,
			Message: message,
		})
	}
}
