// Package config loads and exposes all application configuration values.
// Viper reads from the .env file and environment variables (env vars take priority).
// The Config struct is the single source of truth for every setting in the app.
package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config is the root configuration struct.
// All sub-configs are embedded so callers use cfg.DB.DSN, cfg.JWT.Secret, etc.
type Config struct {
	App      AppConfig
	DB       DatabaseConfig
	JWT      JWTConfig
	Log      LogConfig
	Swagger  SwaggerConfig
	Email    EmailConfig
}

// AppConfig holds HTTP server and general application settings.
type AppConfig struct {
	Name        string        `mapstructure:"APP_NAME"`
	Env         string        `mapstructure:"APP_ENV"`   // development | staging | production
	Port        string        `mapstructure:"APP_PORT"`
	BaseURL     string        `mapstructure:"APP_BASE_URL"`
	ReadTimeout time.Duration `mapstructure:"APP_READ_TIMEOUT"`
	WriteTimeout time.Duration `mapstructure:"APP_WRITE_TIMEOUT"`
	IdleTimeout time.Duration `mapstructure:"APP_IDLE_TIMEOUT"`
}

// DatabaseConfig holds PostgreSQL connection settings.
type DatabaseConfig struct {
	Host            string `mapstructure:"DB_HOST"`
	Port            string `mapstructure:"DB_PORT"`
	Name            string `mapstructure:"DB_NAME"`
	User            string `mapstructure:"DB_USER"`
	Password        string `mapstructure:"DB_PASSWORD"`
	SSLMode         string `mapstructure:"DB_SSLMODE"`
	MaxIdleConns    int    `mapstructure:"DB_MAX_IDLE_CONNS"`
	MaxOpenConns    int    `mapstructure:"DB_MAX_OPEN_CONNS"`
	ConnMaxLifetime int    `mapstructure:"DB_CONN_MAX_LIFETIME_MINUTES"`
}

// DSN builds the PostgreSQL data source name from the config fields.
func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s dbname=%s user=%s password=%s sslmode=%s TimeZone=Asia/Makassar",
		d.Host, d.Port, d.Name, d.User, d.Password, d.SSLMode,
	)
}

// JWTConfig holds JSON Web Token settings.
type JWTConfig struct {
	Secret          string        `mapstructure:"JWT_SECRET"`
	AccessTokenTTL  time.Duration `mapstructure:"JWT_ACCESS_TOKEN_TTL"`
	RefreshTokenTTL time.Duration `mapstructure:"JWT_REFRESH_TOKEN_TTL"`
	Issuer          string        `mapstructure:"JWT_ISSUER"`
}

// LogConfig controls structured logging behaviour.
type LogConfig struct {
	Level  string `mapstructure:"LOG_LEVEL"`  // debug | info | warn | error
	Format string `mapstructure:"LOG_FORMAT"` // json | console
}

// SwaggerConfig controls API documentation settings.
type SwaggerConfig struct {
	Enabled bool   `mapstructure:"SWAGGER_ENABLED"`
	Host    string `mapstructure:"SWAGGER_HOST"`
}

// EmailConfig holds Resend API configuration settings.
type EmailConfig struct {
	ResendAPIKey    string `mapstructure:"RESEND_API_KEY"`
	ResendFromEmail string `mapstructure:"RESEND_FROM_EMAIL"`
}

