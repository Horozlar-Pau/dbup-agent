package daemon

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/Horozlar-Pau/dbup-agent/internal/backend"
	"github.com/Horozlar-Pau/dbup-agent/internal/events"
	"github.com/Horozlar-Pau/dbup-agent/internal/inbox"
)

// Backend is what dbupd needs from dbup-backend.
type Backend interface {
	SendEvents(ctx context.Context, list []events.Event) error
	Reply(ctx context.Context, nudgeID, reply string) error
	VIPSeen(ctx context.Context, vipID string) error
}

const (
	// outboxSize bounds how many unsent events are kept while offline.
	outboxSize  = 500
	sendRetry   = 10 * time.Second
	callTimeout = 15 * time.Second
	cheerNote   = "🎉"
)

// SetBackend connects dbupd to dbup-backend. Call RunSender to start sending.
func (d *Daemon) SetBackend(b Backend) {
	d.backend = b
	d.outbox = make(chan events.Event, outboxSize)
}

// enqueue hands an event to the sender without blocking the caller.
func (d *Daemon) enqueue(e events.Event) {
	if d.outbox == nil {
		return
	}
	select {
	case d.outbox <- e:
	default:
		d.logf("backend outbox full, event %s dropped", e.Type)
	}
}

// RunSender sends queued events in batches until ctx ends. Failed batches are
// retried, so events made while offline arrive later.
func (d *Daemon) RunSender(ctx context.Context) {
	var pending []events.Event
	for {
		if len(pending) == 0 {
			select {
			case <-ctx.Done():
				return
			case e := <-d.outbox:
				pending = append(pending, e)
			}
		}
		pending = drain(d.outbox, pending)
		batch := pending[:min(len(pending), backend.MaxBatch)]
		cctx, cancel := context.WithTimeout(ctx, callTimeout)
		err := d.backend.SendEvents(cctx, batch)
		cancel()
		if err == nil {
			pending = append([]events.Event(nil), pending[len(batch):]...)
			continue
		}
		d.logf("send %d events: %v", len(batch), err)
		select {
		case <-ctx.Done():
			return
		case <-time.After(sendRetry):
		}
	}
}

// drain moves whatever is waiting in ch into pending, keeping at most outboxSize.
func drain(ch chan events.Event, pending []events.Event) []events.Event {
	for {
		select {
		case e := <-ch:
			pending = append(pending, e)
			if over := len(pending) - outboxSize; over > 0 {
				pending = append([]events.Event(nil), pending[over:]...)
			}
		default:
			return pending
		}
	}
}

// HandleLive turns a message from the backend into a nudge or an alert.
func (d *Daemon) HandleLive(m backend.Message) {
	switch m.Type {
	case backend.MsgNudges:
		var data backend.NudgesData
		if err := json.Unmarshal(m.Data, &data); err != nil {
			d.logf("bad %s: %v", m.Type, err)
			return
		}
		for _, n := range data.Nudges {
			kind := n.Label
			if kind == "" {
				kind = n.Kind
			}
			if _, err := d.inbox.Add(inbox.NewNudge{From: n.From, Kind: kind, Note: n.Note, RemoteID: n.ID}); err != nil {
				d.logf("nudge %s: %v", n.ID, err)
			}
		}
	case backend.MsgVIP:
		var v backend.VIPData
		if err := json.Unmarshal(m.Data, &v); err != nil {
			d.logf("bad %s: %v", m.Type, err)
			return
		}
		if _, err := d.alerts.Add(inbox.NewAlert{Kind: inbox.VIP, From: v.From, Note: v.Note, RemoteID: v.ID}); err != nil {
			d.logf("vip %s: %v", v.ID, err)
		}
	case backend.MsgCheer:
		var c backend.CheerData
		if err := json.Unmarshal(m.Data, &c); err != nil {
			d.logf("bad %s: %v", m.Type, err)
			return
		}
		note := strings.TrimSpace(c.Message)
		if note == "" && c.Confetti {
			note = cheerNote
		}
		if _, err := d.alerts.Add(inbox.NewAlert{Kind: inbox.Moral, From: c.From, Note: note}); err != nil {
			d.logf("cheer: %v", err)
		}
	}
}

// sendReply tells the backend about a quick reply, in the background.
func (d *Daemon) sendReply(n inbox.Nudge) {
	if d.backend == nil || n.RemoteID == "" || n.Reply == "" {
		return
	}
	d.async(func() {
		ctx, cancel := context.WithTimeout(context.Background(), callTimeout)
		defer cancel()
		if err := d.backend.Reply(ctx, n.RemoteID, n.Reply); err != nil {
			d.logf("reply to nudge %s: %v", n.RemoteID, err)
		}
	})
}

// sendVIPSeen tells the backend the developer pressed "Gördüm", in the background.
func (d *Daemon) sendVIPSeen(a inbox.Alert) {
	if d.backend == nil || a.RemoteID == "" || a.Kind != inbox.VIP {
		return
	}
	d.async(func() {
		ctx, cancel := context.WithTimeout(context.Background(), callTimeout)
		defer cancel()
		if err := d.backend.VIPSeen(ctx, a.RemoteID); err != nil {
			d.logf("vip %s seen: %v", a.RemoteID, err)
		}
	})
}
