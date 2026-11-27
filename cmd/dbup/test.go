package main

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Horozlar-Pau/dbup-agent/internal/daemon"
	"github.com/Horozlar-Pau/dbup-agent/internal/inbox"
	"github.com/Horozlar-Pau/dbup-agent/internal/ipc"
)

// testCmd fakes what dbup-backend will send, so the UI can be tried without it.
func testCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test",
		Short: "Send fake close-circle messages to dbupd (until the backend is connected)",
	}
	cmd.AddCommand(testNudgeCmd())
	return cmd
}

func testNudgeCmd() *cobra.Command {
	var n inbox.NewNudge
	cmd := &cobra.Command{
		Use:     "nudge",
		Short:   "Queue a nudge; it shows on the wait card when Claude starts working",
		Example: `  dbup test nudge --from Elif --kind "Kahve molası?" --note "Bugün erken çıkabilir misin?"`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			data, err := json.Marshal(n)
			if err != nil {
				return err
			}
			resp, err := call(cmd.Context(), ipc.Request{Type: daemon.TypeNudgeAdd, Data: data})
			if err != nil {
				return err
			}
			var got inbox.Nudge
			if err := json.Unmarshal(resp.Data, &got); err != nil {
				return fmt.Errorf("read nudge: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "queued nudge #%d from %s: %s\n", got.ID, got.From, got.Kind)
			return nil
		},
	}
	cmd.Flags().StringVar(&n.From, "from", "Elif", "sender's name")
	cmd.Flags().StringVar(&n.Kind, "kind", "Kahve molası?", "nudge type, e.g. \"Müsait olunca ara\"")
	cmd.Flags().StringVar(&n.Note, "note", "", "optional note")
	return cmd
}
