// Package bootstrap configures GORM with a PostgreSQL driver.
//
// Connection pool settings follow official GORM recommendations:
//   - MaxIdleConns:    10   (keep warm connections for fast queries)
//   - MaxOpenConns:    100  (bound the number of concurrent DB connections)
//   - ConnMaxLifetime: 1h   (recycle connections to avoid stale TCP sessions)
//
// Architectural decision: the *gorm.DB instance is returned and stored in the
// application container, then passed to every repository via dependency injection.
// No global DB variable is used — this enforces testability and isolation.
package bootstrap

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/dsmes/dsmes-backend/internal/config"
)


// NewDatabase opens a GORM connection to PostgreSQL and configures the
// underlying database/sql connection pool.
func NewDatabase(cfg *config.Config, log *zap.Logger) (*gorm.DB, error) {
	// ── GORM logger mode ──────────────────────────────────────────────────────
	// In development: log all SQL statements to make queries visible.
	// In production: log only slow queries and errors to reduce noise.
	var gormLogLevel gormlogger.LogLevel
	if cfg.IsDevelopment() {
		gormLogLevel = gormlogger.Info
	} else {
		gormLogLevel = gormlogger.Warn
	}

	gormConfig := &gorm.Config{
		// PrepareStmt caches prepared statements — improves repeated-query performance.
		PrepareStmt: true,

		// DisableForeignKeyConstraintWhenMigrating prevents GORM AutoMigrate from
		// creating FK constraints automatically. We manage migrations manually.
		DisableForeignKeyConstraintWhenMigrating: true,

		Logger: gormlogger.Default.LogMode(gormLogLevel),
	}

	log.Info("connecting to PostgreSQL", zap.String("host", cfg.DB.Host), zap.String("dbname", cfg.DB.Name))

	db, err := gorm.Open(postgres.Open(cfg.DB.DSN()), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("database: failed to connect to PostgreSQL: %w", err)
	}

	// ── Connection pool ───────────────────────────────────────────────────────
	// Source: https://gorm.io/docs/connecting_to_the_database.html
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("database: failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(cfg.DB.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.DB.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.DB.ConnMaxLifetime) * time.Minute)

	// ── Ping ──────────────────────────────────────────────────────────────────
	if err = sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("database: failed to ping PostgreSQL: %w", err)
	}

	log.Info("PostgreSQL connected successfully",
		zap.String("host", cfg.DB.Host),
		zap.String("dbname", cfg.DB.Name),
		zap.Int("maxIdleConns", cfg.DB.MaxIdleConns),
		zap.Int("maxOpenConns", cfg.DB.MaxOpenConns),
	)

	return db, nil
}


// CloseDatabase closes the underlying sql.DB connection pool.
// Called during graceful shutdown.
func CloseDatabase(db *gorm.DB, log *zap.Logger) {
	sqlDB, err := db.DB()
	if err != nil {
		log.Error("database: failed to get sql.DB for close", zap.Error(err))
		return
	}

	if err = sqlDB.Close(); err != nil {
		log.Error("database: failed to close connection pool", zap.Error(err))
		return
	}

	log.Info("database connection pool closed")
}
