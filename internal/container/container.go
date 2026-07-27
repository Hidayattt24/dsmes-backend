// Package container implements the dependency injection container.
//
// # Why a Container?
//
// As the application grows, manually threading dependencies through function
// arguments in main.go quickly becomes unmanageable. A typed Container struct:
//
//   - Makes all dependencies explicit and discoverable
//   - Prevents accidental access to uninitialised dependencies
//   - Makes testing easier — swap any field with a mock
//   - Keeps main.go minimal: Build() → Start()
//
// # Architecture Position
//
// The Container lives at the boundary between infrastructure and application.
// It holds the bootstrapped infrastructure (DB, Logger, Config, JWT) and is
// the single source from which business module constructors receive their deps.
//
// # Usage
//
//	c, err := container.Build()
//	if err != nil { log.Fatal(err) }
//	defer c.Close()
//
//	registerRoutes(c.App, c)
//	server.Start(c.App, c.DB, c.Logger, c.Config)
package container

import (
	"fmt"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/dsmes/dsmes-backend/internal/bootstrap"
	"github.com/dsmes/dsmes-backend/internal/config"
	"github.com/dsmes/dsmes-backend/internal/domain"
	"github.com/dsmes/dsmes-backend/internal/infrastructure/email"
	jwtpkg "github.com/dsmes/dsmes-backend/internal/pkg/jwt"
)

// Container holds every bootstrapped dependency in the application.
// All fields are exported so business modules can access them.
// Never add business logic to this struct — it is infrastructure only.
type Container struct {
	// Config is the typed application configuration loaded from .env.
	Config *config.Config

	// Logger is the Zap structured logger instance.
	// Use Logger.With(zap.String("module", "auth")) to add module context.
	Logger *zap.Logger

	// DB is the GORM database connection with a configured connection pool.
	// Repositories receive *gorm.DB via constructor injection — never access
	// this directly from handlers or services.
	DB *gorm.DB

	// App is the Fiber HTTP application instance with all global middleware
	// already registered. Routes are mounted after Build() returns.
	App *fiber.App

	// JWT is the token manager used by the auth service to issue and validate
	// JWT access/refresh token pairs.
	JWT *jwtpkg.Manager

	// Email is the Resend-backed email service used to send OTP and welcome emails.
	Email email.EmailService
}

// Build initialises all infrastructure dependencies in the correct order and
// returns a fully populated Container ready to have routes registered.
//
// Initialisation order (each step depends on the previous):
//  1. Config  — reads .env and environment variables
//  2. Logger  — needs log level/format from Config
//  3. DB      — needs DSN and pool settings from Config + Logger for logging
//  4. App     — needs Config for timeouts and Logger for request logging
//  5. JWT     — needs JWT secret/TTL from Config
func Build() (*Container, error) {
	// 1. Configuration
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("container: failed to load config: %w", err)
	}

	// 2. Logger
	logger, err := bootstrap.NewLogger(cfg)
	if err != nil {
		return nil, fmt.Errorf("container: failed to initialise logger: %w", err)
	}

	// 3. Database
	db, err := bootstrap.NewDatabase(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("container: failed to connect to database: %w", err)
	}

	// Auto-migrate education tracking models to ensure new columns exist in DB table
	if err := db.AutoMigrate(&domain.UserArticleCompletion{}, &domain.PatientEducationActivity{}); err != nil {
		logger.Warn("container: failed to auto-migrate education tracking models", zap.Error(err))
	}

	// 4. Fiber application (global middleware already registered)
	app := bootstrap.NewFiberApp(cfg, logger)

	// 5. JWT manager
	jwtManager := jwtpkg.NewManager(cfg)

	// 6. Email service
	emailService := email.NewResendEmailService(cfg, logger)

	return &Container{
		Config: cfg,
		Logger: logger,
		DB:     db,
		App:    app,
		JWT:    jwtManager,
		Email:  emailService,
	}, nil
}

// Close performs a graceful teardown of all resources held by the container.
// Call this with defer immediately after Build() succeeds in main.go.
func (c *Container) Close() {
	bootstrap.CloseDatabase(c.DB, c.Logger)
	_ = c.Logger.Sync()
}
