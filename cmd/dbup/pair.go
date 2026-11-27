package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/Horozlar-Pau/dbup-agent/internal/backend"
	"github.com/Horozlar-Pau/dbup-agent/internal/config"
)

func pairCmd() *cobra.Command {
	var username, device string
	cmd := &cobra.Command{
		Use:   "pair",
		Short: "Connect this computer to your dbup server (asks for your password)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			if username == "" {
				if username, err = prompt("Kullanıcı adı: "); err != nil {
					return err
				}
			}
			fmt.Fprint(os.Stderr, "Şifre: ")
			pw, err := term.ReadPassword(int(os.Stdin.Fd()))
			fmt.Fprintln(os.Stderr)
			if err != nil {
				return fmt.Errorf("read password: %w", err)
			}
			if device == "" {
				device, _ = os.Hostname()
			}
			c, err := backend.New(cfg.BackendURL, "")
			if err != nil {
				return err
			}
			p, err := c.Pair(cmd.Context(), strings.TrimSpace(username), string(pw), device)
			if err != nil {
				return fmt.Errorf("pair with %s: %w", cfg.BackendURL, err)
			}
			cfg.Token = p.Token
			if err := config.Save(cfg); err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Eşleşti: %s (%s)\n", p.Username, cfg.BackendURL)
			fmt.Fprintln(out, "dbupd'yi yeniden başlatın: bağlantı açılışta kurulur.")
			return nil
		},
	}
	cmd.Flags().StringVarP(&username, "username", "u", "", "your dbup username")
	cmd.Flags().StringVar(&device, "device", "", "name shown for this computer (default: hostname)")
	return cmd
}

func prompt(label string) (string, error) {
	fmt.Fprint(os.Stderr, label)
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil && line == "" {
		return "", errors.New("no input")
	}
	return strings.TrimSpace(line), nil
}
