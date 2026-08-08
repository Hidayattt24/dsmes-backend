// cmd/api/main.go — Application entrypoint.
//
// @title           DSMES Backend API
// @version         1.0.0
// @description     Diabetes Self-Management Education and Support — Backend API
// @termsOfService  http://swagger.io/terms/
//
// @contact.name    DSMES Team
// @contact.email
//
// @license.name    MIT
// @license.url     https://opensource.org/licenses/MIT
//
// @host            localhost:8080
// @BasePath        /api/v1
// @schemes         http https
//
// @securityDefinitions.apikey BearerAuth
// @in                         header
// @name                       Authorization
// @description                Type "Bearer" followed by a space and the JWT token.
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
