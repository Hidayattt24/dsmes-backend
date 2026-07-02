package staff

import (
	"github.com/gofiber/fiber/v3"

	"github.com/dsmes/dsmes-backend/internal/container"
)

// RegisterRoutes registers endpoints.
// In cmd/api/routes.go:
//
//	// adminGroup is protected with JWT + RequireRole("admin")
//	staff.RegisterRoutes(adminGroup, c)
func RegisterRoutes(router fiber.Router, c *container.Container) {
	repo := NewStaffRepository(c.DB, c.Logger)
	svc := NewStaffService(repo, c.Logger)
	h := NewStaffHandler(svc, c.Logger)

	// Since the caller already applies JWT and RequireRole("admin"),
	// these routes inherit that protection automatically.
	router.Get("/staff", h.List)
	router.Post("/staff", h.Create)
	router.Get("/staff/:id", h.GetByID)
	router.Put("/staff/:id", h.Update)
	router.Patch("/staff/:id/status", h.ToggleStatus)
	router.Delete("/staff/:id", h.Delete)

	router.Get("/me", h.GetMe)
	router.Put("/me", h.UpdateMe)
}
