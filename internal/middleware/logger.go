// Package middleware — logger.go
//
// RequestLogger is an HTTP request/response logger middleware.
// It logs method, path, status code, latency, and requester IP for every request.
//
// Backed by Fiber's built-in logger middleware, configured to write to a
// Zap-compatible io.Writer so all logs flow through the same Zap instance.
//
// Format (development):  [METHOD] /path  STATUS  LATENCY  IP
// Format (production):   JSON line per request (through Zap's JSON encoder)
package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// RequestLogger returns the HTTP request logging middleware.
// The provided *zap.Logger is used as the underlying writer so all log output
// is consistent with the application log format (JSON in prod / console in dev).
func RequestLogger(log *zap.Logger, timezone string) fiber.Handler {
	// Zap's sugar logger writes to its internal core; we create a bridge writer
	// so Fiber's logger middleware forwards its formatted lines to Zap.
	writer := &zapWriter{logger: log.With(zap.String("component", "http"))}

	if timezone == "" {
		timezone = "Asia/Jakarta"
	}

	return logger.New(logger.Config{
		// Format includes: timestamp, method, path, status, latency, IP.
		Format: "[HTTP] ${time} | ${status} | ${latency} | ${ip} | ${method} ${path}\n",

		// TimeFormat and TimeZone align with APP_TIMEZONE (default Asia/Jakarta / WIB).
		TimeFormat: "2006-01-02 15:04:05",
		TimeZone:   timezone,

		// Stream routes formatted log lines to our Zap bridge writer.
		Stream: writer,
	})
}

// zapWriter bridges Fiber's logger output (io.Writer) to Zap.
type zapWriter struct {
	logger *zap.Logger
}

// Write implements io.Writer. Each call represents one formatted log line.
func (w *zapWriter) Write(p []byte) (n int, err error) {
	if ce := w.logger.Check(zapcore.InfoLevel, string(p)); ce != nil {
		ce.Write()
	}
	return len(p), nil
}
