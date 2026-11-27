package ipc

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// socketPath returns a short path: macOS limits Unix socket paths to ~104 bytes,
// and t.TempDir() can exceed that.
func socketPath(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("/tmp", "dbup")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	return filepath.Join(dir, "s.sock")
}

func startServer(t *testing.T, path string, h Handler) *Server {
	t.Helper()
	srv, err := Listen(path, h)
	if err != nil {
		t.Fatalf("Listen: %v", err)
	}
	go srv.Serve()
	t.Cleanup(func() { srv.Close() })
	return srv
}

func TestRoundTrip(t *testing.T) {
	path := socketPath(t)
	startServer(t, path, func(r Request) Response {
		return OKResponse(map[string]string{"echo": r.Type})
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	resp, err := Send(ctx, path, Request{Type: "ping"})
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
	var got map[string]string
	if err := json.Unmarshal(resp.Data, &got); err != nil {
		t.Fatal(err)
	}
	if !resp.OK || got["echo"] != "ping" {
		t.Fatalf("got %+v", resp)
	}
}

func TestSocketIsPrivate(t *testing.T) {
	path := socketPath(t)
	startServer(t, path, func(Request) Response { return Response{OK: true} })

	fi, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if perm := fi.Mode().Perm(); perm != 0o600 {
		t.Fatalf("socket mode = %o, want 600", perm)
	}
}

func TestListenRefusesRunningServer(t *testing.T) {
	path := socketPath(t)
	startServer(t, path, func(Request) Response { return Response{OK: true} })

	if _, err := Listen(path, nil); err == nil {
		t.Fatal("second Listen on a live socket should fail")
	}
}

func TestListenReplacesStaleSocket(t *testing.T) {
	path := socketPath(t)
	if err := os.WriteFile(path, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	startServer(t, path, func(Request) Response { return Response{OK: true} })

	if _, err := Send(context.Background(), path, Request{Type: "x"}); err != nil {
		t.Fatalf("Send after stale socket: %v", err)
	}
}

func TestSendWithoutServer(t *testing.T) {
	if _, err := Send(context.Background(), socketPath(t), Request{Type: "x"}); err == nil {
		t.Fatal("want error when nothing is listening")
	}
}
