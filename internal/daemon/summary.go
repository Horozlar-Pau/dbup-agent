package daemon

import (
	"context"
	"time"

	"github.com/Horozlar-Pau/dbup-agent/internal/events"
	"github.com/Horozlar-Pau/dbup-agent/internal/summary"
)

// summaryTimeout bounds one CLIProxy call (docs/summary.md: 8 seconds).
const summaryTimeout = summary.DefaultTimeout

// Summarizer turns a commit subject into a "Ne Yapıyor Bu?" sentence.
type Summarizer interface {
	Summarize(ctx context.Context, commitMessage string) (string, error)
}

// EventRequest is the data of an event request. CommitSubject is set only by
// the git-commit hook when summaries are on. It is used to ask CLIProxy and is
// never stored or sent anywhere else.
type EventRequest struct {
	events.Event
	CommitSubject string `json:"commit_subject,omitempty"`
}

// SetSummarizer turns commit summaries on. A nil s turns them off.
func (d *Daemon) SetSummarizer(s Summarizer) {
	d.summarizer = s
}

func (d *Daemon) cliproxyURL() string {
	if d.summarizer == nil {
		return ""
	}
	if d.cfg.Summary.CLIProxyURL != "" {
		return d.cfg.Summary.CLIProxyURL
	}
	return summary.DefaultURL
}

func (d *Daemon) wantsSummary(req EventRequest) bool {
	return d.summarizer != nil && req.Type == events.GitCommit && req.CommitSubject != ""
}

// withSummary returns e with its summary filled in. If CLIProxy is down or slow,
// e is returned unchanged: the commit is still reported, without a sentence.
func (d *Daemon) withSummary(e events.Event, subject string) events.Event {
	ctx, cancel := context.WithTimeout(context.Background(), summaryTimeout+time.Second)
	defer cancel()
	s, err := d.summarizer.Summarize(ctx, subject)
	if err != nil {
		d.logf("commit summary skipped: %v", err)
		return e
	}
	out := e
	out.HumanSummary = s
	if err := out.Validate(); err != nil {
		d.logf("commit summary dropped: %v", err)
		return e
	}
	return out
}
