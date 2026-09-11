package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadValidatesBaseURL(t *testing.T) {
	t.Setenv("WP_BASE_URL", "http://example.com")
	t.Setenv("WP_USERNAME", "editor")
	t.Setenv("WP_APP_PASSWORD", "password")
	if _, err := Load(); err == nil {
		t.Fatal("Load() accepted insecure non-loopback URL")
	}
}

func TestLoadAcceptsLoopbackHTTP(t *testing.T) {
	t.Setenv("WP_BASE_URL", "http://127.0.0.1:8080/")
	t.Setenv("WP_USERNAME", "editor")
	t.Setenv("WP_APP_PASSWORD", "password")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load(): %v", err)
	}
	if got := cfg.BaseURL.String(); got != "http://127.0.0.1:8080" {
		t.Fatalf("BaseURL = %q", got)
	}
}

func TestLoadEnvFileDoesNotOverrideEnvironment(t *testing.T) {
	t.Setenv("WP_USERNAME", "from-environment")
	path := filepath.Join(t.TempDir(), "mcp.env")
	if err := os.WriteFile(path, []byte("WP_USERNAME=from-file\nWP_TIMEOUT=10s\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := LoadEnvFile(path); err != nil {
		t.Fatal(err)
	}
	if got := os.Getenv("WP_USERNAME"); got != "from-environment" {
		t.Fatalf("WP_USERNAME = %q", got)
	}
	if got := os.Getenv("WP_TIMEOUT"); got != "10s" {
		t.Fatalf("WP_TIMEOUT = %q", got)
	}
}
