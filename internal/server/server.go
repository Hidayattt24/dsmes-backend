// Package server manages the Fiber HTTP server lifecycle.
//
// Graceful shutdown strategy:
//  1. OS sends SIGINT (Ctrl+C) or SIGTERM (Docker / systemd stop).
//  2. server.Start() receives the signal on the quit channel.
//  3. app.ShutdownWithTimeout() stops accepting new connections and waits up
//     to ShutdownTimeout for in-flight requests to complete.
//  4. bootstrap.CloseDatabase() closes the PostgreSQL connection pool.
//  5. logger.Sync() flushes any buffered log entries.
//  6. Process exits cleanly with code 0.
//
// Architectural decision: the server package owns the Listen + shutdown loop.
// This keeps main.go minimal — it only wires dependencies and calls server.Start().
package server

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/dsmes/dsmes-backend/internal/bootstrap"
	"github.com/dsmes/dsmes-backend/internal/config"
)

const (
	// ShutdownTimeout is the maximum time to wait for in-flight requests to finish.
	ShutdownTimeout = 10 * time.Second
)

// Start begins listening on the configured port and blocks until a shutdown
// signal is received. It performs a graceful shutdown before returning.
func Start(app *fiber.App, db *gorm.DB, log *zap.Logger, cfg *config.Config) error {
	// ── Start listening in a goroutine ────────────────────────────────────────
	// app.Listen() is blocking, so we run it in a goroutine and collect any
	// startup errors on the errCh channel.
	errCh := make(chan error, 1)

	go func() {
		addr := fmt.Sprintf(":%s", cfg.App.Port)
		log.Info("server starting",
			zap.String("addr", addr),
			zap.String("env", cfg.App.Env),
			zap.String("app", cfg.App.Name),
		)

		if err := app.Listen(addr, fiber.ListenConfig{
			// DisableStartupMessage: we log our own structured startup message above.
			DisableStartupMessage: true,
		}); err != nil {
			errCh <- fmt.Errorf("server: listen error: %w", err)
		}
	}()

	// ── Wait for OS signal ────────────────────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errCh:
		// Server failed to start — return the error immediately.
		return err

	case sig := <-quit:
		log.Info("shutdown signal received", zap.String("signal", sig.String()))
	}

	// ── Graceful shutdown ─────────────────────────────────────────────────────
	log.Info("initiating graceful shutdown", zap.Duration("timeout", ShutdownTimeout))

	// ShutdownWithTimeout drains in-flight requests within the timeout window.
	if err := app.ShutdownWithTimeout(ShutdownTimeout); err != nil {
		log.Error("server: shutdown error", zap.Error(err))
	}

	// Close PostgreSQL connection pool after all requests are done.
	bootstrap.CloseDatabase(db, log)

	// Flush Zap's buffered log entries (important for async JSON appenders).
	_ = log.Sync()

	log.Info("server stopped gracefully")
	return nil
}
