package config

import (
	"os"
	"testing"
)

func TestLoad_requiresDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	_, err := Load()
	if err == nil {
		t.Fatal("expected error when DATABASE_URL is empty")
	}
}

func TestLoad_defaults(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost/test?sslmode=disable")
	t.Setenv("HTTP_PORT", "")
	t.Setenv("CHAIN_ID", "31337")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.HTTPPort != "8080" {
		t.Fatalf("HTTPPort = %q, want 8080", cfg.HTTPPort)
	}
	if cfg.ChainID != 31337 {
		t.Fatalf("ChainID = %d, want 31337", cfg.ChainID)
	}
}

func TestLoad_customPort(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost/test?sslmode=disable")
	t.Setenv("HTTP_PORT", "9090")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.HTTPPort != "9090" {
		t.Fatalf("HTTPPort = %q, want 9090", cfg.HTTPPort)
	}

	_ = os.Unsetenv("HTTP_PORT")
}
