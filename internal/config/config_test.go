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

func TestSummaryEnabledByDefault(t *testing.T) {
	t.Setenv("DBUP_HOME", t.TempDir())
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Summary.IsEnabled() {
		t.Fatal("commit summary should be on by default")
	}
}

func TestSummarySettingsAreRead(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DBUP_HOME", dir)
	body := `{"summary":{"enabled":false,"cliproxy_url":"http://127.0.0.1:9999","model":"m"}}`
	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Summary.IsEnabled() || cfg.Summary.CLIProxyURL != "http://127.0.0.1:9999" || cfg.Summary.Model != "m" {
		t.Fatalf("got %+v", cfg.Summary)
	}
}

func TestSaveKeepsSettingsAndHidesFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DBUP_HOME", dir)
	off := false
	if err := Save(Config{BackendURL: "https://dbup.example.com", Token: "dbd_x", Summary: Summary{Enabled: &off}}); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Token != "dbd_x" || cfg.Summary.IsEnabled() {
		t.Fatalf("got %+v", cfg)
	}
	fi, err := os.Stat(filepath.Join(dir, "config.json"))
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode().Perm() != 0o600 {
		t.Fatalf("config.json mode = %v, want 0600", fi.Mode().Perm())
	}
}
