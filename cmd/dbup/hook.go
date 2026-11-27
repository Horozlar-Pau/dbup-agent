package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/Horozlar-Pau/dbup-agent/internal/config"
	"github.com/Horozlar-Pau/dbup-agent/internal/daemon"
	"github.com/Horozlar-Pau/dbup-agent/internal/hook"
	"github.com/Horozlar-Pau/dbup-agent/internal/ipc"
)

// hookTimeout bounds how long a hook may wait for dbupd. If dbupd is slow or
// not running the event is dropped: the developer's tool must not wait.
const hookTimeout = 300 * time.Millisecond

func hookCmd() *cobra.Command {
	return &cobra.Command{
		Use:       "hook <claude|git-commit>",
		Short:     "Report a Claude Code or git hook to dbupd (called by hooks, not by people)",
		Args:      cobra.ExactArgs(1),
		ValidArgs: hook.Names,
		// A hook never fails: errors are shown only with DBUP_DEBUG=1.
		Run: func(cmd *cobra.Command, args []string) {
			if err := runHook(args[0]); err != nil && os.Getenv("DBUP_DEBUG") != "" {
				fmt.Fprintln(os.Stderr, "dbup hook:", err)
			}
		},
	}
}

func runHook(name string) error {
	var stdin io.Reader
	if name == hook.Claude && !isTerminal(os.Stdin) {
		stdin = os.Stdin
	}
	e, err := hook.Event(name, stdin, time.Now().UTC())
	if err != nil {
		return err
	}
	data, err := json.Marshal(e)
	if err != nil {
		return err
	}
	sock, err := config.SocketPath()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), hookTimeout)
	defer cancel()
	resp, err := ipc.Send(ctx, sock, ipc.Request{Type: daemon.TypeEvent, Data: data})
	if err != nil {
		return err
	}
	if !resp.OK {
		return fmt.Errorf("dbupd: %s", resp.Error)
	}
	return nil
}

// isTerminal reports whether f is an interactive terminal, so a hook run by
// hand does not block waiting for input.
func isTerminal(f *os.File) bool {
	fi, err := f.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}
