package daemon

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Horozlar-Pau/dbup-agent/internal/backend"
	"github.com/Horozlar-Pau/dbup-agent/internal/ipc"
)

// Request types for messages to the close circle.
const (
	TypeCircle      = "circle"       // no data; answers []backend.Member
	TypeMessageSend = "message.send" // data: MessageSend; answers backend.SentMessage
)

// MessageSend is a message to one person in the close circle.
type MessageSend struct {
	MemberID string `json:"memberId"`
	Text     string `json:"text"`
}

// Messenger is the part of dbup-backend used for messages.
type Messenger interface {
	Circle(ctx context.Context) ([]backend.Member, error)
	SendMessage(ctx context.Context, memberID, text string) (backend.SentMessage, error)
}

var errNotPaired = errors.New("bu bilgisayar sunucuya bağlı değil (dbup pair)")

// SetMessenger lets the developer send messages from this computer.
func (d *Daemon) SetMessenger(m Messenger) {
	d.messenger = m
}

func (d *Daemon) handleMessages(req ipc.Request) ipc.Response {
	if d.messenger == nil {
		return ipc.ErrorResponse(errNotPaired)
	}
	ctx, cancel := context.WithTimeout(context.Background(), callTimeout)
	defer cancel()
	if req.Type == TypeCircle {
		return result(d.messenger.Circle(ctx))
	}
	var m MessageSend
	if err := json.Unmarshal(req.Data, &m); err != nil {
		return ipc.ErrorResponse(fmt.Errorf("bad message: %w", err))
	}
	return result(d.messenger.SendMessage(ctx, m.MemberID, m.Text))
}
