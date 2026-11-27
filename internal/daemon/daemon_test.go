package daemon

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/Horozlar-Pau/dbup-agent/internal/config"
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
