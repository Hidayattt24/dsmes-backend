// Package middleware — jwt.go
//
// JWT returns a Fiber middleware that validates a Bearer token from the
// Authorization header on every protected route it is applied to.
//
// This middleware is NOT registered globally. It is applied selectively at the
// route group level (e.g. only on /api/v1/admin/..., /api/v1/user/...)
// Public routes (login, register, health check) bypass it entirely.
//
// After successful validation the parsed *jwtpkg.Claims is stored directly in context:
//
//	claims := c.Locals("claims").(*jwtpkg.Claims)
//
// Role-based access control (RBAC) is enforced by a separate RequireRole()
// middleware.
package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"

	"github.com/dsmes/dsmes-backend/internal/config"
	jwtpkg "github.com/dsmes/dsmes-backend/internal/pkg/jwt"
)

// JWT returns the JWT validation middleware configured with HS256.
// Apply this to route groups that require authentication.
func JWT(cfg *config.Config) fiber.Handler {
	secret := []byte(cfg.JWT.Secret)
	issuer := cfg.JWT.Issuer

	return func(c fiber.Ctx) error {
		// Extract Bearer token from Authorization header
		authHeader := c.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "unauthorized — missing or malformed Authorization header",
			})
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Parse and validate the token
		claims := &jwtpkg.Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.ErrUnauthorized
			}
			return secret, nil
		}, jwt.WithIssuer(issuer))

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "unauthorized — invalid or expired token",
			})
		}

		// Store the claims directly — no intermediate *jwt.Token wrapper
		c.Locals("claims", claims)
		return c.Next()
	}
}
