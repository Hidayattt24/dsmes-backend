package survey

import (
	"github.com/gofiber/fiber/v3"
)

func RegisterRoutes(admin fiber.Router, staff fiber.Router, patient fiber.Router, h *SurveyHandler) {
	// Admin routes (Full CRUD + Analytics + Export)
	admin.Get("/surveys", h.List)
	admin.Post("/surveys", h.Create)
	admin.Get("/surveys/:id", h.GetByID)
	admin.Put("/surveys/:id", h.Update)
	admin.Delete("/surveys/:id", h.Delete)
	admin.Patch("/surveys/:id/status", h.UpdateStatus)
	admin.Post("/surveys/:id/duplicate", h.Duplicate)
	admin.Get("/surveys/:id/responses", h.GetResponses)
	admin.Get("/surveys/:id/analytics", h.GetAnalytics)
	admin.Get("/surveys/:id/export", h.ExportCSV)

	// Staff routes (Read-only + Analytics + Export)
	staff.Get("/surveys", h.List)
	staff.Get("/surveys/:id", h.GetByID)
	staff.Get("/surveys/:id/responses", h.GetResponses)
	staff.Get("/surveys/:id/analytics", h.GetAnalytics)
	staff.Get("/surveys/:id/export", h.ExportCSV)

	// Patient routes (Active Survey + Submit)
	patient.Get("/surveys/active", h.GetActivePatientSurvey)
	patient.Get("/surveys/:id", h.GetByID)
	patient.Post("/surveys/:id/submit", h.Submit)
}
