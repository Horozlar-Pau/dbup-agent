package daemon

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/Horozlar-Pau/dbup-agent/internal/backend"
	"github.com/Horozlar-Pau/dbup-agent/internal/ipc"
)

type fakeMessenger struct {
	to, text string
}

func (f *fakeMessenger) Circle(context.Context) ([]backend.Member, error) {
	return []backend.Member{{ID: "m1", DisplayName: "Elif", Username: "elif"}}, nil
}

func (f *fakeMessenger) SendMessage(_ context.Context, memberID, text string) (backend.SentMessage, error) {
	f.to, f.text = memberID, text
	return backend.SentMessage{ID: "x", Text: text}, nil
}

func TestMessageSendGoesToBackend(t *testing.T) {
	d := testDaemon(nil)
	f := &fakeMessenger{}
	d.SetMessenger(f)

	resp := d.Handle(ipc.Request{Type: TypeCircle})
	if !resp.OK || !strings.Contains(string(resp.Data), "Elif") {
		t.Fatalf("circle: %+v", resp)
	}
	data, _ := json.Marshal(MessageSend{MemberID: "m1", Text: "Akşam görüşürüz"})
	if resp := d.Handle(ipc.Request{Type: TypeMessageSend, Data: data}); !resp.OK {
		t.Fatalf("send: %s", resp.Error)
	}
	if f.to != "m1" || f.text != "Akşam görüşürüz" {
		t.Fatalf("sent %q to %q", f.text, f.to)
	}
}

func TestMessageNeedsPairing(t *testing.T) {
	resp := testDaemon(nil).Handle(ipc.Request{Type: TypeCircle})
	if resp.OK || !strings.Contains(resp.Error, "dbup pair") {
		t.Fatalf("got %+v", resp)
	}
}
