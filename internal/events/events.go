// Package events defines what the agent reports. Names match dbup-backend's api.md.
//
// An Event carries only its type, time and source: never prompt text, code,
// file names or folder names (see docs/privacy.md).
package events

import (
	"fmt"
	"sync"
	"time"
)

// Event types.
const (
	PromptStarted  = "ai.prompt_started"
	PromptFinished = "ai.prompt_finished"
	SessionEnded   = "ai.session_ended"
	GitCommit      = "git.commit"
)

// Sources.
const (
	SourceClaude = "claude"
	SourceGit    = "git"
)

var known = map[string]bool{
	PromptStarted:  true,
	PromptFinished: true,
	SessionEnded:   true,
	GitCommit:      true,
}

// Event is one thing that happened on the developer's machine.
type Event struct {
	Type   string    `json:"type"`
	At     time.Time `json:"at"`
	Source string    `json:"source"`
}

// Validate reports whether e can be accepted.
func (e Event) Validate() error {
	if !known[e.Type] {
		return fmt.Errorf("unknown event type %q", e.Type)
	}
	if e.At.IsZero() {
		return fmt.Errorf("event %s has no time", e.Type)
	}
	return nil
}

// FromClaudeHook maps a Claude Code hook name (hook_event_name) to an event type.
// ok is false for hooks dbup does not track.
func FromClaudeHook(hookEventName string) (eventType string, ok bool) {
	switch hookEventName {
	case "UserPromptSubmit":
		return PromptStarted, true
	case "Stop":
		return PromptFinished, true
	case "SessionEnd":
		return SessionEnded, true
	}
	return "", false
}

// Log keeps the most recent events in memory, newest last.
type Log struct {
	mu   sync.Mutex
	max  int
	list []Event
}

// NewLog returns a Log that holds up to max events.
func NewLog(max int) *Log {
	return &Log{max: max}
}

// Add appends e, dropping the oldest event when full.
func (l *Log) Add(e Event) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.list = append(l.list, e)
	if over := len(l.list) - l.max; over > 0 {
		l.list = append([]Event(nil), l.list[over:]...)
	}
}

// Recent returns up to n of the newest events, oldest first. n <= 0 returns all.
func (l *Log) Recent(n int) []Event {
	l.mu.Lock()
	defer l.mu.Unlock()
	start := 0
	if n > 0 && n < len(l.list) {
		start = len(l.list) - n
	}
	return append([]Event{}, l.list[start:]...)
}
