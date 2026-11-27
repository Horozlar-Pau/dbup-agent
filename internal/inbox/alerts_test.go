package inbox

import (
	"errors"
	"testing"
)

func TestAlerts(t *testing.T) {
	a := NewAlerts()
	if _, err := a.Add(NewAlert{Kind: "nope", From: "Elif"}); err == nil {
		t.Fatal("unknown kind should fail")
	}
	v, err := a.Add(NewAlert{Kind: VIP, From: "Elif", Note: "Kapıda kaldım"})
	if err != nil {
		t.Fatal(err)
	}
	a.Add(NewAlert{Kind: Moral, From: "Can"})
	if got := len(a.Pending()); got != 2 {
		t.Fatalf("pending %d", got)
	}
	seen, err := a.Ack(v.ID, "10 dk'ya ararım")
	if err != nil || !seen.Seen || seen.Reply != "10 dk'ya ararım" {
		t.Fatalf("ack: %+v %v", seen, err)
	}
	if p := a.Pending(); len(p) != 1 || p[0].Kind != Moral {
		t.Fatalf("pending after ack: %+v", p)
	}
	if _, err := a.Ack(v.ID, ""); !errors.Is(err, ErrAlertNotFound) {
		t.Fatalf("second ack should fail, got %v", err)
	}
}
