// Package middleware — role.go
//
// RequireRole enforces role-based access control (RBAC) on protected routes.
//
// It must be applied AFTER the JWT() middleware, which validates the token and
// stores the parsed claims in Fiber's Locals under the key "user".
// The JWT middleware from gofiber/contrib/v3/jwt stores a *golang-jwt/jwt/v5.Token.
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
	golangJWT "github.com/golang-jwt/jwt/v5"

	jwtpkg "github.com/dsmes/dsmes-backend/internal/pkg/jwt"
)

// RequireRole returns a middleware that allows only the specified roles to proceed.
// Responds with 403 Forbidden if the authenticated user's role is not in the list.
// Must be chained after JWT() middleware.
func RequireRole(roles ...string) fiber.Handler {
	return func(c fiber.Ctx) error {
		// The gofiber/contrib/v3/jwt middleware stores the parsed token in Locals("user")
		// as a *github.com/golang-jwt/jwt/v5.Token.
		token, ok := c.Locals("user").(*golangJWT.Token)
		if !ok || token == nil {
			return fiber.ErrUnauthorized
		}

		claims, ok := token.Claims.(*jwtpkg.Claims)
		if !ok || claims == nil {
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

// ClaimsFromContext extracts the typed JWT claims from the Fiber context.
// Returns nil if the JWT middleware has not run or the token is invalid.
// Use this in handlers and downstream middleware to read the authenticated user's ID and role.
//
//	claims := middleware.ClaimsFromContext(c)
//	if claims == nil {
//	    return fiber.ErrUnauthorized
//	}
//	patientID := claims.UserID
func ClaimsFromContext(c fiber.Ctx) *jwtpkg.Claims {
	token, ok := c.Locals("user").(*golangJWT.Token)
	if !ok || token == nil {
		return nil
	}
	claims, ok := token.Claims.(*jwtpkg.Claims)
	if !ok {
		return nil
	}
	return claims
}
