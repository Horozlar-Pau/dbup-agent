// Command dbupd is the dbup background service.
package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Horozlar-Pau/dbup-agent/internal/backend"
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	d := daemon.New(cfg)
	setupSummary(d, cfg)
	setupBackend(ctx, d, cfg)
	srv, err := ipc.Listen(sock, d.Handle)
	if err != nil {
		log.Fatal(err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-stop
		log.Print("shutting down")
		cancel()
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

// setupBackend connects to dbup-backend when this computer is paired: events
// are sent in the background and nudges arrive over the live connection.
func setupBackend(ctx context.Context, d *daemon.Daemon, cfg config.Config) {
	if cfg.Token == "" {
		log.Print("not paired with a backend (run `dbup pair`)")
		return
	}
	c, err := backend.New(cfg.BackendURL, cfg.Token)
	if err != nil {
		log.Printf("backend off: %v", err)
		return
	}
	d.SetBackend(c)
	d.SetMessenger(c)
	go d.RunSender(ctx)
	go func() {
		err := c.Live(ctx, d.HandleLive, log.Printf)
		if errors.Is(err, backend.ErrUnauthorized) {
			log.Print("backend rejected this device: run `dbup pair` again")
		}
	}()
	log.Printf("connected to backend %s", cfg.BackendURL)
}
