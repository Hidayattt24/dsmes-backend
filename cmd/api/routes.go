// cmd/api/routes.go — Route registration.
//
// This file is the single place where ALL route paths are defined and mounted.
//
// Structure:
//   - /swagger/*         → Swagger UI (development / staging only)
//   - /api/health        → public liveness check (no auth required)
//   - /api/v1/auth/...   → authentication (public + protected)
//   - /api/v1/admin/...  → admin-only routes (JWT + RequireRole("admin"))
//   - /api/v1/puskesmas/ → puskesmas-only routes (JWT + RequireRole("puskesmas"))
//   - /api/v1/patient/   → patient-only routes (JWT + RequireRole("patient"))
package main

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/contrib/v3/swaggerui"

	"github.com/dsmes/dsmes-backend/internal/container"
	"github.com/dsmes/dsmes-backend/internal/middleware"
	"github.com/dsmes/dsmes-backend/internal/modules/auth"
	"github.com/dsmes/dsmes-backend/internal/modules/blood_sugar"
	"github.com/dsmes/dsmes-backend/internal/modules/checkin"
	"github.com/dsmes/dsmes-backend/internal/modules/education"
	"github.com/dsmes/dsmes-backend/internal/modules/nutrition"
	"github.com/dsmes/dsmes-backend/internal/modules/patient"
	"github.com/dsmes/dsmes-backend/internal/modules/reminder"
	"github.com/dsmes/dsmes-backend/internal/modules/routine"
	"github.com/dsmes/dsmes-backend/internal/modules/settings"
	"github.com/dsmes/dsmes-backend/internal/modules/staff"
	"github.com/dsmes/dsmes-backend/internal/modules/summary"
	"github.com/dsmes/dsmes-backend/internal/pkg/response"
)

