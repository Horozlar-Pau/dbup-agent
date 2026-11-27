package summary

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
	"unicode/utf8"
)

// fakeProxy answers /v1/messages with text and records what it received.
type fakeProxy struct {
	text  string
	delay time.Duration
	calls atomic.Int32
	last  atomic.Value // request
	key   atomic.Value // string
}

func (f *fakeProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.calls.Add(1)
	if r.URL.Path != "/v1/messages" || r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}
	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	f.last.Store(req)
	f.key.Store(r.Header.Get("x-api-key"))
	if f.delay > 0 {
		select {
		case <-time.After(f.delay):
		case <-r.Context().Done():
			return
		}
	}
	_ = json.NewEncoder(w).Encode(map[string]any{
		"content": []map[string]string{{"type": "text", "text": f.text}},
	})
}

func newTestClient(t *testing.T, f *fakeProxy, timeout time.Duration) *Client {
	t.Helper()
	srv := httptest.NewServer(f)
	t.Cleanup(srv.Close)
	c, err := New(Options{URL: srv.URL, APIKey: "local-key", Timeout: timeout})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestSummarizeSendsOnlySubject(t *testing.T) {
	f := &fakeProxy{text: "Uygulamadaki bir hatayı düzeltti."}
	c := newTestClient(t, f, time.Second)

	got, err := c.Summarize(context.Background(), "fix: login crash\n\nTouches auth/session.go and secret notes")
	if err != nil {
		t.Fatal(err)
	}
	if got != "Uygulamadaki bir hatayı düzeltti." {
		t.Fatalf("got %q", got)
	}
	req := f.last.Load().(request)
	if req.Model != DefaultModel || len(req.Messages) != 1 || req.Messages[0].Content != "fix: login crash" {
		t.Fatalf("sent %+v", req)
	}
	if k := f.key.Load().(string); k != "local-key" {
		t.Fatalf("api key header = %q", k)
	}
}

func TestSummarizeCachesSameSubject(t *testing.T) {
	f := &fakeProxy{text: "Bir şeyi düzeltti."}
	c := newTestClient(t, f, time.Second)

	for range 3 {
		if _, err := c.Summarize(context.Background(), "fix: same thing"); err != nil {
			t.Fatal(err)
		}
	}
	if n := f.calls.Load(); n != 1 {
		t.Fatalf("CLIProxy called %d times, want 1", n)
	}
}

func TestSummarizeLimitsLengths(t *testing.T) {
	f := &fakeProxy{text: strings.Repeat("çok uzun bir cümle ", 20)}
	c := newTestClient(t, f, time.Second)

	got, err := c.Summarize(context.Background(), strings.Repeat("ş", 500))
	if err != nil {
		t.Fatal(err)
	}
	if n := utf8.RuneCountInString(got); n > MaxSummary {
		t.Fatalf("summary has %d characters, want <= %d", n, MaxSummary)
	}
	if !strings.HasSuffix(got, "…") {
		t.Fatalf("shortened summary should end with an ellipsis: %q", got)
	}
	sent := f.last.Load().(request).Messages[0].Content
	if n := utf8.RuneCountInString(sent); n != MaxInput {
		t.Fatalf("sent %d characters, want %d", n, MaxInput)
	}
}

func TestSummarizeTimesOut(t *testing.T) {
	f := &fakeProxy{text: "geç", delay: time.Second}
	c := newTestClient(t, f, 50*time.Millisecond)

	if _, err := c.Summarize(context.Background(), "feat: slow"); err == nil {
		t.Fatal("want timeout error")
	}
}

func TestSummarizeProxyDown(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	url := srv.URL
	srv.Close()
	c, err := New(Options{URL: url, Timeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.Summarize(context.Background(), "feat: x"); err == nil {
		t.Fatal("want error when CLIProxy is down")
	}
}

func TestSummarizeErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "no", http.StatusUnauthorized)
	}))
	defer srv.Close()
	c, _ := New(Options{URL: srv.URL})
	if _, err := c.Summarize(context.Background(), "feat: x"); err == nil {
		t.Fatal("want error on 401")
	}
}

func TestSummarizeEmptyMessage(t *testing.T) {
	c, _ := New(Options{})
	if _, err := c.Summarize(context.Background(), " \n\n "); err == nil {
		t.Fatal("want error for empty message")
	}
}

func TestNewRejectsRemoteAddress(t *testing.T) {
	for _, raw := range []string{
		"https://api.anthropic.com",
		"http://192.168.1.10:8317",
		"http://127.0.0.1.evil.com:8317",
		"ftp://127.0.0.1:8317",
		"127.0.0.1:8317",
	} {
		if _, err := New(Options{URL: raw}); err == nil {
			t.Errorf("%s: want error", raw)
		}
	}
	if _, err := New(Options{URL: "http://example.com"}); !errors.Is(err, ErrNotLocal) {
		t.Fatalf("want ErrNotLocal, got %v", err)
	}
}

func TestNewAcceptsLoopback(t *testing.T) {
	for _, raw := range []string{"http://127.0.0.1:8317", "http://localhost:8317", "http://[::1]:8317", ""} {
		if _, err := New(Options{URL: raw}); err != nil {
			t.Errorf("%q: %v", raw, err)
		}
	}
}

func TestRedirectIsNotFollowed(t *testing.T) {
	var remoteHit atomic.Bool
	remote := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { remoteHit.Store(true) }))
	defer remote.Close()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, remote.URL, http.StatusTemporaryRedirect)
	}))
	defer srv.Close()

	c, _ := New(Options{URL: srv.URL})
	if _, err := c.Summarize(context.Background(), "feat: x"); err == nil {
		t.Fatal("want error on redirect")
	}
	if remoteHit.Load() {
		t.Fatal("redirect was followed")
	}
}

func TestSubject(t *testing.T) {
	cases := map[string]string{
		"feat: a\nbody":     "feat: a",
		"\n\n  fix: b  \nx": "fix: b",
		"":                  "",
	}
	for in, want := range cases {
		if got := Subject(in); got != want {
			t.Errorf("Subject(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestCleanStripsQuotesAndNewlines(t *testing.T) {
	if got := clean("  \"Bir hatayı\n düzeltti.\"  "); got != "Bir hatayı düzeltti." {
		t.Fatalf("got %q", got)
	}
}

func TestCacheDropsOldest(t *testing.T) {
	c := newCache(2)
	c.put("a", "1")
	c.put("b", "2")
	c.put("c", "3")
	if _, ok := c.get("a"); ok {
		t.Fatal("oldest entry should be dropped")
	}
	if v, ok := c.get("c"); !ok || v != "3" {
		t.Fatal("newest entry missing")
	}
}
