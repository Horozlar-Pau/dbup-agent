package backend

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/coder/websocket"
)

// Message is one message from WS /v1/agent/live.
type Message struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// Live message types the agent handles.
const (
	MsgNudges = "nudges.release"
	MsgVIP    = "vip.alert"
	MsgCheer  = "cheer.received"
)

// Nudge is a nudge released to the agent.
type Nudge struct {
	ID    string `json:"id"`
	Kind  string `json:"kind"`
	Label string `json:"label"`
	Note  string `json:"note"`
	From  string `json:"from"`
}

// NudgesData is the data of nudges.release.
type NudgesData struct {
	Nudges []Nudge `json:"nudges"`
}

// VIPData is the data of vip.alert.
type VIPData struct {
	ID   string `json:"id"`
	From string `json:"from"`
	Note string `json:"note"`
}

// CheerData is the data of cheer.received.
type CheerData struct {
	From     string `json:"from"`
	Message  string `json:"message"`
	Confetti bool   `json:"confetti"`
}

const (
	retryMin    = 2 * time.Second
	retryMax    = time.Minute
	maxMessage  = 1 << 20
	dialTimeout = 15 * time.Second
)

// Live keeps WS /v1/agent/live open until ctx ends, reconnecting with backoff,
// and calls handle for each message. It returns ErrUnauthorized when the
// token is rejected, since retrying cannot help then.
func (c *Client) Live(ctx context.Context, handle func(Message), logf func(string, ...any)) error {
	wait := retryMin
	for {
		err := c.liveOnce(ctx, handle, func() { wait = retryMin })
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if errors.Is(err, ErrUnauthorized) {
			return err
		}
		logf("backend live connection lost: %v (retrying in %s)", err, wait)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(wait):
		}
		wait = min(wait*2, retryMax)
	}
}

func (c *Client) liveOnce(ctx context.Context, handle func(Message), connected func()) error {
	u := *c.base.JoinPath("/v1/agent/live")
	if u.Scheme == "https" {
		u.Scheme = "wss"
	} else {
		u.Scheme = "ws"
	}
	dctx, cancel := context.WithTimeout(ctx, dialTimeout)
	conn, resp, err := websocket.Dial(dctx, u.String(), &websocket.DialOptions{
		HTTPHeader: http.Header{"Authorization": {"Bearer " + c.token}},
	})
	cancel()
	if resp != nil && resp.StatusCode == http.StatusUnauthorized {
		return ErrUnauthorized
	}
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer conn.CloseNow()
	conn.SetReadLimit(maxMessage)
	connected()

	for {
		_, data, err := conn.Read(ctx)
		if err != nil {
			if websocket.CloseStatus(err) == websocket.StatusPolicyViolation {
				return ErrUnauthorized // device removed
			}
			return err
		}
		var m Message
		if err := json.Unmarshal(data, &m); err != nil {
			continue
		}
		handle(m)
	}
}
