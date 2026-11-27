package inbox

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

// Alert kinds.
const (
	VIP   = "vip"   // full screen with sound, until the developer presses "Gördüm"
	Moral = "moral" // short confetti and message
)

// maxAlerts is how many alerts are kept, seen or not.
const maxAlerts = 50

// Alert is an urgent (VIP) or cheering (moral) message from the close circle.
type Alert struct {
	ID     int64     `json:"id"`
	Kind   string    `json:"kind"`
	From   string    `json:"from"`
	Note   string    `json:"note,omitempty"`
	At     time.Time `json:"at"`
	Seen   bool      `json:"seen"`
	SeenAt time.Time `json:"seen_at,omitzero"`
	Reply  string    `json:"reply,omitempty"`
}

// NewAlert is the input for AddAlert.
type NewAlert struct {
	Kind string `json:"kind"`
	From string `json:"from"`
	Note string `json:"note"`
}

// ErrAlertNotFound is returned for an unknown or already seen alert id.
var ErrAlertNotFound = errors.New("alert not found")

// Alerts is safe for concurrent use.
type Alerts struct {
	mu     sync.Mutex
	now    func() time.Time
	nextID int64
	list   []Alert
}

// NewAlerts returns an empty alert list.
func NewAlerts() *Alerts {
	return &Alerts{now: time.Now, nextID: 1}
}

// Add stores a new, unseen alert.
func (a *Alerts) Add(n NewAlert) (Alert, error) {
	n.From = strings.TrimSpace(n.From)
	if n.Kind != VIP && n.Kind != Moral {
		return Alert{}, fmt.Errorf("unknown alert kind %q", n.Kind)
	}
	if n.From == "" {
		return Alert{}, fmt.Errorf("alert needs a sender")
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	al := Alert{ID: a.nextID, Kind: n.Kind, From: n.From, Note: strings.TrimSpace(n.Note), At: a.now()}
	a.nextID++
	a.list = append(a.list, al)
	if over := len(a.list) - maxAlerts; over > 0 {
		a.list = append([]Alert(nil), a.list[over:]...)
	}
	return al, nil
}

// Pending returns unseen alerts, oldest first.
func (a *Alerts) Pending() []Alert {
	a.mu.Lock()
	defer a.mu.Unlock()
	out := []Alert{}
	for _, al := range a.list {
		if !al.Seen {
			out = append(out, al)
		}
	}
	return out
}

// Ack marks an alert as seen, with an optional quick reply.
// The sender will see this once the backend is connected (DBUP-34).
func (a *Alerts) Ack(id int64, reply string) (Alert, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	for i := range a.list {
		if a.list[i].ID == id && !a.list[i].Seen {
			a.list[i].Seen = true
			a.list[i].SeenAt = a.now()
			a.list[i].Reply = strings.TrimSpace(reply)
			return a.list[i], nil
		}
	}
	return Alert{}, ErrAlertNotFound
}
