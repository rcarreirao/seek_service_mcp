// Package config loads and validates the application configuration from
// environment variables, applying development-friendly defaults where safe.
package config

import (
	"errors"
	"os"
	"time"
)

// Config holds all runtime configuration for the API process.
type Config struct {
	HTTPAddr        string
	DatabaseDSN     string
	JWTAdminSecret  string
	JWTPanelSecret  string
	ShutdownTimeout time.Duration
}

var (
	// ErrDatabaseDSNMissing is returned when DATABASE_DSN is empty.
	ErrDatabaseDSNMissing = errors.New("config: DATABASE_DSN is required")
	// ErrJWTAdminSecretMissing is returned when JWT_ADMIN_SECRET is empty.
	ErrJWTAdminSecretMissing = errors.New("config: JWT_ADMIN_SECRET is required")
	// ErrJWTPanelSecretMissing is returned when JWT_PANEL_SECRET is empty.
	ErrJWTPanelSecretMissing = errors.New("config: JWT_PANEL_SECRET is required")
)

const defaultShutdownTimeout = 10 * time.Second

// Load reads configuration from the environment. HTTP_ADDR and
// DATABASE_DSN have defaults; JWT secrets must be provided explicitly.
func Load() (Config, error) {
	cfg := Config{
		HTTPAddr:        envOr("HTTP_ADDR", ":3000"),
		DatabaseDSN:     os.Getenv("DATABASE_DSN"),
		ShutdownTimeout: defaultShutdownTimeout,
	}

	if cfg.DatabaseDSN == "" {
		return Config{}, ErrDatabaseDSNMissing
	}

	cfg.JWTAdminSecret = os.Getenv("JWT_ADMIN_SECRET")
	if cfg.JWTAdminSecret == "" {
		return Config{}, ErrJWTAdminSecretMissing
	}

	cfg.JWTPanelSecret = os.Getenv("JWT_PANEL_SECRET")
	if cfg.JWTPanelSecret == "" {
		return Config{}, ErrJWTPanelSecretMissing
	}

	if v := os.Getenv("SHUTDOWN_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			cfg.ShutdownTimeout = d
		}
	}

	return cfg, nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}