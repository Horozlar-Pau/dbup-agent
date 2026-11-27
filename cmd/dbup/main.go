// Command dbup is the CLI that people and hooks call. It talks to dbupd.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/Horozlar-Pau/dbup-agent/internal/config"
	"github.com/Horozlar-Pau/dbup-agent/internal/daemon"
	"github.com/Horozlar-Pau/dbup-agent/internal/ipc"
	"github.com/Horozlar-Pau/dbup-agent/internal/version"
)

func main() {
	root := &cobra.Command{
		Use:           "dbup",
		Short:         "DoNotBreakUp agent CLI",
		Version:       version.Version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.AddCommand(statusCmd())

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "dbup:", err)
		os.Exit(1)
	}
}

func statusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show whether dbupd is running and how it is configured",
		RunE: func(cmd *cobra.Command, _ []string) error {
			resp, err := call(cmd.Context(), ipc.Request{Type: daemon.TypeStatus})
			if err != nil {
				return err
			}
			var st daemon.Status
			if err := json.Unmarshal(resp.Data, &st); err != nil {
				return fmt.Errorf("read status: %w", err)
			}
			token := "not set"
			if st.TokenSet {
				token = "set"
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "dbupd:   running (%s, up %s)\n", st.Version, time.Duration(st.UptimeSeconds)*time.Second)
			fmt.Fprintf(out, "backend: %s\n", st.BackendURL)
			fmt.Fprintf(out, "token:   %s\n", token)
			return nil
		},
	}
}

// call sends one request to dbupd and turns a failed response into an error.
func call(ctx context.Context, req ipc.Request) (ipc.Response, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	sock, err := config.SocketPath()
	if err != nil {
		return ipc.Response{}, err
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	resp, err := ipc.Send(ctx, sock, req)
	if err != nil {
		return resp, fmt.Errorf("cannot reach dbupd, is it running? (start it with `dbupd`): %w", err)
	}
	if !resp.OK {
		return resp, errors.New(resp.Error)
	}
	return resp, nil
}
