// Package hook turns a Claude Code or git hook call into an event.
//
// Hooks run inside the developer's tools, so this code must be fast and must
// never fail loudly: the caller always exits 0 and prints nothing to stdout
// (Claude Code adds a UserPromptSubmit hook's stdout to the prompt).
package hook

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/Horozlar-Pau/dbup-agent/internal/events"
)

// Hook names accepted by `dbup hook <name>`.
const (
	Claude    = "claude"
	GitCommit = "git-commit"
)

// Names lists the accepted hook names.
var Names = []string{Claude, GitCommit}

// ErrIgnored means the hook call is valid but produces no event.
var ErrIgnored = errors.New("hook ignored")

// maxInput caps how much of the hook's stdin is read. A prompt can be long, and
// only hook_event_name is needed.
const maxInput = 1 << 20

// claudeInput is the part of Claude Code's hook JSON dbup reads. Everything
// else, including the prompt text, is skipped by the decoder and never stored.
type claudeInput struct {
	HookEventName string `json:"hook_event_name"`
}

// Event builds the event for hook name. stdin is read only for Claude hooks.
func Event(name string, stdin io.Reader, now time.Time) (events.Event, error) {
	switch name {
	case Claude:
		if stdin == nil {
			return events.Event{}, fmt.Errorf("%w: no hook input", ErrIgnored)
		}
		var in claudeInput
		if err := json.NewDecoder(io.LimitReader(stdin, maxInput)).Decode(&in); err != nil {
			return events.Event{}, fmt.Errorf("read Claude hook input: %w", err)
		}
		t, ok := events.FromClaudeHook(in.HookEventName)
		if !ok {
			return events.Event{}, fmt.Errorf("%w: Claude hook %q", ErrIgnored, in.HookEventName)
		}
		return events.Event{Type: t, At: now, Source: events.SourceClaude}, nil
	case GitCommit:
		return events.Event{Type: events.GitCommit, At: now, Source: events.SourceGit}, nil
	default:
		return events.Event{}, fmt.Errorf("unknown hook %q (want one of %v)", name, Names)
	}
}
