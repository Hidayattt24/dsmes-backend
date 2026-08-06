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
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
)

// clientIP extracts the original client IP. When behind a reverse proxy it uses
// the left-most (original) entry of X-Forwarded-For; otherwise the socket peer.
// Parsing only the first element prevents a spoofed multi-IP header from being
// used verbatim as the rate-limit key.
func clientIP(c fiber.Ctx) string {
	forwarded := c.Get("X-Forwarded-For")
	if forwarded != "" {
		if i := strings.Index(forwarded, ","); i != -1 {
			forwarded = forwarded[:i]
		}
		return strings.TrimSpace(forwarded)
	}
	return c.IP()
}

// RateLimiter returns the global rate limiting middleware.
func RateLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		// Max requests allowed within the Expiration window.
		Max: 100,

		// Expiration defines the sliding window duration.
		Expiration: 1 * time.Minute,

		// KeyGenerator identifies unique clients by their IP address.
		KeyGenerator: func(c fiber.Ctx) string {
			return clientIP(c)
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

// StrictRateLimiter returns a tighter limiter for sensitive endpoints such as
// authentication/OTP, where 100 requests/minute would allow brute force.
func StrictRateLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        10,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c fiber.Ctx) string {
			return clientIP(c)
		},
		LimitReached: func(c fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success": false,
				"message": "too many requests — please try again later",
			})
		},
	})
}
