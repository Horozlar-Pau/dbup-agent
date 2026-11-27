// Package daemon is the core of dbupd: it answers requests coming over the local socket.
package daemon

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/Horozlar-Pau/dbup-agent/internal/config"
	"github.com/Horozlar-Pau/dbup-agent/internal/events"
	"github.com/Horozlar-Pau/dbup-agent/internal/inbox"
	"github.com/Horozlar-Pau/dbup-agent/internal/ipc"
	"github.com/Horozlar-Pau/dbup-agent/internal/version"
)

// Request types understood by dbupd.
const (
	TypeStatus = "status" // no data; answers Status
	TypeEvent  = "event"  // data: EventRequest; answers nothing
	TypeEvents = "events" // data: EventsQuery; answers []events.Event

	TypeNudges       = "nudges"        // no data; answers inbox.Snapshot
	TypeNudgeAdd     = "nudge.add"     // data: inbox.NewNudge; answers inbox.Nudge (test command for now)
	TypeNudgeReply   = "nudge.reply"   // data: NudgeAction with Reply; answers inbox.Nudge
	TypeNudgeDeliver = "nudge.deliver" // data: NudgeAction; answers inbox.Nudge

	TypeAlerts   = "alerts"    // no data; answers unseen []inbox.Alert
	TypeAlertAdd = "alert.add" // data: inbox.NewAlert; answers inbox.Alert (test command for now)
	TypeAlertAck = "alert.ack" // data: AlertAck; answers inbox.Alert
)

// AlertAck marks an alert as seen, with an optional quick reply.
type AlertAck struct {
	ID    int64  `json:"id"`
	Reply string `json:"reply,omitempty"`
}

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
	// SummaryOn is true when commit summaries are on and CLIProxy is configured.
	SummaryOn   bool   `json:"summary_on"`
	CLIProxyURL string `json:"cliproxy_url,omitempty"`
}

// EventsQuery asks for the newest Limit events (0 = all kept).
type EventsQuery struct {
	Limit int `json:"limit"`
}

// NudgeAction points at one nudge, with the reply text for nudge.reply.
type NudgeAction struct {
	ID    int64  `json:"id"`
	Reply string `json:"reply,omitempty"`
}

// Daemon keeps dbupd's state.
type Daemon struct {
	cfg       config.Config
	startedAt time.Time
	now       func() time.Time
	events    *events.Log
	inbox     *inbox.Inbox
	alerts    *inbox.Alerts
	logf      func(format string, args ...any)

	summarizer Summarizer
	// async runs background work; tests replace it to run work inline.
	async func(func())
}

// New creates a Daemon with the given settings.
func New(cfg config.Config) *Daemon {
	return &Daemon{
		cfg:       cfg,
		startedAt: time.Now(),
		now:       time.Now,
		events:    events.NewLog(recentEvents),
		inbox:     inbox.New(),
		alerts:    inbox.NewAlerts(),
		logf:      log.Printf,
		async:     func(f func()) { go f() },
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
	case TypeNudges:
		return ipc.OKResponse(d.inbox.Snapshot())
	case TypeNudgeAdd:
		var n inbox.NewNudge
		if err := json.Unmarshal(req.Data, &n); err != nil {
			return ipc.ErrorResponse(fmt.Errorf("bad nudge: %w", err))
		}
		return result(d.inbox.Add(n))
	case TypeNudgeReply, TypeNudgeDeliver:
		var a NudgeAction
		if err := json.Unmarshal(req.Data, &a); err != nil {
			return ipc.ErrorResponse(fmt.Errorf("bad nudge action: %w", err))
		}
		if req.Type == TypeNudgeReply {
			return result(d.inbox.Reply(a.ID, a.Reply))
		}
		return result(d.inbox.Deliver(a.ID))
	case TypeAlerts:
		return ipc.OKResponse(d.alerts.Pending())
	case TypeAlertAdd:
		var n inbox.NewAlert
		if err := json.Unmarshal(req.Data, &n); err != nil {
			return ipc.ErrorResponse(fmt.Errorf("bad alert: %w", err))
		}
		return result(d.alerts.Add(n))
	case TypeAlertAck:
		var a AlertAck
		if err := json.Unmarshal(req.Data, &a); err != nil {
			return ipc.ErrorResponse(fmt.Errorf("bad alert ack: %w", err))
		}
		return result(d.alerts.Ack(a.ID, a.Reply))
	default:
		return ipc.ErrorResponse(fmt.Errorf("unknown request type %q", req.Type))
	}
}

func (d *Daemon) addEvent(data json.RawMessage) ipc.Response {
	var req EventRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return ipc.ErrorResponse(fmt.Errorf("bad event: %w", err))
	}
	e := req.Event
	// Only dbupd writes summaries, so the "commit özeti" setting cannot be bypassed.
	e.HumanSummary = ""
	if err := e.Validate(); err != nil {
		return ipc.ErrorResponse(err)
	}
	if d.wantsSummary(req) {
		// The hook must not wait for CLIProxy: summarize in the background.
		d.async(func() { d.store(d.withSummary(e, req.CommitSubject)) })
		return ipc.Response{OK: true}
	}
	d.store(e)
	return ipc.Response{OK: true}
}

func (d *Daemon) store(e events.Event) {
	d.events.Add(e)
	if e.HumanSummary != "" {
		d.logf("event %s from %s (with summary)", e.Type, e.Source)
		return
	}
	d.logf("event %s from %s", e.Type, e.Source)
}

func (d *Daemon) status() Status {
	return Status{
		Version:       version.Version,
		StartedAt:     d.startedAt,
		UptimeSeconds: int64(d.now().Sub(d.startedAt).Seconds()),
		BackendURL:    d.cfg.BackendURL,
		TokenSet:      d.cfg.Token != "",
		EventCount:    len(d.events.Recent(0)),
		SummaryOn:     d.summarizer != nil,
		CLIProxyURL:   d.cliproxyURL(),
	}
}

func result(v any, err error) ipc.Response {
	if err != nil {
		return ipc.ErrorResponse(err)
	}
	return ipc.OKResponse(v)
}
