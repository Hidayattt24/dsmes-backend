// Package middleware — role.go
//
// RequireRole enforces role-based access control (RBAC) on protected routes.
//
// It must be applied AFTER the JWT() middleware, which validates the token and
// stores the parsed claims in Fiber's Locals under the key "claims".
//
// # Usage
//
//	admin := v1.Group("/admin")
//	admin.Use(middleware.JWT(cfg), middleware.RequireRole("admin"))
//
//	staff := v1.Group("/staff")
//	staff.Use(middleware.JWT(cfg), middleware.RequireRole("staff"))
//
//	// Multi-role: accessible by admin AND staff
//	shared := v1.Group("/shared")
//	shared.Use(middleware.JWT(cfg), middleware.RequireRoles("admin", "staff"))
package middleware

import (
	"github.com/gofiber/fiber/v3"

	jwtpkg "github.com/dsmes/dsmes-backend/internal/pkg/jwt"
)

// RequireRole returns a middleware that allows only the specified roles to proceed.
// Responds with 403 Forbidden if the authenticated user's role is not in the list.
// Must be chained after JWT() middleware.
func RequireRole(roles ...string) fiber.Handler {
	return func(c fiber.Ctx) error {
		claims := ClaimsFromContext(c)
		if claims == nil {
			return fiber.ErrUnauthorized
		}

		for _, role := range roles {
			if claims.Role == role {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "access denied — insufficient permissions",
		})
	}
}

// RequireRoles is an alias for RequireRole to improve readability when specifying multiple roles.
func RequireRoles(roles ...string) fiber.Handler {
	return RequireRole(roles...)
}

// ClaimsFromContext extracts the typed JWT claims stored by the JWT() middleware.
// Returns nil if the JWT middleware has not run or the token is invalid.
func ClaimsFromContext(c fiber.Ctx) *jwtpkg.Claims {
	claims, ok := c.Locals("claims").(*jwtpkg.Claims)
	if !ok || claims == nil {
		return nil
	}
	return claims
}
