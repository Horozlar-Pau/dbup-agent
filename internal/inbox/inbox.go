// Package inbox keeps the nudges sent by the close circle until the developer sees them.
//
// Nudges come from dbup-backend (or the `dbup test` command). A nudge from the
// backend carries its RemoteID so replies can be sent back and resends are ignored.
package inbox

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

// Nudge states.
const (
	Queued    = "queued"    // waiting for a good moment
	Delivered = "delivered" // shown, no reply
	Replied   = "replied"   // the developer picked a quick reply
)

// maxHistory is how many delivered or replied nudges are kept.
const maxHistory = 50

// Nudge is one message from someone in the close circle.
type Nudge struct {
	ID        int64     `json:"id"`
	RemoteID  string    `json:"remote_id,omitempty"`
	From      string    `json:"from"`
	Kind      string    `json:"kind"`
	Note      string    `json:"note,omitempty"`
	At        time.Time `json:"at"`
	State     string    `json:"state"`
	Reply     string    `json:"reply,omitempty"`
	RepliedAt time.Time `json:"replied_at,omitzero"`
}

// NewNudge is the input for Add.
type NewNudge struct {
	From     string `json:"from"`
	Kind     string `json:"kind"`
	Note     string `json:"note"`
	RemoteID string `json:"remote_id,omitempty"`
}

// Snapshot is the inbox content, newest first in both lists.
type Snapshot struct {
	Queued  []Nudge `json:"queued"`
	History []Nudge `json:"history"`
}

// ErrNotFound is returned for an unknown or already handled nudge id.
var ErrNotFound = errors.New("nudge not found")

// Inbox is safe for concurrent use.
type Inbox struct {
	mu      sync.Mutex
	now     func() time.Time
	nextID  int64
	queued  []Nudge
	history []Nudge
}

// New returns an empty inbox.
func New() *Inbox {
	return &Inbox{now: time.Now, nextID: 1}
}

// Add queues a nudge and returns it.
func (b *Inbox) Add(n NewNudge) (Nudge, error) {
	n.From = strings.TrimSpace(n.From)
	n.Kind = strings.TrimSpace(n.Kind)
	if n.From == "" || n.Kind == "" {
		return Nudge{}, fmt.Errorf("nudge needs a sender and a kind")
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if old, ok := b.findRemote(n.RemoteID); ok {
		return old, nil // the backend resends nudges after a reconnect
	}
	nd := Nudge{ID: b.nextID, RemoteID: n.RemoteID, From: n.From, Kind: n.Kind, Note: strings.TrimSpace(n.Note), At: b.now(), State: Queued}
	b.nextID++
	b.queued = append(b.queued, nd)
	return nd, nil
}

// findRemote returns the nudge with this backend id. Caller holds mu.
func (b *Inbox) findRemote(id string) (Nudge, bool) {
	if id == "" {
		return Nudge{}, false
	}
	for _, list := range [][]Nudge{b.queued, b.history} {
		for _, n := range list {
			if n.RemoteID == id {
				return n, true
			}
		}
	}
	return Nudge{}, false
}

// Snapshot returns a copy of the inbox.
func (b *Inbox) Snapshot() Snapshot {
	b.mu.Lock()
	defer b.mu.Unlock()
	return Snapshot{Queued: reversed(b.queued), History: reversed(b.history)}
}

// Reply marks a queued nudge as answered with text.
func (b *Inbox) Reply(id int64, text string) (Nudge, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return Nudge{}, fmt.Errorf("reply is empty")
	}
	return b.close(id, func(n *Nudge) {
		n.State = Replied
		n.Reply = text
		n.RepliedAt = b.now()
	})
}

// Deliver marks a queued nudge as seen without a reply.
func (b *Inbox) Deliver(id int64) (Nudge, error) {
	return b.close(id, func(n *Nudge) { n.State = Delivered })
}

func (b *Inbox) close(id int64, change func(*Nudge)) (Nudge, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for i, n := range b.queued {
		if n.ID != id {
			continue
		}
		change(&n)
		b.queued = append(b.queued[:i:i], b.queued[i+1:]...)
		b.history = append(b.history, n)
		if over := len(b.history) - maxHistory; over > 0 {
			b.history = append([]Nudge(nil), b.history[over:]...)
		}
		return n, nil
	}
	return Nudge{}, ErrNotFound
}

func reversed(in []Nudge) []Nudge {
	out := make([]Nudge, len(in))
	for i, n := range in {
		out[len(in)-1-i] = n
	}
	return out
}
