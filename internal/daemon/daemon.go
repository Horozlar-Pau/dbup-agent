// Package daemon is the core of dbupd: it answers requests coming over the local socket.
package daemon

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/Horozlar-Pau/dbup-agent/internal/config"
	"github.com/Horozlar-Pau/dbup-agent/internal/events"
	"github.com/Horozlar-Pau/dbup-agent/internal/ipc"
	"github.com/Horozlar-Pau/dbup-agent/internal/version"
)

// Request types understood by dbupd.
const (
	TypeStatus = "status" // no data; answers Status
	TypeEvent  = "event"  // data: events.Event; answers nothing
	TypeEvents = "events" // data: EventsQuery; answers []events.Event
)

// recentEvents is how many events dbupd keeps in memory.
const recentEvents = 200

// Status is the answer to a status request.
type Status struct {
	Version       string    `json:"version"`
	StartedAt     time.Time `json:"started_at"`
	UptimeSeconds int64     `json:"uptime_seconds"`
	BackendURL    string    `json:"backend_url"`
	TokenSet      bool      `json:"token_set"`
	EventCount    int       `json:"event_count"`
}

// EventsQuery asks for the newest Limit events (0 = all kept).
type EventsQuery struct {
	Limit int `json:"limit"`
}

// Daemon keeps dbupd's state.
type Daemon struct {
	cfg       config.Config
	startedAt time.Time
	now       func() time.Time
	events    *events.Log
	logf      func(format string, args ...any)
}

// New creates a Daemon with the given settings.
func New(cfg config.Config) *Daemon {
	return &Daemon{
		cfg:       cfg,
		startedAt: time.Now(),
		now:       time.Now,
		events:    events.NewLog(recentEvents),
		logf:      log.Printf,
	}
}

// Handle answers one request from the socket.
func (d *Daemon) Handle(req ipc.Request) ipc.Response {
	switch req.Type {
	case TypeStatus:
		return ipc.OKResponse(d.status())
	case TypeEvent:
		return d.addEvent(req.Data)
	case TypeEvents:
		var q EventsQuery
		if len(req.Data) > 0 {
			if err := json.Unmarshal(req.Data, &q); err != nil {
				return ipc.ErrorResponse(fmt.Errorf("bad events query: %w", err))
			}
		}
		return ipc.OKResponse(d.events.Recent(q.Limit))
	default:
		return ipc.ErrorResponse(fmt.Errorf("unknown request type %q", req.Type))
	}
}

func (d *Daemon) addEvent(data json.RawMessage) ipc.Response {
	var e events.Event
	if err := json.Unmarshal(data, &e); err != nil {
		return ipc.ErrorResponse(fmt.Errorf("bad event: %w", err))
	}
	if err := e.Validate(); err != nil {
		return ipc.ErrorResponse(err)
	}
	d.events.Add(e)
	d.logf("event %s from %s", e.Type, e.Source)
	return ipc.Response{OK: true}
}

func (d *Daemon) status() Status {
	return Status{
		Version:       version.Version,
		StartedAt:     d.startedAt,
		UptimeSeconds: int64(d.now().Sub(d.startedAt).Seconds()),
		BackendURL:    d.cfg.BackendURL,
		TokenSet:      d.cfg.Token != "",
		EventCount:    len(d.events.Recent(0)),
	}
}
