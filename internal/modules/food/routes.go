package food

import (
	"github.com/gofiber/fiber/v3"
)

func RegisterRoutes(adminGroup fiber.Router, sharedAuthGroup fiber.Router, patientGroup fiber.Router, handler *FoodHandler) {
	// Shared / Public food routes (Authenticated user/staff/admin)
	if sharedAuthGroup != nil {
		sharedAuthGroup.Get("/foods/search", handler.SearchFoods)
		sharedAuthGroup.Get("/foods/stats", handler.GetStats)
		sharedAuthGroup.Get("/foods", handler.GetFoods)
		sharedAuthGroup.Get("/foods/:id", handler.GetByID)
	}

	// Patient food requests (User role)
	if patientGroup != nil {
		patientGroup.Post("/food-requests", handler.CreateFoodRequest)
	}

	// Admin-only CRUD, Import, Export, Stats & Food Requests Review
	if adminGroup != nil {
		adminFoods := adminGroup.Group("/foods")

		// 1. Static subpaths MUST come before parameterized :id
		adminFoods.Get("/stats", handler.GetStats)
		adminFoods.Get("/export", handler.Export)
		adminFoods.Post("/import/preview", handler.PreviewImport)
		adminFoods.Post("/import/confirm", handler.ConfirmImport)

		// 2. Collection root (support both "" and "/")
		adminFoods.Get("", handler.GetFoods)
		adminFoods.Get("/", handler.GetFoods)
		adminFoods.Post("", handler.Create)
		adminFoods.Post("/", handler.Create)

		// 3. Item parameter routes
		adminFoods.Get("/:id", handler.GetByID)
		adminFoods.Put("/:id", handler.Update)
		adminFoods.Delete("/:id", handler.Delete)

		// 4. Food requests management
		adminFoodRequests := adminGroup.Group("/food-requests")
		adminFoodRequests.Get("/stats", handler.GetFoodRequestStats)
		adminFoodRequests.Get("", handler.GetFoodRequests)
		adminFoodRequests.Get("/", handler.GetFoodRequests)
		adminFoodRequests.Post("/:id/approve", handler.ApproveFoodRequest)
		adminFoodRequests.Post("/:id/reject", handler.RejectFoodRequest)
	}
}
