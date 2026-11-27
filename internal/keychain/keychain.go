// Package keychain reads secrets from the macOS Keychain.
package keychain

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Account is the Keychain account name dbup stores its items under.
const Account = "dbup"

// CLIProxyService is the Keychain item holding CLIProxy's local access key.
const CLIProxyService = "dbup-cliproxy"

// ErrNotFound means the Keychain has no such item.
var ErrNotFound = errors.New("keychain item not found")

const (
	securityBin = "/usr/bin/security"
	readTimeout = 5 * time.Second
	// itemNotFound is the exit code `security` uses for a missing item.
	itemNotFound = 44
)

// Get returns the password stored for service under Account.
func Get(ctx context.Context, service string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, readTimeout)
	defer cancel()
	var stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, securityBin, "find-generic-password", "-a", Account, "-s", service, "-w")
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == itemNotFound {
		return "", ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("read keychain item %s: %w (%s)", service, err, strings.TrimSpace(stderr.String()))
	}
	return strings.TrimRight(string(out), "\n"), nil
}