// registerRoutes mounts all routes on the Fiber application.
func registerRoutes(app *fiber.App, c *container.Container) {
	// ── Swagger UI (development / staging only) ───────────────────────────────
	if c.Config.Swagger.Enabled {
		app.Use(swaggerui.New(swaggerui.Config{
			BasePath: "/",
			FilePath: "./docs/swagger.json",
			Path:     "swagger",
			Title:    "DSMES Backend API Documentation",
		}))
		c.Logger.Sugar().Infof("Swagger UI → http://%s/swagger/", c.Config.Swagger.Host)
	}

	// ── Health check ──────────────────────────────────────────────────────────
	app.Get("/api/health", func(ctx fiber.Ctx) error {
		return response.Success(ctx, "server is running", fiber.Map{
			"status":  "ok",
			"version": "1.0.0",
		})
	})

	// ── Instantiations ────────────────────────────────────────────────────────
	// We instantiate all repositories, services, and handlers here using GORM DB
	// and Logger injected from the dependency container.

	// 1. Staff
	staffRepo := staff.NewStaffRepository(c.DB, c.Logger)
	staffSvc := staff.NewStaffService(staffRepo, c.Logger)
	staffHandler := staff.NewStaffHandler(staffSvc, c.Logger)

	// 2. Patient
	patientRepo := patient.NewPatientRepository(c.DB, c.Logger)
	patientSvc := patient.NewPatientService(patientRepo, c.Logger)
	patientHandler := patient.NewPatientHandler(patientSvc, c.Logger)

	// 3. Routine
	routineRepo := routine.NewRoutineRepository(c.DB, c.Logger)
	routineSvc := routine.NewRoutineService(routineRepo, c.Logger)
	routineHandler := routine.NewRoutineHandler(routineSvc, c.Logger)

	// 4. Blood Sugar
	bsRepo := blood_sugar.NewBloodSugarRepository(c.DB, c.Logger)
	bsSvc := blood_sugar.NewBloodSugarService(bsRepo, c.Logger)
	bsHandler := blood_sugar.NewBloodSugarHandler(bsSvc, c.Logger)

	// 5. Checkin
	checkinRepo := checkin.NewCheckinRepository(c.DB, c.Logger)
	checkinSvc := checkin.NewCheckinService(checkinRepo, c.Logger)
	checkinHandler := checkin.NewCheckinHandler(checkinSvc, c.Logger)

	// 6. Nutrition
	nutritionRepo := nutrition.NewNutritionRepository(c.DB, c.Logger)
	nutritionSvc := nutrition.NewNutritionService(nutritionRepo, c.Logger)
	nutritionHandler := nutrition.NewNutritionHandler(nutritionSvc, c.Logger)

	// 7. Reminder & Notification
	reminderRepo := reminder.NewReminderRepository(c.DB, c.Logger)
	reminderSvc := reminder.NewReminderService(reminderRepo, c.Logger)
	reminderHandler := reminder.NewReminderHandler(reminderSvc, c.Logger)

	// 8. Education
	eduRepo := education.NewEducationRepository(c.DB, c.Logger)
	eduSvc := education.NewEducationService(eduRepo, c.Logger)
	eduHandler := education.NewEducationHandler(eduSvc, c.Logger)

	// 9. Summary
	summaryRepo := summary.NewSummaryRepository(c.DB, c.Logger)
	summarySvc := summary.NewSummaryService(summaryRepo, c.Logger)
	summaryHandler := summary.NewSummaryHandler(summarySvc, c.Logger)

	// 10. Settings & Support
	settingsRepo := settings.NewSettingsRepository(c.DB, c.Logger)
	settingsSvc := settings.NewSettingsService(settingsRepo, c.Logger)
	settingsHandler := settings.NewSettingsHandler(settingsSvc, c.Logger)

	// ── API v1 ────────────────────────────────────────────────────────────────
	v1 := app.Group("/api/v1")

	// ── Public routes (no JWT) ────────────────────────────────────────────────
	auth.RegisterRoutes(v1, c)
	v1.Post("/auth/register", patientHandler.Register)

	// ── Protected: Admin Group (JWT + RequireRole("admin")) ───────────────────
	admin := v1.Group("/admin",
		middleware.JWT(c.Config),
		middleware.RequireRole("admin"),
	)
	{
		// Staff Management
		admin.Get("/staff", staffHandler.List)
		admin.Post("/staff", staffHandler.Create)
		admin.Get("/staff/:id", staffHandler.GetByID)
		admin.Put("/staff/:id", staffHandler.Update)
		admin.Patch("/staff/:id/status", staffHandler.ToggleStatus)
		admin.Delete("/staff/:id", staffHandler.Delete)

		admin.Get("/me", staffHandler.GetMe)
		admin.Put("/me", staffHandler.UpdateMe)

		// Patient Management
		admin.Get("/patients", patientHandler.List)
		admin.Get("/patients/:id", patientHandler.GetByID)
		admin.Patch("/patients/:id/status", patientHandler.ToggleStatus)
		admin.Patch("/patients/:id/assign", patientHandler.AssignPuskesmas)
		admin.Delete("/patients/:id", patientHandler.Delete)

		// Blood Sugar Logs view
		admin.Get("/patients/:id/blood-sugar", bsHandler.GetPatientHistory)

		// Global Foods Management
		admin.Post("/foods", nutritionHandler.CreateFood)
		admin.Put("/foods/:id", nutritionHandler.UpdateFood)

		// Education content CRUD
		admin.Get("/education/articles", eduHandler.ListAdmin)
		admin.Post("/education/articles", eduHandler.Create)
		admin.Put("/education/articles/:id", eduHandler.Update)
		admin.Patch("/education/articles/:id/publish", eduHandler.Publish)

		// Support Tickets Management
		admin.Get("/support/tickets", settingsHandler.GetAllTickets)
		admin.Patch("/support/tickets/:id/resolve", settingsHandler.ResolveTicket)
	}

	// ── Protected: Puskesmas Group (JWT + RequireRole("puskesmas")) ───────────
	puskesmas := v1.Group("/puskesmas",
		middleware.JWT(c.Config),
		middleware.RequireRole("puskesmas"),
	)
	{
		puskesmas.Get("/me", staffHandler.GetMe)
		puskesmas.Put("/me", staffHandler.UpdateMe)

		// Patient Monitoring
		puskesmas.Get("/patients", patientHandler.ListPuskesmas)
		puskesmas.Get("/patients/:id", patientHandler.GetByID)
		puskesmas.Get("/patients/:id/blood-sugar", bsHandler.GetPatientHistory)

		// Dashboard statistics
		puskesmas.Get("/dashboard/blood-sugar", bsHandler.GetDashboard)

		// Education articles list
		puskesmas.Get("/education/articles", eduHandler.ListPublished)
		puskesmas.Get("/education/articles/:id", eduHandler.GetByID)
	}

	// ── Protected: Patient Group (JWT + RequireRole("patient")) ───────────────
	patientGroup := v1.Group("/patient",
		middleware.JWT(c.Config),
		middleware.RequireRole("patient"),
	)
	{
		patientGroup.Get("/me", patientHandler.GetMe)
		patientGroup.Put("/me", patientHandler.UpdateMe)

		// Routines habit logging
		patientGroup.Get("/routines", routineHandler.List)
		patientGroup.Put("/routines/:routineTimeId", routineHandler.Configure)
		patientGroup.Post("/routines/log", routineHandler.Log)
		patientGroup.Get("/routines/status", routineHandler.Status)

		// Blood sugar records
		patientGroup.Post("/blood-sugar", bsHandler.Log)
		patientGroup.Get("/blood-sugar", bsHandler.GetHistory)

		// Daily Checkin Calendar
		patientGroup.Post("/checkin", checkinHandler.Checkin)
		patientGroup.Get("/checkin/calendar", checkinHandler.GetCalendar)

		// Nutrition meal logs
		patientGroup.Get("/foods/recent", nutritionHandler.GetRecent)
		patientGroup.Post("/meals", nutritionHandler.LogMeal)
		patientGroup.Get("/meals/summary", nutritionHandler.GetSummary)

		// Personal reminders
		patientGroup.Get("/reminders", reminderHandler.List)
		patientGroup.Post("/reminders", reminderHandler.Create)
		patientGroup.Put("/reminders/:id", reminderHandler.Update)
		patientGroup.Patch("/reminders/:id/toggle", reminderHandler.Toggle)
		patientGroup.Delete("/reminders/:id", reminderHandler.Delete)

		// Notifications inbox
		patientGroup.Get("/notifications", reminderHandler.GetNotifications)
		patientGroup.Patch("/notifications/read", reminderHandler.MarkRead)

		// Education bookmarks & completions
		patientGroup.Post("/education/:id/complete", eduHandler.Complete)
		patientGroup.Post("/education/:id/save", eduHandler.Save)
		patientGroup.Get("/education/saved", eduHandler.ListSaved)

		// Weekly analytical summary cards
		patientGroup.Get("/summary/weekly", summaryHandler.GetLatest)

		// Support Ticket submission
		patientGroup.Post("/support/tickets", settingsHandler.SubmitTicket)
		patientGroup.Get("/support/tickets", settingsHandler.GetMyTickets)
	}

	// ── Authenticated Shared Group (any authenticated role) ───────────────────
	shared := v1.Group("", middleware.JWT(c.Config))
	{
		// Food database search
		shared.Get("/foods", nutritionHandler.Search)

		// Helpdesk FAQs
		shared.Get("/faqs", settingsHandler.GetFAQs)

		// General education articles view
		shared.Get("/education/categories", eduHandler.ListCategories)
		shared.Get("/education/articles", eduHandler.ListPublished)
		shared.Get("/education/articles/:id", eduHandler.GetByID)
	}

	// ── Internal / Cron Group ─────────────────────────────────────────────────
	internal := v1.Group("/internal")
	{
		internal.Post("/summary/generate", summaryHandler.Generate)
	}

	// ── 404 fallback ──────────────────────────────────────────────────────────
	app.Use(func(ctx fiber.Ctx) error {
		return response.Error(ctx, fiber.StatusNotFound, "route not found")
	})
}
