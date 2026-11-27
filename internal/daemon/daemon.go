// Package daemon is the core of dbupd: it answers requests coming over the local socket.
package daemon

import (
	"fmt"
	"time"

	"github.com/Horozlar-Pau/dbup-agent/internal/config"
	"github.com/Horozlar-Pau/dbup-agent/internal/ipc"
	"github.com/Horozlar-Pau/dbup-agent/internal/version"
)

// Request types understood by dbupd.
const (
	TypeStatus = "status"
)

// Status is the answer to a status request.
type Status struct {
	Version       string    `json:"version"`
	StartedAt     time.Time `json:"started_at"`
	UptimeSeconds int64     `json:"uptime_seconds"`
	BackendURL    string    `json:"backend_url"`
	TokenSet      bool      `json:"token_set"`
}

// Daemon keeps dbupd's state.
type Daemon struct {
	cfg       config.Config
	startedAt time.Time
	now       func() time.Time
}

// New creates a Daemon with the given settings.
func New(cfg config.Config) *Daemon {
	return &Daemon{cfg: cfg, startedAt: time.Now(), now: time.Now}
}

// Handle answers one request from the socket.
func (d *Daemon) Handle(req ipc.Request) ipc.Response {
	switch req.Type {
	case TypeStatus:
		return ipc.OKResponse(d.status())
	default:
		return ipc.ErrorResponse(fmt.Errorf("unknown request type %q", req.Type))
	}
}

func (d *Daemon) status() Status {
	return Status{
		Version:       version.Version,
		StartedAt:     d.startedAt,
		UptimeSeconds: int64(d.now().Sub(d.startedAt).Seconds()),
		BackendURL:    d.cfg.BackendURL,
		TokenSet:      d.cfg.Token != "",
	}
}
