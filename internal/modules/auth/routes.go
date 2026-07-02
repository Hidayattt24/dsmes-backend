package auth

import (
	"github.com/gofiber/fiber/v3"

	"github.com/dsmes/dsmes-backend/internal/container"
	"github.com/dsmes/dsmes-backend/internal/middleware"
)

// RegisterRoutes mounts all auth routes on the provided router group.
// Call from cmd/api/routes.go:
//
//	v1 := app.Group("/api/v1")
//	auth.RegisterRoutes(v1, c)
func RegisterRoutes(router fiber.Router, c *container.Container) {
	// Wire: Repository ← Service ← Handler
	repo := NewAuthRepository(c.DB, c.Logger)
	svc := NewAuthService(repo, c.JWT, c.Logger)
	h := NewAuthHandler(svc, c.Logger)

	// Public auth routes — no JWT middleware
	auth := router.Group("/auth")
	auth.Post("/staff/login", h.StaffLogin)
	auth.Post("/patient/login", h.PatientLogin)
	auth.Post("/forgot-password", h.ForgotPassword)
	auth.Post("/verify-otp", h.VerifyOTP)
	auth.Post("/reset-password", h.ResetPassword)
	auth.Post("/refresh", h.RefreshToken)

	// Protected: logout requires a valid JWT (to extract session info)
	auth.Post("/logout", middleware.JWT(c.Config), h.Logout)
}
