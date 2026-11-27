// Package version holds the build version, set with -ldflags at release time.
package version

// Version is overwritten by the build: -ldflags "-X .../internal/version.Version=v1.2.3".
var Version = "dev"
