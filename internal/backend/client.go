// Package backend talks to dbup-backend: pairing, sending events, quick
// replies and the live WebSocket (docs/api.md in dbup-backend).
package backend

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Horozlar-Pau/dbup-agent/internal/events"
)

const (
	requestTimeout = 15 * time.Second
	maxResponse    = 1 << 20
	// MaxBatch is the most events the backend accepts in one request.
	MaxBatch = 100
)

// ErrUnauthorized means the device token was rejected (removed or wrong).
var ErrUnauthorized = errors.New("device token rejected by backend")

// APIError is an error envelope returned by the backend.
type APIError struct {
	Status  int
	Code    string
	Message string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("backend: %s (%d): %s", e.Code, e.Status, e.Message)
}

// Client calls dbup-backend with a device token.
type Client struct {
	base  *url.URL
	token string
	http  *http.Client
}

// New returns a client for baseURL. token may be empty for Pair.
func New(baseURL, token string) (*Client, error) {
	u, err := url.Parse(strings.TrimRight(baseURL, "/"))
	if err != nil {
		return nil, fmt.Errorf("parse backend URL: %w", err)
	}
	local := u.Hostname() == "localhost" || u.Hostname() == "127.0.0.1"
	if u.Scheme != "https" && !(u.Scheme == "http" && local) {
		return nil, fmt.Errorf("backend URL must use https (http only for localhost): %q", baseURL)
	}
	return &Client{base: u, token: token, http: &http.Client{Timeout: requestTimeout}}, nil
}

type envelope struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
	Error   *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// do sends a JSON request and decodes the envelope's data into out (may be nil).
func (c *Client) do(ctx context.Context, method, path string, in, out any) error {
	var body io.Reader
	if in != nil {
		b, err := json.Marshal(in)
		if err != nil {
			return err
		}
		body = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.base.JoinPath(path).String(), body)
	if err != nil {
		return err
	}
	req.Header.Set("accept", "application/json")
	if in != nil {
		req.Header.Set("content-type", "application/json")
	}
	if c.token != "" {
		req.Header.Set("authorization", "Bearer "+c.token)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("reach backend: %w", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxResponse))
	if err != nil {
		return fmt.Errorf("read backend answer: %w", err)
	}
	var env envelope
	if err := json.Unmarshal(data, &env); err != nil {
		return fmt.Errorf("backend answered %s with no JSON", resp.Status)
	}
	if !env.Success {
		if resp.StatusCode == http.StatusUnauthorized && c.token != "" {
			return ErrUnauthorized
		}
		e := &APIError{Status: resp.StatusCode}
		if env.Error != nil {
			e.Code, e.Message = env.Error.Code, env.Error.Message
		}
		return e
	}
	if out == nil || len(env.Data) == 0 {
		return nil
	}
	return json.Unmarshal(env.Data, out)
}

// Pairing is the result of Pair. Token is shown only once by the backend.
type Pairing struct {
	Token    string `json:"token"`
	DeviceID string `json:"deviceId"`
	Username string `json:"username"`
}

// Pair exchanges the owner's username and password for a device token.
func (c *Client) Pair(ctx context.Context, username, password, deviceName string) (Pairing, error) {
	var p Pairing
	err := c.do(ctx, http.MethodPost, "/v1/agent/pair", map[string]string{
		"username": username, "password": password, "deviceName": deviceName,
	}, &p)
	if err != nil {
		return Pairing{}, err
	}
	if p.Token == "" {
		return Pairing{}, errors.New("backend returned no token")
	}
	return p, nil
}

type eventIn struct {
	Type         string    `json:"type"`
	OccurredAt   time.Time `json:"occurredAt"`
	HumanSummary string    `json:"humanSummary,omitempty"`
}

// SendEvents posts up to MaxBatch events.
func (c *Client) SendEvents(ctx context.Context, list []events.Event) error {
	if len(list) > MaxBatch {
		return fmt.Errorf("at most %d events per request", MaxBatch)
	}
	in := make([]eventIn, len(list))
	for i, e := range list {
		in[i] = eventIn{Type: e.Type, OccurredAt: e.At, HumanSummary: e.HumanSummary}
	}
	return c.do(ctx, http.MethodPost, "/v1/agent/events", map[string]any{"events": in}, nil)
}

// Reply sends a quick reply to a nudge.
func (c *Client) Reply(ctx context.Context, nudgeID, reply string) error {
	return c.do(ctx, http.MethodPost, "/v1/agent/nudges/"+url.PathEscape(nudgeID)+"/reply", map[string]string{"reply": reply}, nil)
}

// VIPSeen reports that the developer pressed "Gördüm" on an urgent alert.
func (c *Client) VIPSeen(ctx context.Context, vipID string) error {
	return c.do(ctx, http.MethodPost, "/v1/agent/vip/"+url.PathEscape(vipID)+"/seen", nil, nil)
}
