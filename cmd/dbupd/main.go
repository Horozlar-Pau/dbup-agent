// Command dbupd is the dbup background service.
package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Horozlar-Pau/dbup-agent/internal/config"
	"github.com/Horozlar-Pau/dbup-agent/internal/daemon"
	"github.com/Horozlar-Pau/dbup-agent/internal/ipc"
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
