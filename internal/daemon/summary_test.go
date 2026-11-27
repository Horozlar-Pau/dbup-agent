package daemon

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Horozlar-Pau/dbup-agent/internal/config"
	"github.com/Horozlar-Pau/dbup-agent/internal/events"
	"github.com/Horozlar-Pau/dbup-agent/internal/ipc"
)

type fakeSummarizer struct {
	text string
	err  error
	got  []string
}

func (f *fakeSummarizer) Summarize(_ context.Context, msg string) (string, error) {
	f.got = append(f.got, msg)
	return f.text, f.err
}

func testDaemon(s Summarizer) *Daemon {
	d := New(config.Config{})
	d.logf = func(string, ...any) {}
	d.async = func(f func()) { f() }
	if s != nil {
		d.SetSummarizer(s)
	}
	return d
}

func sendEvent(t *testing.T, d *Daemon, req EventRequest) {
	t.Helper()
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp := d.Handle(ipc.Request{Type: TypeEvent, Data: data}); !resp.OK {
		t.Fatalf("event rejected: %s", resp.Error)
	}
}

func lastEvent(t *testing.T, d *Daemon) events.Event {
	t.Helper()
	list := d.events.Recent(1)
	if len(list) != 1 {
		t.Fatalf("got %d events, want 1", len(list))
	}
	return list[0]
}

func commit(subject string) EventRequest {
	return EventRequest{
		Event:         events.Event{Type: events.GitCommit, At: time.Now(), Source: events.SourceGit},
		CommitSubject: subject,
	}
}

func TestCommitGetsSummary(t *testing.T) {
	f := &fakeSummarizer{text: "Uygulamadaki bir hatayı düzeltti."}
	d := testDaemon(f)

	sendEvent(t, d, commit("fix: login crash"))

	if e := lastEvent(t, d); e.HumanSummary != "Uygulamadaki bir hatayı düzeltti." {
		t.Fatalf("got %+v", e)
	}
	if len(f.got) != 1 || f.got[0] != "fix: login crash" {
		t.Fatalf("summarizer got %v", f.got)
	}
}

func TestCommitMessageIsNotStored(t *testing.T) {
	d := testDaemon(&fakeSummarizer{text: "Bir şey yaptı."})
	sendEvent(t, d, commit("feat: secret-project payments"))

	resp := d.Handle(ipc.Request{Type: TypeEvents})
	if strings.Contains(string(resp.Data), "secret-project") {
		t.Fatalf("commit subject leaked into events: %s", resp.Data)
	}
}

func TestCommitWithoutSummaryWhenProxyFails(t *testing.T) {
	d := testDaemon(&fakeSummarizer{err: errors.New("connection refused")})
	sendEvent(t, d, commit("fix: x"))

	if e := lastEvent(t, d); e.Type != events.GitCommit || e.HumanSummary != "" {
		t.Fatalf("got %+v", e)
	}
}

func TestSummaryOffDoesNotCallProxy(t *testing.T) {
	d := testDaemon(nil)
	sendEvent(t, d, commit("fix: x"))

	if e := lastEvent(t, d); e.HumanSummary != "" {
		t.Fatalf("got %+v", e)
	}
}

func TestClientCannotSetSummary(t *testing.T) {
	d := testDaemon(nil)
	req := commit("")
	req.HumanSummary = "injected"
	sendEvent(t, d, req)

	if e := lastEvent(t, d); e.HumanSummary != "" {
		t.Fatalf("client-sent summary kept: %+v", e)
	}
}

func TestOverlongSummaryIsDropped(t *testing.T) {
	d := testDaemon(&fakeSummarizer{text: strings.Repeat("ş", events.MaxHumanSummary+1)})
	sendEvent(t, d, commit("fix: x"))

	if e := lastEvent(t, d); e.HumanSummary != "" {
		t.Fatalf("overlong summary kept: %+v", e)
	}
}

func TestStatusShowsSummary(t *testing.T) {
	d := testDaemon(&fakeSummarizer{})
	if st := d.status(); !st.SummaryOn || st.CLIProxyURL == "" {
		t.Fatalf("got %+v", st)
	}
	if st := testDaemon(nil).status(); st.SummaryOn || st.CLIProxyURL != "" {
		t.Fatalf("got %+v", st)
	}
}
