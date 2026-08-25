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
	svc := NewAuthService(repo, c.JWT, c.Email, c.Logger)
	h := NewAuthHandler(svc, c.Logger)

	// Public auth routes — no JWT middleware.
	// A strict limiter protects login/OTP endpoints from brute force.
	auth := router.Group("/auth", middleware.StrictRateLimiter())
	auth.Post("/login", h.PatientLogin)
	auth.Post("/staff/login", h.StaffLogin)
	auth.Post("/patient/login", h.PatientLogin)

	auth.Post("/forgot-password", h.ForgotPassword)
	auth.Post("/verify-otp", h.VerifyOTP)
	auth.Post("/reset-password", h.ResetPassword)

	auth.Post("/forgot-password/check-phone", h.CheckPhoneNumber)
	auth.Post("/reset-password-by-phone", h.ResetPasswordByPhone)

	auth.Post("/forgot-password/check-email", h.CheckEmail)
	auth.Post("/reset-password-by-email", h.ResetPasswordByEmail)

	// Refresh is rate-limited but more lenient than OTP (valid JWT needed).
	auth.Post("/refresh", middleware.RateLimiter(), h.RefreshToken)

	// Protected: logout requires a valid JWT (to extract session info)
	auth.Post("/logout", middleware.JWT(c.Config), h.Logout)
}
