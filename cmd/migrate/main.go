package main

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"

	"github.com/dsmes/dsmes-backend/internal/bootstrap"
	"github.com/dsmes/dsmes-backend/internal/config"
)

func main() {
	// Initialize logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load config", zap.Error(err))
	}

	// Connect to database
	db, err := bootstrap.NewDatabase(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer func() {
		bootstrap.CloseDatabase(db, logger)
	}()

	logger.Info("Starting database migration process...")

	// Migration files in order
	migrationFiles := []string{
		"000001_create_auth_and_staff_tables.up.sql",
		"000002_create_remaining_tables.up.sql",
		"000003_add_patient_gaps_and_quiz_tables.up.sql",
	}

	for _, fileName := range migrationFiles {
		filePath := filepath.Join("migrations", fileName)
		logger.Info("Reading migration file", zap.String("file", filePath))

		content, err := os.ReadFile(filePath)
		if err != nil {
			logger.Fatal("Failed to read migration file", zap.String("file", filePath), zap.Error(err))
		}

		sqlDB, err := db.DB()
		if err != nil {
			logger.Fatal("Failed to get underlying sql.DB", zap.Error(err))
		}

		logger.Info("Executing migration", zap.String("file", fileName))
		// Execute the raw SQL using raw SQL database connection to support multiple semicolon-separated commands
		if _, err := sqlDB.Exec(string(content)); err != nil {
			logger.Fatal("Failed to execute migration", zap.String("file", fileName), zap.Error(err))
		}
		logger.Info("Migration file applied successfully", zap.String("file", fileName))
	}

	logger.Info("All database migrations applied successfully!")
}
