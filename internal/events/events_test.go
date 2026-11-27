package events

import (
	"testing"
	"time"
)

func ev(typ string, sec int) Event {
	return Event{Type: typ, At: time.Unix(int64(sec), 0), Source: SourceClaude}
}

func TestValidate(t *testing.T) {
	if err := ev(PromptStarted, 1).Validate(); err != nil {
		t.Fatalf("valid event rejected: %v", err)
	}
	if err := ev("ai.unknown", 1).Validate(); err == nil {
		t.Fatal("unknown type accepted")
	}
	if err := (Event{Type: GitCommit}).Validate(); err == nil {
		t.Fatal("event without time accepted")
	}
}

func TestLogKeepsNewest(t *testing.T) {
	l := NewLog(3)
	for i := 1; i <= 5; i++ {
		l.Add(ev(GitCommit, i))
	}
	all := l.Recent(0)
	if len(all) != 3 || all[0].At.Unix() != 3 || all[2].At.Unix() != 5 {
		t.Fatalf("got %+v", all)
	}
	last := l.Recent(2)
	if len(last) != 2 || last[0].At.Unix() != 4 {
		t.Fatalf("Recent(2) = %+v", last)
	}
}

func TestRecentReturnsCopy(t *testing.T) {
	l := NewLog(2)
	l.Add(ev(GitCommit, 1))
	got := l.Recent(0)
	got[0].Type = "changed"
	if l.Recent(0)[0].Type != GitCommit {
		t.Fatal("Recent must not expose internal slice")
	}
}
