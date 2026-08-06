// Package middleware — cors.go
//
// CORS (Cross-Origin Resource Sharing) allows the frontend (Flutter Web /
// Next.js) to make requests from a different origin than the API server.
//
// Policy:
//   - Development: allow all origins (permissive, convenient for local dev).
//   - Production:  restrict to the explicit list of allowed origins from config.
//
// Security note: do NOT use AllowOrigins: "*" with AllowCredentials: true —
// this is rejected by browsers. The production config handles this correctly.
package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"

	"github.com/dsmes/dsmes-backend/internal/config"
)

// CORS returns a configured CORS middleware based on the application environment.
func CORS(cfg *config.Config) fiber.Handler {
	if cfg.IsDevelopment() {
		// Development: permissive — allow any origin so frontend dev servers work.
		return cors.New(cors.Config{
			AllowOrigins: []string{"*"},
			AllowHeaders: []string{
				"Origin", "Content-Type", "Accept",
				"Authorization", "X-Request-ID",
			},
			AllowMethods: []string{
				"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS",
			},
		})
	}

	// Production: restrict to the explicit list of allowed origins from config.
	// Configure them with APP_ALLOWED_ORIGINS (comma-separated) in .env; falls
	// back to the API's own base URL when the list is empty.
	origins := cfg.App.AllowedOrigins
	if len(origins) == 0 {
		origins = []string{cfg.App.BaseURL}
	}
	return cors.New(cors.Config{
		AllowOrigins: origins,
		AllowHeaders: []string{
			"Origin", "Content-Type", "Accept",
			"Authorization", "X-Request-ID",
		},
		AllowMethods: []string{
			"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS",
		},
		AllowCredentials: false,
	})
}
