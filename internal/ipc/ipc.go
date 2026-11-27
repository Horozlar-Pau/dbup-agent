// Package ipc is the local link between the dbup CLI (and later the UI) and dbupd.
//
// Each connection carries one JSON request line and one JSON response line
// over a Unix socket that only the current user can open.
package ipc

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"sync"
	"time"
)

// Request is a message sent to dbupd.
type Request struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data,omitempty"`
}

// Response is dbupd's answer.
type Response struct {
	OK    bool            `json:"ok"`
	Data  json.RawMessage `json:"data,omitempty"`
	Error string          `json:"error,omitempty"`
}

// Handler answers a single request.
type Handler func(Request) Response

// OKResponse wraps v as a successful response.
func OKResponse(v any) Response {
	b, err := json.Marshal(v)
	if err != nil {
		return ErrorResponse(err)
	}
	return Response{OK: true, Data: b}
}

// ErrorResponse wraps err as a failed response.
func ErrorResponse(err error) Response {
	return Response{OK: false, Error: err.Error()}
}

const ioTimeout = 2 * time.Second

// Server listens on a Unix socket and passes each request to a Handler.
type Server struct {
	ln      net.Listener
	handler Handler
	wg      sync.WaitGroup
}

// Listen opens the socket at path. A stale socket file from a crashed run is removed,
// but an address that another dbupd is still serving is refused.
func Listen(path string, h Handler) (*Server, error) {
	if c, err := net.DialTimeout("unix", path, 200*time.Millisecond); err == nil {
		c.Close()
		return nil, fmt.Errorf("dbupd is already running (%s)", path)
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("remove stale socket: %w", err)
	}
	ln, err := net.Listen("unix", path)
	if err != nil {
		return nil, fmt.Errorf("listen on %s: %w", path, err)
	}
	if err := os.Chmod(path, 0o600); err != nil {
		ln.Close()
		return nil, fmt.Errorf("restrict socket: %w", err)
	}
	return &Server{ln: ln, handler: h}, nil
}

// Serve accepts connections until Close is called.
func (s *Server) Serve() error {
	for {
		c, err := s.ln.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return nil
			}
			return err
		}
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			s.handle(c)
		}()
	}
}

// Close stops accepting, waits for open requests and removes the socket file.
func (s *Server) Close() error {
	err := s.ln.Close()
	s.wg.Wait()
	return err
}

func (s *Server) handle(c net.Conn) {
	defer c.Close()
	_ = c.SetDeadline(time.Now().Add(ioTimeout))

	var resp Response
	line, err := bufio.NewReader(c).ReadBytes('\n')
	if err != nil {
		return
	}
	var req Request
	if err := json.Unmarshal(line, &req); err != nil {
		resp = ErrorResponse(fmt.Errorf("bad request: %w", err))
	} else {
		resp = s.handler(req)
	}
	_ = json.NewEncoder(c).Encode(resp)
}

// Send delivers one request to the socket at path and returns the response.
func Send(ctx context.Context, path string, req Request) (Response, error) {
	var d net.Dialer
	c, err := d.DialContext(ctx, "unix", path)
	if err != nil {
		return Response{}, fmt.Errorf("connect to dbupd: %w", err)
	}
	defer c.Close()

	deadline := time.Now().Add(ioTimeout)
	if dl, ok := ctx.Deadline(); ok && dl.Before(deadline) {
		deadline = dl
	}
	_ = c.SetDeadline(deadline)

	if err := json.NewEncoder(c).Encode(req); err != nil {
		return Response{}, fmt.Errorf("send request: %w", err)
	}
	var resp Response
	if err := json.NewDecoder(c).Decode(&resp); err != nil {
		return Response{}, fmt.Errorf("read response: %w", err)
	}
	return resp, nil
}
