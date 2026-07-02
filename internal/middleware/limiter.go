// Package middleware — limiter.go
//
// RateLimiter protects every API endpoint from brute-force attacks and
// excessive request floods. Applied globally AFTER CORS so pre-flight OPTIONS
// requests are not counted against the limit.
//
// Configuration:
//   - Max: 100 requests per window (generous for normal app usage).
//   - Window: 1 minute.
//   - Key: client IP address.
//
// For stricter limits on sensitive endpoints (e.g. /auth/login), apply an
// additional, more restrictive limiter.New() at the route group level.
package middleware

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
)

// RateLimiter returns the global rate limiting middleware.
func RateLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		// Max requests allowed within the Expiration window.
		Max: 100,

		// Expiration defines the sliding window duration.
		Expiration: 1 * time.Minute,

		// KeyGenerator identifies unique clients by their IP address.
		// In production behind a reverse proxy, use X-Forwarded-For.
		KeyGenerator: func(c fiber.Ctx) string {
			// Prefer the forwarded IP when behind nginx / Cloudflare.
			forwarded := c.Get("X-Forwarded-For")
			if forwarded != "" {
				return forwarded
			}
			return c.IP()
		},

		// LimitReached returns a structured JSON 429 response.
		LimitReached: func(c fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success": false,
				"message": "too many requests — please slow down and try again later",
			})
		},
	})
}
