package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMissingFileReturnsDefaults(t *testing.T) {
	t.Setenv("DBUP_HOME", t.TempDir())

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.BackendURL != DefaultBackendURL || cfg.Token != "" {
		t.Fatalf("got %+v, want defaults", cfg)
	}
}

func TestLoadReadsFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DBUP_HOME", dir)
	body := `{"backend_url":"https://dbup.example.com","token":"abc"}`
	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.BackendURL != "https://dbup.example.com" || cfg.Token != "abc" {
		t.Fatalf("got %+v", cfg)
	}
}

func TestLoadEmptyBackendFallsBackToDefault(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DBUP_HOME", dir)
	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte(`{"token":"abc"}`), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.BackendURL != DefaultBackendURL {
		t.Fatalf("BackendURL = %q, want default", cfg.BackendURL)
	}
}

func TestLoadBadJSON(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DBUP_HOME", dir)
	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte(`{`), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := Load(); err == nil {
		t.Fatal("want error for malformed config")
	}
}
