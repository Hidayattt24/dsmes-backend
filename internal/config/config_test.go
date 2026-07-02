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

	if cfg.DB.Host != "localhost" {
		t.Errorf("expected DB.Host to be 'localhost', got: %s", cfg.DB.Host)
	}
}

func TestEnvironmentOverrides(t *testing.T) {
	// Set environment variables to test override behavior
	os.Setenv("APP_ENV", "production")
	os.Setenv("APP_PORT", "9090")
	os.Setenv("DB_NAME", "production_dsmes_db")

	defer func() {
		os.Unsetenv("APP_ENV")
		os.Unsetenv("APP_PORT")
		os.Unsetenv("DB_NAME")
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

	if !cfg.IsProduction() {
		t.Error("expected IsProduction() to be true")
	}

	if cfg.IsDevelopment() {
		t.Error("expected IsDevelopment() to be false")
	}
}
