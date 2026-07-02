// cmd/api/main.go — Application entrypoint.
//
// Responsibilities (in order):
//  1. Build the dependency container (config → logger → db → app → jwt).
//  2. Register routes on the Fiber application.
//  3. Start the HTTP server and block until graceful shutdown.
//
// This file contains ZERO business logic. All wiring is delegated to the
// container package. Adding a new module means calling its router from
// registerRoutes() — nothing here changes.
//
// Dependency injection strategy:
//
//	container.Build() → *Container (all deps in one struct)
//	↓
//	registerRoutes(c.App, c)   ← passes container to route layer
//	↓
//	ModuleRouter(v1, c.DB, c.JWT, c.Logger)  ← handler/service/repo receive only what they need
package main

import (
	"log"

	_ "github.com/dsmes/dsmes-backend/docs" // Swagger generated docs — blank import triggers init()

	"github.com/dsmes/dsmes-backend/internal/container"
	"github.com/dsmes/dsmes-backend/internal/server"
)

func main() {
	// Build all dependencies (config → logger → db → fiber app → jwt).
	// If any step fails the error is logged and the process exits non-zero.
	c, err := container.Build()
	if err != nil {
		log.Fatalf("failed to build application container: %v", err)
	}
	defer c.Close()

	// Register all routes on the Fiber application.
	registerRoutes(c.App, c)

	// Start the HTTP server. Blocks until SIGINT or SIGTERM is received,
	// then performs graceful shutdown (drains requests, closes DB, flushes logs).
	if err = server.Start(c.App, c.DB, c.Logger, c.Config); err != nil {
		c.Logger.Fatal("server exited with error")
	}
}
