package config

import (
	"errors"
	"testing"
	"time"
)

func TestLoadRequiresDatabaseDSN(t *testing.T) {
	clearEnv(t)

	_, err := Load()
	if !errors.Is(err, ErrDatabaseDSNMissing) {
		t.Fatalf("expected ErrDatabaseDSNMissing, got %v", err)
	}
}

func TestLoadRequiresJWTSecrets(t *testing.T) {
	clearEnv(t)
	t.Setenv("DATABASE_DSN", "user:pass@tcp(127.0.0.1:3306)/db")

	_, err := Load()
	if !errors.Is(err, ErrJWTAdminSecretMissing) {
		t.Fatalf("expected ErrJWTAdminSecretMissing, got %v", err)
	}

	t.Setenv("JWT_ADMIN_SECRET", "admin-secret")
	_, err = Load()
	if !errors.Is(err, ErrJWTPanelSecretMissing) {
		t.Fatalf("expected ErrJWTPanelSecretMissing, got %v", err)
	}
}

func TestLoadAppliesDefaults(t *testing.T) {
	clearEnv(t)
	t.Setenv("DATABASE_DSN", "user:pass@tcp(127.0.0.1:3306)/db")
	t.Setenv("JWT_ADMIN_SECRET", "admin-secret")
	t.Setenv("JWT_PANEL_SECRET", "panel-secret")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.HTTPAddr != ":3000" {
		t.Errorf("HTTPAddr = %q, want %q", cfg.HTTPAddr, ":3000")
	}
	if cfg.ShutdownTimeout != 10*time.Second {
		t.Errorf("ShutdownTimeout = %v, want 10s", cfg.ShutdownTimeout)
	}
	if cfg.JWTAdminSecret != "admin-secret" || cfg.JWTPanelSecret != "panel-secret" {
		t.Errorf("secrets not loaded: %+v", cfg)
	}
}

func TestLoadOverrides(t *testing.T) {
	clearEnv(t)
	t.Setenv("HTTP_ADDR", ":8081")
	t.Setenv("DATABASE_DSN", "dsn")
	t.Setenv("JWT_ADMIN_SECRET", "a")
	t.Setenv("JWT_PANEL_SECRET", "b")
	t.Setenv("SHUTDOWN_TIMEOUT", "3s")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.HTTPAddr != ":8081" {
		t.Errorf("HTTPAddr = %q, want :8081", cfg.HTTPAddr)
	}
	if cfg.ShutdownTimeout != 3*time.Second {
		t.Errorf("ShutdownTimeout = %v, want 3s", cfg.ShutdownTimeout)
	}
}

func clearEnv(t *testing.T) {
	t.Helper()
	for _, k := range []string{"HTTP_ADDR", "DATABASE_DSN", "JWT_ADMIN_SECRET", "JWT_PANEL_SECRET", "SHUTDOWN_TIMEOUT"} {
		t.Setenv(k, "")
	}
}