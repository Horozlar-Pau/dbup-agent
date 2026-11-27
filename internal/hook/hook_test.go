package hook

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Horozlar-Pau/dbup-agent/internal/events"
)

var now = time.Date(2026, 11, 27, 14, 32, 0, 0, time.UTC)

func TestClaudeHooks(t *testing.T) {
	cases := map[string]string{
		"UserPromptSubmit": events.PromptStarted,
		"Stop":             events.PromptFinished,
		"SessionEnd":       events.SessionEnded,
	}
	for hookName, want := range cases {
		in := `{"session_id":"abc","cwd":"/Users/x/secret-project","hook_event_name":"` + hookName + `"}`
		e, err := Event(Claude, strings.NewReader(in), now)
		if err != nil {
			t.Fatalf("%s: %v", hookName, err)
		}
		if e.Type != want || e.Source != events.SourceClaude || !e.At.Equal(now) {
			t.Fatalf("%s: got %+v", hookName, e)
		}
	}
}

func TestClaudePromptTextIsNotKept(t *testing.T) {
	in := `{"hook_event_name":"UserPromptSubmit","prompt":"fix the bug in payments.go","cwd":"/Users/x/odeme-servisi"}`
	e, err := Event(Claude, strings.NewReader(in), now)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := json.Marshal(e)
	for _, secret := range []string{"payments.go", "fix the bug", "odeme-servisi"} {
		if strings.Contains(string(b), secret) {
			t.Fatalf("event leaks %q: %s", secret, b)
		}
	}
}

func TestUntrackedClaudeHookIsIgnored(t *testing.T) {
	_, err := Event(Claude, strings.NewReader(`{"hook_event_name":"PreToolUse"}`), now)
	if !errors.Is(err, ErrIgnored) {
		t.Fatalf("err = %v, want ErrIgnored", err)
	}
}

func TestClaudeBadInput(t *testing.T) {
	if _, err := Event(Claude, strings.NewReader("not json"), now); err == nil {
		t.Fatal("want error for bad input")
	}
	if _, err := Event(Claude, nil, now); !errors.Is(err, ErrIgnored) {
		t.Fatalf("no stdin: err = %v, want ErrIgnored", err)
	}
}

func TestGitCommit(t *testing.T) {
	e, err := Event(GitCommit, nil, now)
	if err != nil {
		t.Fatal(err)
	}
	if e.Type != events.GitCommit || e.Source != events.SourceGit {
		t.Fatalf("got %+v", e)
	}
}

func TestUnknownHook(t *testing.T) {
	if _, err := Event("nope", nil, now); err == nil {
		t.Fatal("want error for unknown hook")
	}
}
