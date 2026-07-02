package _template

import (
	"github.com/gofiber/fiber/v3"

	"github.com/dsmes/dsmes-backend/internal/container"
)

// RegisterRoutes mounts all Template routes on the provided router group.
//
// Call this from cmd/api/routes.go, passing the API version group:
//
//	v1 := app.Group("/api/v1")
//	_template.RegisterRoutes(v1, c)
//
// For protected routes, apply the JWT middleware before passing the group:
//
//	protected := v1.Use(middleware.JWT(c.Config))
//	_template.RegisterRoutes(protected, c)
func RegisterRoutes(router fiber.Router, c *container.Container) {
	// Wire dependencies via constructor injection.
	// Order: Repository ← Service ← Handler
	repo := NewTemplateRepository(c.DB, c.Logger)
	svc := NewTemplateService(repo, c.Logger)
	h := NewTemplateHandler(svc, c.Logger)

	// Mount routes on the provided group.
	// The group prefix (e.g. /api/v1) is set by the caller.
	g := router.Group("/templates")
	g.Get("/", h.List)
	g.Get("/:id", h.GetByID)
	g.Post("/", h.Create)
	g.Put("/:id", h.Update)
	g.Delete("/:id", h.Delete)
}
