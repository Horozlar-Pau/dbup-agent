package main

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Horozlar-Pau/dbup-agent/internal/daemon"
	"github.com/Horozlar-Pau/dbup-agent/internal/events"
	"github.com/Horozlar-Pau/dbup-agent/internal/ipc"
)

func eventsCmd() *cobra.Command {
	var limit int
	cmd := &cobra.Command{
		Use:   "events",
		Short: "List the most recent events dbupd has received",
		RunE: func(cmd *cobra.Command, _ []string) error {
			q, err := json.Marshal(daemon.EventsQuery{Limit: limit})
			if err != nil {
				return err
			}
			resp, err := call(cmd.Context(), ipc.Request{Type: daemon.TypeEvents, Data: q})
			if err != nil {
				return err
			}
			var list []events.Event
			if err := json.Unmarshal(resp.Data, &list); err != nil {
				return fmt.Errorf("read events: %w", err)
			}
			out := cmd.OutOrStdout()
			if len(list) == 0 {
				fmt.Fprintln(out, "no events yet")
				return nil
			}
			for _, e := range list {
				line := fmt.Sprintf("%s  %-6s  %s", e.At.Local().Format("15:04:05"), e.Source, e.Type)
				if e.HumanSummary != "" {
					line += "  \"" + e.HumanSummary + "\""
				}
				fmt.Fprintln(out, line)
			}
			return nil
		},
	}
	cmd.Flags().IntVarP(&limit, "limit", "n", 20, "how many events to show (0 = all)")
	return cmd
}
