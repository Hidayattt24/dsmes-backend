package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/dsmes/dsmes-backend/internal/bootstrap"
	"github.com/dsmes/dsmes-backend/internal/config"
)

const (
	migrationsDir = "migrations"
	// dsmes_migrations tracks applied migration files. The name deliberately
	// differs from golang-migrate's legacy "schema_migrations" table (which
	// exists in older databases and uses a different schema), so the two never
	// collide.
	schemaMigrationsTbl = "dsmes_migrations"
)

// legacyBaseline lists the exact migration files the pre-tracking runner
// executed (it ran these on every start, swallowing already-exists errors).
// It is only used to bootstrap a legacy database whose schema_migrations table
// is empty but whose tables already exist. Migrations added later (e.g. 000016,
// 000019) are intentionally NOT in this list so they still get applied.
var legacyBaseline = []string{
	"000001_create_auth_and_staff_tables.up.sql",
	"000002_create_remaining_tables.up.sql",
	"000003_add_patient_gaps_and_quiz_tables.up.sql",
	"000004_add_patient_extended_fields.up.sql",
	"000004_redesign_questionnaires.up.sql",
	"000005_add_routine_icon_and_schedule.up.sql",
	"000006_add_education_progress_fields.up.sql",
	"000007_add_patient_measurements_and_calorie_recommendations.up.sql",
	"000008_add_education_tracking_and_activities.up.sql",
	"000009_fix_quiz_fk_constraints.up.sql",
	"000010_create_patient_activity_logs.up.sql",
	"000011_create_ai_chat_tables.up.sql",
	"000012_auth_phone_number_primary.up.sql",
	"000013_pre_test_dmses_refactor.up.sql",
	"000014_add_notif_type_to_notification_logs.up.sql",
	"000015_create_education_reviews.up.sql",
	"000017_update_foods_table_schema.up.sql",
	"000018_add_nutrition_basis_to_foods.up.sql",
}

func main() {
	rollback := flag.Int("rollback", 0, "rollback the last N applied migrations (best-effort)")
	flag.Parse()

	logger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load config", zap.Error(err))
	}

	db, err := bootstrap.NewDatabase(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer func() {
		bootstrap.CloseDatabase(db, logger)
	}()

	if err := ensureSchemaMigrationsTable(db); err != nil {
		logger.Fatal("Failed to create schema_migrations table", zap.Error(err))
	}

	if *rollback > 0 {
		if err := runRollback(db, *rollback, logger); err != nil {
			logger.Fatal("Migration rollback failed", zap.Error(err))
		}
		logger.Info("Migration rollback completed successfully")
		return
	}

	if err := runUp(db, logger); err != nil {
		logger.Fatal("Migration failed", zap.Error(err))
	}
	logger.Info("All database migrations applied successfully!")
}

// ensureSchemaMigrationsTable creates the bookkeeping table used to track
// which migration files have already been applied. This makes the runner
// idempotent and safe to run on every deployment.
func ensureSchemaMigrationsTable(db *gorm.DB) error {
	return db.Exec(fmt.Sprintf(
		`CREATE TABLE IF NOT EXISTS %s (
			version    TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`, schemaMigrationsTbl)).Error
}

// listUpMigrations returns all *.up.sql filenames from the migrations dir,
// sorted lexicographically so version ordering is deterministic.
func listUpMigrations() ([]string, error) {
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".up.sql") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)
	return files, nil
}

// appliedVersions returns the set of migration filenames already applied.
func appliedVersions(db *gorm.DB) (map[string]bool, error) {
	rows, err := db.Table(schemaMigrationsTbl).Select("version").Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		applied[v] = true
	}
	return applied, rows.Err()
}

// isLegacyDB reports whether the connected database already contains the core
// `patients` table, meaning it was populated by the pre-tracking migration
// runner (schema_migrations tracking was added later).
func isLegacyDB(db *gorm.DB) (bool, error) {
	var exists bool
	err := db.Raw("SELECT to_regclass('patients') IS NOT NULL").Scan(&exists).Error
	if err != nil {
		return false, err
	}
	return exists, nil
}

// runUp applies every pending migration in order. Each file is recorded in
// schema_migrations only after it succeeds; a failure aborts the run so a
// partially migrated database is never silently ignored.
func runUp(db *gorm.DB, logger *zap.Logger) error {
	files, err := listUpMigrations()
	if err != nil {
		return fmt.Errorf("failed to list migration files: %w", err)
	}
	if len(files) == 0 {
		return fmt.Errorf("no migration files found in %s", migrationsDir)
	}

	applied, err := appliedVersions(db)
	if err != nil {
		return fmt.Errorf("failed to read applied migrations: %w", err)
	}

	// Legacy bootstrap: databases created before this version-tracked runner
	// have no schema_migrations rows. If the DB already contains tables (the
	// old runner applied everything best-effort), baseline the previously
	// applied versions so only new migrations run. Fresh databases skip this.
	if len(applied) == 0 {
		legacy, lerr := isLegacyDB(db)
		if lerr != nil {
			return fmt.Errorf("failed to detect legacy database: %w", lerr)
		}
		if legacy {
			logger.Info("Detected legacy database — baselining previously applied migrations")
			for _, v := range legacyBaseline {
				if applied[v] {
					continue
				}
				applied[v] = true
				if err := db.Table(schemaMigrationsTbl).Create(map[string]any{
					"version": v,
				}).Error; err != nil {
					return fmt.Errorf("failed to record baseline migration %s: %w", v, err)
				}
			}
		}
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	for _, file := range files {
		if applied[file] {
			logger.Info("Migration already applied — skipping", zap.String("file", file))
			continue
		}

		content, err := os.ReadFile(filepath.Join(migrationsDir, file))
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file, err)
		}

		logger.Info("Applying migration", zap.String("file", file))
		if _, err := sqlDB.Exec(string(content)); err != nil {
			return fmt.Errorf("migration %s failed: %w", file, err)
		}

		if err := db.Table(schemaMigrationsTbl).Create(map[string]any{
			"version": file,
		}).Error; err != nil {
			return fmt.Errorf("failed to record migration %s: %w", file, err)
		}
		logger.Info("Migration applied successfully", zap.String("file", file))
	}
	return nil
}

// runRollback applies the down migrations for the last N applied versions in
// reverse order (most recent first) and removes their tracking records.
func runRollback(db *gorm.DB, n int, logger *zap.Logger) error {
	applied, err := appliedVersions(db)
	if err != nil {
		return fmt.Errorf("failed to read applied migrations: %w", err)
	}

	var versions []string
	for v := range applied {
		if strings.HasSuffix(v, ".up.sql") {
			versions = append(versions, v)
		}
	}
	sort.Sort(sort.Reverse(sort.StringSlice(versions)))

	if len(versions) == 0 {
		logger.Info("No applied migrations to roll back")
		return nil
	}
	if n > len(versions) {
		n = len(versions)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	for i := 0; i < n; i++ {
		up := versions[i]
		down := strings.TrimSuffix(up, ".up.sql") + ".down.sql"
		path := filepath.Join(migrationsDir, down)

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("down migration %s not found: %w", down, err)
		}

		logger.Info("Rolling back migration", zap.String("file", down))
		if _, err := sqlDB.Exec(string(content)); err != nil {
			return fmt.Errorf("down migration %s failed: %w", down, err)
		}

		if err := db.Table(schemaMigrationsTbl).Where("version = ?", up).Delete(nil).Error; err != nil {
			return fmt.Errorf("failed to remove migration record %s: %w", up, err)
		}
		logger.Info("Migration rolled back", zap.String("file", down))
	}
	return nil
}
