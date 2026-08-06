// cmd/api/routes.go — Route registration.
//
// This file is the single place where ALL route paths are defined and mounted.
//
// Structure:
//   - /swagger/*         → Swagger UI (development / staging only)
//   - /api/health        → public liveness check (no auth required)
//   - /api/v1/auth/...   → authentication (public + protected)
//   - /api/v1/admin/...  → admin-only routes (JWT + RequireRole("admin"))
//   - /api/v1/staff/     → staff-only monitoring routes (JWT + RequireRole("staff"))
//   - /api/v1/patient/   → patient-only routes (JWT + RequireRole("user"))
package main

import (
	"github.com/gofiber/contrib/v3/swaggerui"
	"github.com/gofiber/fiber/v3"

	"github.com/dsmes/dsmes-backend/internal/container"
	"github.com/dsmes/dsmes-backend/internal/middleware"
	"github.com/dsmes/dsmes-backend/internal/modules/ai_chat"
	"github.com/dsmes/dsmes-backend/internal/modules/auth"
	"github.com/dsmes/dsmes-backend/internal/modules/blood_sugar"
	"github.com/dsmes/dsmes-backend/internal/modules/checkin"
	"github.com/dsmes/dsmes-backend/internal/modules/dashboard"
	"github.com/dsmes/dsmes-backend/internal/modules/education"
	"github.com/dsmes/dsmes-backend/internal/modules/education_progress"
	"github.com/dsmes/dsmes-backend/internal/modules/food"
	"github.com/dsmes/dsmes-backend/internal/modules/history"
	"github.com/dsmes/dsmes-backend/internal/modules/nutrition"
	"github.com/dsmes/dsmes-backend/internal/modules/patient"
	"github.com/dsmes/dsmes-backend/internal/modules/quiz"
	"github.com/dsmes/dsmes-backend/internal/modules/reminder"
	"github.com/dsmes/dsmes-backend/internal/modules/routine"
	"github.com/dsmes/dsmes-backend/internal/modules/settings"
	"github.com/dsmes/dsmes-backend/internal/modules/staff"
	"github.com/dsmes/dsmes-backend/internal/modules/summary"
	"github.com/dsmes/dsmes-backend/internal/modules/survey"
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
			CacheAge: -1, // Disable caching so spec changes are reflected immediately
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
	authRepo := auth.NewAuthRepository(c.DB, c.Logger)
	patientRepo := patient.NewPatientRepository(c.DB, c.Logger)
	patientSvc := patient.NewPatientService(patientRepo, authRepo, c.JWT, c.Email, c.Logger)
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

	// 9. Education Progress
	eduProgressRepo := education_progress.NewEducationProgressRepository(c.DB, c.Logger)
	eduProgressSvc := education_progress.NewEducationProgressService(eduProgressRepo, eduRepo, c.Logger)
	eduProgressHandler := education_progress.NewEducationProgressHandler(eduProgressSvc, c.Logger)

	// 10. Summary
	summaryRepo := summary.NewSummaryRepository(c.DB, c.Logger)
	summarySvc := summary.NewSummaryService(summaryRepo, c.Logger)
	summaryHandler := summary.NewSummaryHandler(summarySvc, c.Logger)

	// 10. Settings & Support
	settingsRepo := settings.NewSettingsRepository(c.DB, c.Logger)
	settingsSvc := settings.NewSettingsService(settingsRepo, c.Logger)
	settingsHandler := settings.NewSettingsHandler(settingsSvc, c.Logger)

	// 11. Quiz / Questionnaire
	quizRepo := quiz.NewQuizRepository(c.DB, c.Logger)
	quizSvc := quiz.NewQuizService(quizRepo, c.Logger)
	quizHandler := quiz.NewQuizHandler(quizSvc, c.Logger)

	// 12. History
	historyRepo := history.NewHistoryRepository(c.DB, c.Logger)
	historySvc := history.NewHistoryService(historyRepo, c.Logger)
	historyHandler := history.NewHistoryHandler(historySvc, c.Logger)

	// 13. Dashboard
	dashboardRepo := dashboard.NewDashboardRepository(c.DB, c.Logger)
	dashboardSvc := dashboard.NewDashboardService(dashboardRepo, c.Logger)
	dashboardHandler := dashboard.NewDashboardHandler(dashboardSvc, c.Logger)

	// 14. AI Personal Diabetes Assistant
	aiChatRepo := ai_chat.NewAIChatRepository(c.DB, c.Logger)
	aiChatSvc := ai_chat.NewAIChatService(aiChatRepo, c.Config.AI, c.Logger)
	aiChatHandler := ai_chat.NewAIChatHandler(aiChatSvc, c.Logger)

	// 15. Survey Module
	surveyRepo := survey.NewSurveyRepository(c.DB)
	surveySvc := survey.NewSurveyService(surveyRepo, c.Logger)
	surveyHandler := survey.NewSurveyHandler(surveySvc, c.Logger)

	// 16. Food Master Module
	foodRepo := food.NewFoodRepository(c.DB, c.Logger)
	foodSvc := food.NewFoodService(foodRepo, c.Logger)
	foodHandler := food.NewFoodHandler(foodSvc, c.Logger)

	// ── API v1 ────────────────────────────────────────────────────────────────
	v1 := app.Group("/api/v1")

	// ── Public routes (no JWT) ────────────────────────────────────────────────
	auth.RegisterRoutes(v1, c)
	v1.Post("/auth/register", patientHandler.Register)
	v1.Post("/nutrition/calculate-calories", nutritionHandler.CalculateCalories)

	// ── Protected: Admin Group (JWT + RequireRole("admin")) ───────────────────
	admin := v1.Group("/admin",
		middleware.JWT(c.Config),
		middleware.RequireRole("admin"),
	)
	{
		// Dashboard Statistics
		admin.Get("/dashboard", dashboardHandler.GetAdmin)
		admin.Get("/dashboard/stats", dashboardHandler.GetAdmin)
		admin.Get("/dashboard/top-articles", dashboardHandler.GetTopArticles)
		admin.Get("/dashboard/activity-chart", dashboardHandler.GetActivityChart)

		// Staff Management
		admin.Get("/staff", staffHandler.List)
		admin.Post("/staff", staffHandler.Create)
		admin.Get("/staff/:id", staffHandler.GetByID)
		admin.Put("/staff/:id", staffHandler.Update)
		admin.Patch("/staff/:id/status", staffHandler.ToggleStatus)
		admin.Delete("/staff/:id", staffHandler.Delete)

		admin.Get("/me", staffHandler.GetMe)
		admin.Put("/me", staffHandler.UpdateMe)
		admin.Put("/me/password", staffHandler.ChangePassword)

		// Patient Management
		admin.Get("/patients/stats", patientHandler.GetStats)
		admin.Get("/patients", patientHandler.List)
		admin.Get("/patients/:id", patientHandler.GetByID)
		admin.Put("/patients/:id", patientHandler.UpdatePatientByAdmin)
		admin.Patch("/patients/:id/status", patientHandler.ToggleStatus)
		admin.Patch("/patients/:id/assign", patientHandler.AssignStaff)
		admin.Delete("/patients/:id", patientHandler.Delete)

		// Patient Health Measurements & Logs views
		admin.Get("/patients/:id/measurements", patientHandler.GetPatientMeasurements)
		admin.Post("/patients/:id/measurements", patientHandler.CreateMeasurement)
		admin.Put("/patients/:id/measurements/:measurementId", patientHandler.UpdateMeasurement)
		admin.Get("/patients/:id/blood-sugar", bsHandler.GetPatientHistory)
		admin.Get("/patients/:id/meals", nutritionHandler.GetPatientMealLogs)
		admin.Get("/patients/:id/activities", routineHandler.GetPatientActivityLogs)
		admin.Get("/patients/:id/activities/education", eduProgressHandler.GetPatientEducationActivities)
		admin.Get("/patients/:id/medications", reminderHandler.GetPatientMedicationLogs)
		admin.Get("/patients/:id/activity-analytics", patientHandler.GetPatientActivityAnalytics)

		// Global Foods Management
		admin.Post("/foods", nutritionHandler.CreateFood)
		admin.Put("/foods/:id", nutritionHandler.UpdateFood)

		// Education content CRUD
		admin.Get("/education/stats", eduHandler.GetStats)
		admin.Get("/education/articles", eduHandler.ListAdmin)
		admin.Post("/education/articles", eduHandler.Create)
		admin.Put("/education/articles/:id", eduHandler.Update)
		admin.Patch("/education/articles/:id/publish", eduHandler.Publish)
		admin.Delete("/education/articles/:id", eduHandler.Delete)

		// Education Progress Tracking & Reviews
		admin.Get("/education/:id/progress", eduProgressHandler.GetArticleProgress)
		admin.Get("/education/:id/progress/analytics", eduProgressHandler.GetArticleAnalytics)
		admin.Post("/education/:id/progress/read-article", eduProgressHandler.MarkArticleReadAdmin)
		admin.Post("/education/:id/progress/watch-video", eduProgressHandler.MarkVideoWatchedAdmin)
		admin.Get("/education/:id/reviews", eduHandler.GetAdminReviews)

		// Quiz / Questionnaire Management
		admin.Get("/quiz/stats", quizHandler.GetStats)
		admin.Get("/quiz", quizHandler.List)
		admin.Get("/quiz/pre-test", quizHandler.GetActivePreTest)
		admin.Get("/quiz/post-test", quizHandler.GetPostTestByEducation)
		admin.Get("/quiz/:id", quizHandler.GetByID)
		admin.Post("/quiz", quizHandler.Create)
		admin.Put("/quiz/:id", quizHandler.Update)
		admin.Delete("/quiz/:id", quizHandler.Delete)
		admin.Get("/quiz/:id/participants", quizHandler.ListParticipants)
		admin.Get("/quiz/:id/participant/:participant_id", quizHandler.GetParticipantDetail)

		// Support Tickets Management
		admin.Get("/support/tickets", settingsHandler.GetAllTickets)
		admin.Patch("/support/tickets/:id/resolve", settingsHandler.ResolveTicket)
	}

	// ── Protected: Staff Group (JWT + RequireRole("staff")) ───────────────────
	staff := v1.Group("/staff",
		middleware.JWT(c.Config),
		middleware.RequireRole("staff"),
	)
	{
		staff.Get("/me", staffHandler.GetMe)
		staff.Put("/me", staffHandler.UpdateMe)
		staff.Put("/me/password", staffHandler.ChangePassword)

		// Patient Monitoring
		staff.Get("/patients/stats", patientHandler.GetStatsStaff)
		staff.Get("/patients", patientHandler.ListStaff)
		staff.Get("/patients/:id", patientHandler.GetByID)
		staff.Get("/patients/:id/measurements", patientHandler.GetPatientMeasurements)
		staff.Get("/patients/:id/blood-sugar", bsHandler.GetPatientHistory)
		staff.Get("/patients/:id/meals", nutritionHandler.GetPatientMealLogs)
		staff.Get("/patients/:id/activities", routineHandler.GetPatientActivityLogs)
		staff.Get("/patients/:id/activities/education", eduProgressHandler.GetPatientEducationActivities)
		staff.Get("/patients/:id/medications", reminderHandler.GetPatientMedicationLogs)
		staff.Get("/patients/:id/activity-analytics", patientHandler.GetPatientActivityAnalytics)

		// Dashboard statistics
		staff.Get("/dashboard/blood-sugar", bsHandler.GetDashboard)
		staff.Get("/dashboard/stats", dashboardHandler.GetStaff)
		staff.Get("/dashboard/population-metrics", dashboardHandler.GetPopulationMetrics)
		staff.Get("/dashboard/patient-trends", dashboardHandler.GetPatientTrends)

		// Education articles list
		staff.Get("/education/articles", eduHandler.ListPublished)
		staff.Get("/education/articles/:id", eduHandler.GetByID)

		// Quiz Monitoring
		staff.Get("/quiz/stats", quizHandler.GetStats)
		staff.Get("/quiz", quizHandler.List)
		staff.Get("/quiz/:id", quizHandler.GetByID)
		staff.Get("/quiz/:id/participants", quizHandler.ListParticipants)
		staff.Get("/quiz/:id/participant/:participant_id", quizHandler.GetParticipantDetail)
	}

	// ── Protected: Patient Group (JWT + RequireRole("user")) ───────────────
	patientGroup := v1.Group("/patient",
		middleware.JWT(c.Config),
		middleware.RequireRole("user"),
	)
	{
		patientGroup.Get("/me", patientHandler.GetMe)
		patientGroup.Put("/me", patientHandler.UpdateMe)
		patientGroup.Put("/me/password", patientHandler.ChangePassword)
		patientGroup.Post("/profile/setup", patientHandler.SetupHealthProfile)

		// Health Measurements
		patientGroup.Get("/measurements", patientHandler.GetPatientMeasurements)
		patientGroup.Post("/measurements", patientHandler.CreateMeasurement)

		// Routines habit logging
		patientGroup.Get("/routines", routineHandler.List)
		patientGroup.Put("/routines/:routineTimeId", routineHandler.Configure)
		patientGroup.Post("/routines/setup", routineHandler.BulkSetup)
		patientGroup.Post("/routines/log", routineHandler.Log)
		patientGroup.Post("/activities/log", routineHandler.LogActivity)
		patientGroup.Get("/routines/status", routineHandler.Status)

		// Blood sugar records
		patientGroup.Post("/blood-sugar", bsHandler.Log)
		patientGroup.Get("/blood-sugar", bsHandler.GetHistory)
		patientGroup.Put("/blood-sugar/:id", bsHandler.Update)
		patientGroup.Delete("/blood-sugar/:id", bsHandler.Delete)

		// Daily Checkin Calendar
		patientGroup.Post("/checkin", checkinHandler.Checkin)
		patientGroup.Get("/checkin/calendar", checkinHandler.GetCalendar)

		// Nutrition meal logs
		patientGroup.Get("/foods/recent", nutritionHandler.GetRecent)
		patientGroup.Post("/meals", nutritionHandler.LogMeal)
		patientGroup.Get("/meals/summary", nutritionHandler.GetSummary)
		patientGroup.Put("/meals/:id", nutritionHandler.UpdateMeal)
		patientGroup.Delete("/meals/:id", nutritionHandler.DeleteMeal)

		// Personal reminders
		patientGroup.Get("/reminders", reminderHandler.List)
		patientGroup.Post("/reminders", reminderHandler.Create)
		patientGroup.Put("/reminders/:id", reminderHandler.Update)
		patientGroup.Patch("/reminders/:id/toggle", reminderHandler.Toggle)
		patientGroup.Delete("/reminders/:id", reminderHandler.Delete)
		patientGroup.Post("/medications/log", reminderHandler.LogMedication)

		// Notifications inbox
		patientGroup.Get("/notifications", reminderHandler.GetNotifications)
		patientGroup.Patch("/notifications/read", reminderHandler.MarkRead)
		patientGroup.Patch("/notifications/:id/read", reminderHandler.MarkNotificationReadByID)
		patientGroup.Delete("/notifications/:id", reminderHandler.DeleteNotificationByID)
		patientGroup.Delete("/notifications", reminderHandler.DeleteNotificationByID)

		// Education bookmarks & completions
		patientGroup.Post("/education/:id/complete", eduHandler.Complete)
		patientGroup.Post("/education/:id/save", eduHandler.Save)
		patientGroup.Delete("/education/:id/save", eduHandler.Unsave)
		patientGroup.Get("/education/saved", eduHandler.ListSaved)

		// Education Progress (Mobile-ready API)
		admin.Get("/education/:id/progress", eduProgressHandler.GetPatientProgress)
		patientGroup.Get("/education/:id/progress", eduProgressHandler.GetPatientProgress)
		patientGroup.Post("/education/:id/read-article", eduProgressHandler.MarkArticleRead)
		patientGroup.Post("/education/:id/watch-video", eduProgressHandler.MarkVideoWatched)
		patientGroup.Post("/education/:id/review", eduHandler.SubmitReview)
		patientGroup.Get("/education/:id/review", eduHandler.GetReview)
		patientGroup.Get("/education/:id/rating", eduHandler.GetRatingSummary)

		// Weekly analytical summary cards
		patientGroup.Get("/summary/weekly", summaryHandler.GetLatest)

		// Support Ticket submission
		patientGroup.Post("/support/tickets", settingsHandler.SubmitTicket)
		patientGroup.Get("/support/tickets", settingsHandler.GetMyTickets)

		// Patient Activity History
		patientGroup.Get("/history", historyHandler.GetPatientHistory)
		patientGroup.Delete("/history/:type/:id", historyHandler.DeleteHistoryItem)

		// Questionnaire (Pre-Test / Post-Test) for Patient Mobile
		patientGroup.Get("/questionnaires", quizHandler.ListPatient)
		patientGroup.Get("/questionnaires/my-history", quizHandler.GetMyHistory)
		patientGroup.Get("/questionnaires/pre-test", quizHandler.GetActivePreTest)
		patientGroup.Get("/questionnaires/post-test", quizHandler.GetPostTestByEducation)
		patientGroup.Get("/questionnaires/:id", quizHandler.GetByID)
		patientGroup.Get("/questionnaires/:id/my-attempt", quizHandler.GetMyAttempt)
		patientGroup.Get("/questionnaires/:id/my-attempt/detail", quizHandler.GetMyAttemptDetail)
		patientGroup.Post("/questionnaires/:id/submit", quizHandler.Submit)

		// AI Personal Diabetes Assistant
		patientGroup.Get("/ai/conversations", aiChatHandler.ListConversations)
		patientGroup.Post("/ai/conversations", aiChatHandler.CreateConversation)
		patientGroup.Get("/ai/conversations/:id/messages", aiChatHandler.GetMessages)
		patientGroup.Delete("/ai/conversations/:id", aiChatHandler.DeleteConversation)
		patientGroup.Post("/ai/chat", aiChatHandler.SendMessage)
	}

	// Register Survey Module Routes
	survey.RegisterRoutes(admin, staff, patientGroup, surveyHandler)

	// ── Authenticated Shared Group (any authenticated role) ───────────────────
	jwtAuth := middleware.JWT(c.Config)
	sharedGroup := v1.Group("", jwtAuth)

	// Register Food Master Module Routes
	food.RegisterRoutes(admin, sharedGroup, foodHandler)

	aiGroup := v1.Group("/ai",
		jwtAuth,
		middleware.RequireRole("user"),
	)
	{
		aiGroup.Get("/conversations", aiChatHandler.ListConversations)
		aiGroup.Post("/conversations", aiChatHandler.CreateConversation)
		aiGroup.Get("/conversations/:id/messages", aiChatHandler.GetMessages)
		aiGroup.Delete("/conversations/:id", aiChatHandler.DeleteConversation)
		aiGroup.Post("/chat", aiChatHandler.SendMessage)
	}

	v1.Get("/foods", jwtAuth, nutritionHandler.Search)
	v1.Get("/faqs", jwtAuth, settingsHandler.GetFAQs)
	v1.Get("/education/categories", jwtAuth, eduHandler.ListCategories)
	v1.Get("/education/articles", jwtAuth, eduHandler.ListPublished)
	v1.Get("/education/articles/:id", jwtAuth, eduHandler.GetByID)
	v1.Post("/education/:id/review", jwtAuth, eduHandler.SubmitReview)
	v1.Get("/education/:id/review", jwtAuth, eduHandler.GetReview)
	v1.Get("/education/:id/rating", jwtAuth, eduHandler.GetRatingSummary)

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
