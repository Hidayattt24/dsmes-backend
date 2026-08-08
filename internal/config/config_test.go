package config

import (
	"os"
	"testing"
)

func TestConfigDefaults(t *testing.T) {
	// Clear environment variables that might interfere with defaults
	os.Unsetenv("APP_ENV")
	os.Unsetenv("APP_PORT")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error loading config, got: %v", err)
	}

	// Verify defaults are populated
	if cfg.App.Name != "DSMES Backend" {
		t.Errorf("expected App.Name to be 'DSMES Backend', got: %s", cfg.App.Name)
	}

	if cfg.App.Port != "8080" {
		t.Errorf("expected App.Port to be '8080', got: %s", cfg.App.Port)
	}

	if cfg.App.Timezone != "Asia/Jakarta" {
		t.Errorf("expected App.Timezone to default to 'Asia/Jakarta', got: %s", cfg.App.Timezone)
	}

	if cfg.DB.Timezone != "Asia/Jakarta" {
		t.Errorf("expected DB.Timezone to default to 'Asia/Jakarta', got: %s", cfg.DB.Timezone)
	}

	if cfg.DB.Host != "localhost" {
		t.Errorf("expected DB.Host to be 'localhost', got: %s", cfg.DB.Host)
	}
}

func TestEnvironmentOverrides(t *testing.T) {
	// Set environment variables to test override behavior
	os.Setenv("APP_ENV", "production")
	os.Setenv("APP_PORT", "9090")
	os.Setenv("DB_NAME", "production_dsmes_db")
	os.Setenv("JWT_SECRET", "this-is-a-strong-random-secret-for-testing-only-1234")
	os.Setenv("APP_ALLOWED_ORIGINS", "https://app.example.com, https://admin.example.com")
	os.Setenv("APP_TIMEZONE", "Asia/Makassar")

	defer func() {
		os.Unsetenv("APP_ENV")
		os.Unsetenv("APP_PORT")
		os.Unsetenv("DB_NAME")
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("APP_ALLOWED_ORIGINS")
		os.Unsetenv("APP_TIMEZONE")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error loading config, got: %v", err)
	}

	if cfg.App.Env != "production" {
		t.Errorf("expected App.Env to override to 'production', got: %s", cfg.App.Env)
	}

	if cfg.App.Port != "9090" {
		t.Errorf("expected App.Port to override to '9090', got: %s", cfg.App.Port)
	}

	if cfg.DB.Name != "production_dsmes_db" {
		t.Errorf("expected DB.Name to override to 'production_dsmes_db', got: %s", cfg.DB.Name)
	}

	if len(cfg.App.AllowedOrigins) != 2 {
		t.Errorf("expected 2 allowed origins, got: %v", cfg.App.AllowedOrigins)
	}

	if cfg.App.Timezone != "Asia/Makassar" {
		t.Errorf("expected App.Timezone to override to 'Asia/Makassar', got: %s", cfg.App.Timezone)
	}

	if !cfg.IsProduction() {
		t.Error("expected IsProduction() to be true")
	}

	if cfg.IsDevelopment() {
		t.Error("expected IsDevelopment() to be false")
	}
}

func TestProductionGuardsWeakJWTSecret(t *testing.T) {
	os.Setenv("APP_ENV", "production")
	os.Setenv("JWT_SECRET", "change-me-in-production")

	defer func() {
		os.Unsetenv("APP_ENV")
		os.Unsetenv("JWT_SECRET")
	}()

	if _, err := Load(); err == nil {
		t.Error("expected error when production uses the default JWT secret")
	}
}
