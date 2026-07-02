// Package bootstrap creates and configures the Fiber application instance.
//
// Fiber v3 is used (not v2). Key differences from v2:
//   - fiber.Ctx is a value type (interface), not *fiber.Ctx pointer.
//   - Handler signature: func(fiber.Ctx) error
//   - Middleware imports use /v3 suffix.
//
// Architectural decision: the Fiber app is constructed here with all global
// middleware registered in order. Route groups and business handlers are
// registered separately by the router layer, which is wired in main.go.
package bootstrap

import (
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"

	"github.com/dsmes/dsmes-backend/internal/config"
	"github.com/dsmes/dsmes-backend/internal/middleware"
)

// NewFiberApp creates a Fiber application with global middleware configured.
// Middleware is registered in this exact order (execution order matters):
//  1. Recover  — catches panics before anything else runs
//  2. Logger   — logs every request (after recover so panics are logged too)
//  3. CORS     — sets response headers before any handler runs
//  4. Limiter  — rate limits before business logic
func NewFiberApp(cfg *config.Config, log *zap.Logger) *fiber.App {
	app := fiber.New(fiber.Config{
		// AppName appears in the Server header and in Swagger info.
		AppName: cfg.App.Name,

		// StrictRouting: "/foo" and "/foo/" are treated as different routes.
		StrictRouting: true,

		// CaseSensitive: "/Foo" and "/foo" are different routes.
		CaseSensitive: true,

		// ErrorHandler: centralised JSON error response for all returned errors.
		// Uses the AppError type from internal/pkg/errs.
		ErrorHandler: errorHandler(log),
	})

	// ── Global middleware (order matters) ─────────────────────────────────────
	app.Use(middleware.Recover())
	app.Use(middleware.RequestLogger(log))
	app.Use(middleware.CORS(cfg))
	app.Use(middleware.RateLimiter())

	return app
}
