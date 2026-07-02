// Package middleware — jwt.go
//
// JWT returns a Fiber middleware that validates a Bearer token from the
// Authorization header on every protected route it is applied to.
//
// This middleware is NOT registered globally. It is applied selectively at the
// route group level (e.g. only on /api/v1/admin/..., /api/v1/user/...).
// Public routes (login, register, health check) bypass it entirely.
//
// After successful validation the parsed *jwt.Token is available in context:
//
//	token := c.Locals("user").(*jwt.Token)
//	claims := token.Claims.(*jwtpkg.Claims)
//
// Role-based access control (RBAC) is enforced by a separate RequireRole()
// middleware (to be implemented when business modules are added).
//
// Design note: we intentionally do NOT pass a custom Extractor field.
// gofiber/contrib/v3/jwt extracts the Bearer token from the Authorization
// header by default. Only override this if using cookies or query params.
package middleware

import (
	jwtware "github.com/gofiber/contrib/v3/jwt"
	"github.com/gofiber/fiber/v3"

	"github.com/dsmes/dsmes-backend/internal/config"
	jwtpkg "github.com/dsmes/dsmes-backend/internal/pkg/jwt"
)

// JWT returns the JWT validation middleware configured with HS256.
// Apply this to route groups that require authentication:
//
//	api := app.Group("/api/v1")
//	api.Use(middleware.JWT(cfg))
func JWT(cfg *config.Config) fiber.Handler {
	return jwtware.New(jwtware.Config{
		// SigningKey — HS256 shared secret loaded from JWT_SECRET in .env.
		// For RS256, replace Key with the RSA public key.
		SigningKey: jwtware.SigningKey{Key: []byte(cfg.JWT.Secret)},

		// Claims — use our typed Claims struct (not jwt.MapClaims) for type safety.
		// The library reads the "Authorization: Bearer <token>" header by default.
		Claims: &jwtpkg.Claims{},

		// ErrorHandler — returns a structured JSON 401 on invalid/expired tokens.
		ErrorHandler: func(c fiber.Ctx, err error) error {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "unauthorized — invalid or expired token",
			})
		},
	})
}