// Load reads configuration from the .env file and environment variables.
// Environment variables always override .env values (12-factor app).
func Load() (*Config, error) {
	v := viper.New()

	// ── Defaults ──────────────────────────────────────────────────────────────
	v.SetDefault("APP_NAME", "DSMES Backend")
	v.SetDefault("APP_ENV", "development")
	v.SetDefault("APP_PORT", "8080")
	v.SetDefault("APP_BASE_URL", "http://localhost:8080")
	v.SetDefault("APP_READ_TIMEOUT", 5*time.Second)
	v.SetDefault("APP_WRITE_TIMEOUT", 10*time.Second)
	v.SetDefault("APP_IDLE_TIMEOUT", 120*time.Second)

	v.SetDefault("DB_HOST", "localhost")
	v.SetDefault("DB_PORT", "5432")
	v.SetDefault("DB_NAME", "dsmes_db")
	v.SetDefault("DB_USER", "postgres")
	v.SetDefault("DB_PASSWORD", "")
	v.SetDefault("DB_SSLMODE", "disable")
	v.SetDefault("DB_MAX_IDLE_CONNS", 10)
	v.SetDefault("DB_MAX_OPEN_CONNS", 100)
	v.SetDefault("DB_CONN_MAX_LIFETIME_MINUTES", 60)

	v.SetDefault("JWT_SECRET", "change-me-in-production")
	v.SetDefault("JWT_ACCESS_TOKEN_TTL", 15*time.Minute)
	v.SetDefault("JWT_REFRESH_TOKEN_TTL", 7*24*time.Hour)
	v.SetDefault("JWT_ISSUER", "dsmes-backend")

	v.SetDefault("LOG_LEVEL", "info")
	v.SetDefault("LOG_FORMAT", "console")

	v.SetDefault("SWAGGER_ENABLED", true)
	v.SetDefault("SWAGGER_HOST", "localhost:8080")

	v.SetDefault("RESEND_API_KEY", "")
	v.SetDefault("RESEND_FROM_EMAIL", "onboarding@resend.dev")

	// ── .env file ─────────────────────────────────────────────────────────────
	v.SetConfigName(".env")
	v.SetConfigType("env")
	v.AddConfigPath(".")    // project root when running via `go run` or Air
	v.AddConfigPath("../")  // when running from cmd/api/

	if err := v.ReadInConfig(); err != nil {
		// It is acceptable if .env is missing — env vars alone are valid.
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("config: failed to read .env: %w", err)
		}
	}

	// ── Environment variables override .env ───────────────────────────────────
	// AutomaticEnv makes every env var (e.g. DB_HOST) override .env values.
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// ── Explicit mapping ──────────────────────────────────────────────────────
	// We intentionally use v.GetString/GetInt/GetDuration instead of
	// v.Unmarshal to avoid mapstructure tag reliance and prevent Viper from
	// silently skipping flat env-var keys that don't match nested struct paths.
	cfg := &Config{}

	// Map flat env keys into sub-structs manually (Viper flat key approach).
	cfg.App = AppConfig{
		Name:         v.GetString("APP_NAME"),
		Env:          v.GetString("APP_ENV"),
		Port:         v.GetString("APP_PORT"),
		BaseURL:      v.GetString("APP_BASE_URL"),
		ReadTimeout:  v.GetDuration("APP_READ_TIMEOUT"),
		WriteTimeout: v.GetDuration("APP_WRITE_TIMEOUT"),
		IdleTimeout:  v.GetDuration("APP_IDLE_TIMEOUT"),
	}
	cfg.DB = DatabaseConfig{
		Host:            v.GetString("DB_HOST"),
		Port:            v.GetString("DB_PORT"),
		Name:            v.GetString("DB_NAME"),
		User:            v.GetString("DB_USER"),
		Password:        v.GetString("DB_PASSWORD"),
		SSLMode:         v.GetString("DB_SSLMODE"),
		MaxIdleConns:    v.GetInt("DB_MAX_IDLE_CONNS"),
		MaxOpenConns:    v.GetInt("DB_MAX_OPEN_CONNS"),
		ConnMaxLifetime: v.GetInt("DB_CONN_MAX_LIFETIME_MINUTES"),
	}
	cfg.JWT = JWTConfig{
		Secret:          v.GetString("JWT_SECRET"),
		AccessTokenTTL:  v.GetDuration("JWT_ACCESS_TOKEN_TTL"),
		RefreshTokenTTL: v.GetDuration("JWT_REFRESH_TOKEN_TTL"),
		Issuer:          v.GetString("JWT_ISSUER"),
	}
	cfg.Log = LogConfig{
		Level:  v.GetString("LOG_LEVEL"),
		Format: v.GetString("LOG_FORMAT"),
	}
	cfg.Swagger = SwaggerConfig{
		Enabled: v.GetBool("SWAGGER_ENABLED"),
		Host:    v.GetString("SWAGGER_HOST"),
	}
	cfg.Email = EmailConfig{
		ResendAPIKey:    v.GetString("RESEND_API_KEY"),
		ResendFromEmail: v.GetString("RESEND_FROM_EMAIL"),
	}

	return cfg, nil
}

// IsDevelopment returns true when APP_ENV is "development".
func (c *Config) IsDevelopment() bool {
	return c.App.Env == "development"
}

// IsStaging returns true when APP_ENV is "staging".
func (c *Config) IsStaging() bool {
	return c.App.Env == "staging"
}

// IsProduction returns true when APP_ENV is "production".
func (c *Config) IsProduction() bool {
	return c.App.Env == "production"
}
