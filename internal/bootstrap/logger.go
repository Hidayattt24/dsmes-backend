// Package bootstrap initializes the Zap structured logger.
//
// Two modes are supported:
//   - development / staging → pretty-printed console output (human-readable)
//   - production            → JSON structured output (for log aggregators such as Loki, ELK)
//
// Architectural decision: the logger is NOT a global variable.
// It is constructed here and passed explicitly through dependency injection.
// This keeps each component testable in isolation.
package bootstrap

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/dsmes/dsmes-backend/internal/config"
)

// NewLogger builds and returns a *zap.Logger configured from cfg.Log.
// The caller is responsible for calling logger.Sync() on shutdown.
func NewLogger(cfg *config.Config) (*zap.Logger, error) {
	level, err := zapcore.ParseLevel(cfg.Log.Level)
	if err != nil {
		return nil, fmt.Errorf("logger: invalid log level %q: %w", cfg.Log.Level, err)
	}

	var zapCfg zap.Config

	if cfg.Log.Format == "json" || cfg.IsProduction() {
		// Production: JSON output for log aggregators (Loki, ELK, Datadog, etc.)
		zapCfg = zap.NewProductionConfig()
	} else {
		// Development / staging: coloured, human-readable console output.
		zapCfg = zap.NewDevelopmentConfig()
		zapCfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	zapCfg.Level = zap.NewAtomicLevelAt(level)

	logger, err := zapCfg.Build(
		zap.AddCallerSkip(0),
		// AddCaller adds file:line to every log entry — valuable in production post-mortems.
		zap.WithCaller(true),
	)
	if err != nil {
		return nil, fmt.Errorf("logger: failed to build zap logger: %w", err)
	}

	return logger, nil
}
