// Command dbupd is the dbup background service.
package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Horozlar-Pau/dbup-agent/internal/config"
	"github.com/Horozlar-Pau/dbup-agent/internal/daemon"
	"github.com/Horozlar-Pau/dbup-agent/internal/ipc"
	"github.com/Horozlar-Pau/dbup-agent/internal/keychain"
	"github.com/Horozlar-Pau/dbup-agent/internal/summary"
	"github.com/Horozlar-Pau/dbup-agent/internal/version"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lmsgprefix)
	log.SetPrefix("dbupd: ")

	if _, err := config.EnsureDir(); err != nil {
		log.Fatal(err)
	}
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	sock, err := config.SocketPath()
	if err != nil {
		log.Fatal(err)
	}

	d := daemon.New(cfg)
	setupSummary(d, cfg)
	srv, err := ipc.Listen(sock, d.Handle)
	if err != nil {
		log.Fatal(err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-stop
		log.Print("shutting down")
		srv.Close()
	}()

	log.Printf("%s listening on %s (backend %s)", version.Version, sock, cfg.BackendURL)
	if err := srv.Serve(); err != nil {
		log.Fatal(err)
	}
}

// setupSummary turns commit summaries on when the setting allows it. Any problem
// is logged and summaries stay off: commits are still reported without a sentence.
func setupSummary(d *daemon.Daemon, cfg config.Config) {
	if !cfg.Summary.IsEnabled() {
		log.Print("commit summary off (settings)")
		return
	}
	key, err := keychain.Get(context.Background(), keychain.CLIProxyService)
	if errors.Is(err, keychain.ErrNotFound) {
		log.Printf("no CLIProxy key in Keychain (%s), sending requests without one", keychain.CLIProxyService)
	} else if err != nil {
		log.Printf("commit summary off: %v", err)
		return
	}
	c, err := summary.New(summary.Options{URL: cfg.Summary.CLIProxyURL, APIKey: key, Model: cfg.Summary.Model})
	if err != nil {
		log.Printf("commit summary off: %v", err)
		return
	}
	d.SetSummarizer(c)
}
