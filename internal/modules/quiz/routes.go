package quiz

import (
	"github.com/gofiber/fiber/v3"

	"github.com/dsmes/dsmes-backend/internal/container"
)

// RegisterRoutes is a placeholder that implements the router module structure.
// Actual wiring is done explicitly in cmd/api/routes.go to cleanly separate
// public, admin, staff, and user route groups.
func RegisterRoutes(router fiber.Router, c *container.Container) {
	// Instantiation is handled in cmd/api/routes.go
}
