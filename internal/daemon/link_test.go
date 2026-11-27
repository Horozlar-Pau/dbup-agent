package daemon

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/Horozlar-Pau/dbup-agent/internal/backend"
	"github.com/Horozlar-Pau/dbup-agent/internal/events"
	"github.com/Horozlar-Pau/dbup-agent/internal/inbox"
	"github.com/Horozlar-Pau/dbup-agent/internal/ipc"
)

type fakeBackend struct {
	mu      sync.Mutex
	sent    []events.Event
	failN   int
	replies map[string]string
	seen    []string
}

func (f *fakeBackend) SendEvents(_ context.Context, list []events.Event) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.failN > 0 {
		f.failN--
		return errors.New("offline")
	}
	f.sent = append(f.sent, list...)
	return nil
}

func (f *fakeBackend) Reply(_ context.Context, id, reply string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.replies[id] = reply
	return nil
}

func (f *fakeBackend) VIPSeen(_ context.Context, id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.seen = append(f.seen, id)
	return nil
}

func (f *fakeBackend) sentCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.sent)
}

func linkedDaemon() (*Daemon, *fakeBackend) {
	d := testDaemon(nil)
	b := &fakeBackend{replies: map[string]string{}}
	d.SetBackend(b)
	return d, b
}

func live(t *testing.T, typ string, data any) backend.Message {
	t.Helper()
	b, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}
	return backend.Message{Type: typ, Data: b}
}

func TestEventsAreSentToBackend(t *testing.T) {
	d, b := linkedDaemon()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go d.RunSender(ctx)

	sendEvent(t, d, commit(""))
	deadline := time.Now().Add(2 * time.Second)
	for b.sentCount() == 0 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if b.sentCount() != 1 {
		t.Fatalf("sent %d events, want 1", b.sentCount())
	}
}

func TestReleasedNudgeIsQueuedOnceAndReplyGoesBack(t *testing.T) {
	d, b := linkedDaemon()
	msg := live(t, backend.MsgNudges, backend.NudgesData{Nudges: []backend.Nudge{{ID: "n1", Kind: "coffee", Label: "Kahve molası?", From: "Ayşe"}}})
	d.HandleLive(msg)
	d.HandleLive(msg) // resent after reconnect

	q := d.inbox.Snapshot().Queued
	if len(q) != 1 || q[0].Kind != "Kahve molası?" || q[0].From != "Ayşe" || q[0].RemoteID != "n1" {
		t.Fatalf("queued = %+v", q)
	}
	data, _ := json.Marshal(NudgeAction{ID: q[0].ID, Reply: "10 dk'ya ararım"})
	if resp := d.Handle(ipc.Request{Type: TypeNudgeReply, Data: data}); !resp.OK {
		t.Fatalf("reply: %s", resp.Error)
	}
	if b.replies["n1"] != "10 dk'ya ararım" {
		t.Fatalf("backend replies = %v", b.replies)
	}
}

func TestVIPAndCheerBecomeAlerts(t *testing.T) {
	d, b := linkedDaemon()
	d.HandleLive(live(t, backend.MsgVIP, backend.VIPData{ID: "v1", From: "Ayşe", Note: "Kapıdayım"}))
	d.HandleLive(live(t, backend.MsgCheer, backend.CheerData{From: "Ayşe", Confetti: true}))

	pending := d.alerts.Pending()
	if len(pending) != 2 {
		t.Fatalf("pending = %+v", pending)
	}
	var vip inbox.Alert
	for _, a := range pending {
		if a.Kind == inbox.VIP {
			vip = a
		} else if a.Note != cheerNote {
			t.Fatalf("confetti cheer note = %q", a.Note)
		}
	}
	data, _ := json.Marshal(AlertAck{ID: vip.ID})
	if resp := d.Handle(ipc.Request{Type: TypeAlertAck, Data: data}); !resp.OK {
		t.Fatalf("ack: %s", resp.Error)
	}
	if len(b.seen) != 1 || b.seen[0] != "v1" {
		t.Fatalf("vip seen = %v", b.seen)
	}
}

func TestDrainKeepsNewestWhenFull(t *testing.T) {
	ch := make(chan events.Event, outboxSize+10)
	for i := range outboxSize + 10 {
		ch <- events.Event{Type: events.GitCommit, At: time.Unix(int64(i), 0)}
	}
	got := drain(ch, nil)
	if len(got) != outboxSize || got[len(got)-1].At.Unix() != outboxSize+9 {
		t.Fatalf("kept %d, last %v", len(got), got[len(got)-1].At)
	}
}
