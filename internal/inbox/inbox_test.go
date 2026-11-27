package inbox

import (
	"errors"
	"testing"
)

func TestAddReplyDeliver(t *testing.T) {
	b := New()
	a, err := b.Add(NewNudge{From: "Elif", Kind: "Kahve molası?", Note: " erken çık "})
	if err != nil {
		t.Fatal(err)
	}
	c, _ := b.Add(NewNudge{From: "Can", Kind: "Müsait olunca ara"})
	if a.Note != "erken çık" || a.State != Queued {
		t.Fatalf("got %+v", a)
	}

	s := b.Snapshot()
	if len(s.Queued) != 2 || s.Queued[0].ID != c.ID {
		t.Fatalf("queued should be newest first: %+v", s.Queued)
	}

	if _, err := b.Reply(a.ID, "Akşam konuşalım"); err != nil {
		t.Fatal(err)
	}
	if _, err := b.Deliver(c.ID); err != nil {
		t.Fatal(err)
	}
	s = b.Snapshot()
	if len(s.Queued) != 0 || len(s.History) != 2 {
		t.Fatalf("got %+v", s)
	}
	if s.History[1].State != Replied || s.History[1].Reply != "Akşam konuşalım" || s.History[0].State != Delivered {
		t.Fatalf("history: %+v", s.History)
	}

	if _, err := b.Reply(a.ID, "again"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("second reply should fail, got %v", err)
	}
}

func TestAddNeedsSenderAndKind(t *testing.T) {
	if _, err := New().Add(NewNudge{From: " ", Kind: "x"}); err == nil {
		t.Fatal("expected error")
	}
}

func TestHistoryIsCapped(t *testing.T) {
	b := New()
	for i := 0; i < maxHistory+5; i++ {
		n, _ := b.Add(NewNudge{From: "a", Kind: "b"})
		b.Deliver(n.ID)
	}
	if got := len(b.Snapshot().History); got != maxHistory {
		t.Fatalf("history len %d", got)
	}
}

func TestAddIgnoresResentRemoteNudge(t *testing.T) {
	b := New()
	first, err := b.Add(NewNudge{From: "Elif", Kind: "Kahve molası?", RemoteID: "abc"})
	if err != nil {
		t.Fatal(err)
	}
	again, err := b.Add(NewNudge{From: "Elif", Kind: "Kahve molası?", RemoteID: "abc"})
	if err != nil {
		t.Fatal(err)
	}
	if again.ID != first.ID || len(b.Snapshot().Queued) != 1 {
		t.Fatalf("resent nudge was added twice: %+v", b.Snapshot())
	}
	if _, err := b.Reply(first.ID, "Geliyorum"); err != nil {
		t.Fatal(err)
	}
	if n, _ := b.Add(NewNudge{From: "Elif", Kind: "Kahve molası?", RemoteID: "abc"}); n.State != Replied {
		t.Fatalf("answered nudge came back as %q", n.State)
	}
}
