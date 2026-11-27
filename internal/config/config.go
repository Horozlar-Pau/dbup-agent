// Package config reads the agent's settings from the dbup home directory.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// DefaultBackendURL is used when the config file does not set one.
const DefaultBackendURL = "https://dbup.ardadagdeviren.com.tr"

// Config holds the settings shared by dbupd and the dbup CLI.
type Config struct {
	// BackendURL is the base address of dbup-backend.
	BackendURL string `json:"backend_url"`
	// Token is the device token from `dbup pair`. config.json is saved 0600.
	// Temporary: it will move to the macOS Keychain.
	Token string `json:"token"`
	// Summary controls the "Ne Yapıyor Bu?" commit sentence (docs/summary.md).
	Summary Summary `json:"summary"`
}

// Summary holds the commit summary settings. The CLIProxy access key is not
// here: it lives in the macOS Keychain.
type Summary struct {
	// Enabled is nil when unset; use IsEnabled.
	Enabled *bool `json:"enabled,omitempty"`
	// CLIProxyURL is CLIProxy's address. Empty means summary.DefaultURL.
	CLIProxyURL string `json:"cliproxy_url,omitempty"`
	// Model is the Claude model asked. Empty means summary.DefaultModel.
	Model string `json:"model,omitempty"`
}

// IsEnabled reports whether commit summaries are on. They are on by default.
func (s Summary) IsEnabled() bool {
	return s.Enabled == nil || *s.Enabled
}

// Dir returns the dbup home directory: $DBUP_HOME, or ~/.dbup.
func Dir() (string, error) {
	if d := os.Getenv("DBUP_HOME"); d != "" {
		return d, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("find home directory: %w", err)
	}
	return filepath.Join(home, ".dbup"), nil
}

// Path returns the location of config.json.
func Path() (string, error) {
	d, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, "config.json"), nil
}

// SocketPath returns the Unix socket dbupd listens on.
func SocketPath() (string, error) {
	d, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, "dbupd.sock"), nil
}

// Load reads config.json. A missing file is not an error: defaults are returned.
func Load() (Config, error) {
	cfg := Config{BackendURL: DefaultBackendURL}
	p, err := Path()
	if err != nil {
		return cfg, err
	}
	b, err := os.ReadFile(p)
	if errors.Is(err, os.ErrNotExist) {
		return cfg, nil
	}
	if err != nil {
		return cfg, fmt.Errorf("read config: %w", err)
	}
	if err := json.Unmarshal(b, &cfg); err != nil {
		return cfg, fmt.Errorf("parse %s: %w", p, err)
	}
	if cfg.BackendURL == "" {
		cfg.BackendURL = DefaultBackendURL
	}
	return cfg, nil
}

// EnsureDir creates the dbup home directory, readable only by the user.
func EnsureDir() (string, error) {
	d, err := Dir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(d, 0o700); err != nil {
		return "", fmt.Errorf("create %s: %w", d, err)
	}
	return d, nil
}

// Save writes cfg to config.json, readable only by the user (it holds the device token).
func Save(cfg Config) error {
	if _, err := EnsureDir(); err != nil {
		return err
	}
	p, err := Path()
	if err != nil {
		return err
	}
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, append(b, '\n'), 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	if err := os.Rename(tmp, p); err != nil {
		return fmt.Errorf("save config: %w", err)
	}
	return nil
}
