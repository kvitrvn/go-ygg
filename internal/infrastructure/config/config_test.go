package config

import (
	"os"
	"strings"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("GO_YGG_SERVER_HOST", "")
	t.Setenv("GO_YGG_SERVER_PORT", "")
	t.Setenv("GO_YGG_DATABASE_DSN", "")
	t.Setenv("GO_YGG_LOG_LEVEL", "")
	t.Setenv("GO_YGG_LOG_FORMAT", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Server.Host != "0.0.0.0" {
		t.Fatalf("Server.Host = %q, want %q", cfg.Server.Host, "0.0.0.0")
	}
	if cfg.Server.Port != 8080 {
		t.Fatalf("Server.Port = %d, want %d", cfg.Server.Port, 8080)
	}
	if cfg.Database.DSN != "" {
		t.Fatalf("Database.DSN = %q, want empty", cfg.Database.DSN)
	}
	if cfg.Log.Level != "info" {
		t.Fatalf("Log.Level = %q, want %q", cfg.Log.Level, "info")
	}
	if cfg.Log.Format != "json" {
		t.Fatalf("Log.Format = %q, want %q", cfg.Log.Format, "json")
	}
}

func TestLoadFromEnvironment(t *testing.T) {
	t.Setenv("GO_YGG_SERVER_HOST", "127.0.0.1")
	t.Setenv("GO_YGG_SERVER_PORT", "9090")
	t.Setenv("GO_YGG_DATABASE_DSN", "postgres://db")
	t.Setenv("GO_YGG_LOG_LEVEL", "debug")
	t.Setenv("GO_YGG_LOG_FORMAT", "text")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Server.Host != "127.0.0.1" {
		t.Fatalf("Server.Host = %q, want %q", cfg.Server.Host, "127.0.0.1")
	}
	if cfg.Server.Port != 9090 {
		t.Fatalf("Server.Port = %d, want %d", cfg.Server.Port, 9090)
	}
	if cfg.Database.DSN != "postgres://db" {
		t.Fatalf("Database.DSN = %q, want %q", cfg.Database.DSN, "postgres://db")
	}
	if cfg.Log.Level != "debug" {
		t.Fatalf("Log.Level = %q, want %q", cfg.Log.Level, "debug")
	}
	if cfg.Log.Format != "text" {
		t.Fatalf("Log.Format = %q, want %q", cfg.Log.Format, "text")
	}
}

func TestLoadInvalidPort(t *testing.T) {
	t.Setenv("GO_YGG_SERVER_PORT", "not-a-number")
	t.Setenv("GO_YGG_SERVER_HOST", "")
	t.Setenv("GO_YGG_DATABASE_DSN", "")
	t.Setenv("GO_YGG_LOG_LEVEL", "")
	t.Setenv("GO_YGG_LOG_FORMAT", "")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil, want non-nil")
	}
	if !strings.Contains(err.Error(), "GO_YGG_SERVER_PORT") {
		t.Fatalf("Load() error = %q, want mention of GO_YGG_SERVER_PORT", err)
	}
}

func TestMain(m *testing.M) {
	for _, key := range []string{
		"GO_YGG_SERVER_HOST",
		"GO_YGG_SERVER_PORT",
		"GO_YGG_DATABASE_DSN",
		"GO_YGG_LOG_LEVEL",
		"GO_YGG_LOG_FORMAT",
	} {
		_ = os.Unsetenv(key)
	}

	os.Exit(m.Run())
}
