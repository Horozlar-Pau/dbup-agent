package daemon

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/Horozlar-Pau/dbup-agent/internal/config"
	"github.com/Horozlar-Pau/dbup-agent/internal/events"
	"github.com/Horozlar-Pau/dbup-agent/internal/ipc"
)

func TestStatus(t *testing.T) {
	d := New(config.Config{BackendURL: "https://dbup.example.com", Token: "abc"})
	d.now = func() time.Time { return d.startedAt.Add(90 * time.Second) }

	resp := d.Handle(ipc.Request{Type: TypeStatus})
	if !resp.OK {
		t.Fatalf("status failed: %s", resp.Error)
	}
	var st Status
	if err := json.Unmarshal(resp.Data, &st); err != nil {
		t.Fatal(err)
	}
	if st.BackendURL != "https://dbup.example.com" || !st.TokenSet || st.UptimeSeconds != 90 {
		t.Fatalf("got %+v", st)
	}
}

func TestStatusDoesNotLeakToken(t *testing.T) {
	d := New(config.Config{Token: "secret-token"})
	resp := d.Handle(ipc.Request{Type: TypeStatus})
	if strings.Contains(string(resp.Data), "secret-token") {
		t.Fatal("status response must not include the token")
	}
}

func TestUnknownType(t *testing.T) {
	resp := New(config.Config{}).Handle(ipc.Request{Type: "nope"})
	if resp.OK {
		t.Fatal("unknown request type should fail")
	}
}

func TestEventIsStoredAndListed(t *testing.T) {
	d := New(config.Config{})
	d.logf = func(string, ...any) {}

	data, _ := json.Marshal(events.Event{Type: events.GitCommit, At: time.Now(), Source: events.SourceGit})
	if resp := d.Handle(ipc.Request{Type: TypeEvent, Data: data}); !resp.OK {
		t.Fatalf("event rejected: %s", resp.Error)
	}

	resp := d.Handle(ipc.Request{Type: TypeEvents})
	var list []events.Event
	if err := json.Unmarshal(resp.Data, &list); err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].Type != events.GitCommit {
		t.Fatalf("got %+v", list)
	}
}

func TestInvalidEventRejected(t *testing.T) {
	d := New(config.Config{})
	d.logf = func(string, ...any) {}

	data, _ := json.Marshal(events.Event{Type: "made.up", At: time.Now()})
	if resp := d.Handle(ipc.Request{Type: TypeEvent, Data: data}); resp.OK {
		t.Fatal("invalid event accepted")
	}
	if resp := d.Handle(ipc.Request{Type: TypeEvent, Data: []byte("{")}); resp.OK {
		t.Fatal("malformed event accepted")
	}
}

func TestNudgeFlow(t *testing.T) {
	d := New(config.Config{})
	resp := d.Handle(ipc.Request{Type: TypeNudgeAdd, Data: json.RawMessage(`{"from":"Elif","kind":"Kahve molası?"}`)})
	if !resp.OK {
		t.Fatalf("add: %s", resp.Error)
	}
	resp = d.Handle(ipc.Request{Type: TypeNudgeReply, Data: json.RawMessage(`{"id":1,"reply":"10 dk'ya ararım"}`)})
	if !resp.OK {
		t.Fatalf("reply: %s", resp.Error)
	}
	resp = d.Handle(ipc.Request{Type: TypeNudges})
	if !strings.Contains(string(resp.Data), `"state":"replied"`) {
		t.Fatalf("snapshot: %s", resp.Data)
	}
	if resp := d.Handle(ipc.Request{Type: TypeNudgeDeliver, Data: json.RawMessage(`{"id":1}`)}); resp.OK {
		t.Fatal("handled nudge should not be delivered again")
	}
}
