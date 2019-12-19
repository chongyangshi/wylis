package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/monzo/slog"
	"github.com/monzo/typhon"

	"github.com/icydoge/wylis/config"
	"github.com/icydoge/wylis/incoming"
)

func main() {
	initContext := context.Background()
	svc := incoming.Service()
	srv, err := typhon.Listen(svc, fmt.Sprintf("%s:%s", config.ListenAddr, config.ConfigIncomingListenPort))
	if err != nil {
		panic(err)
	}
	slog.Info(initContext, "Wylis incoming listening on %v", srv.Listener().Addr())

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	<-done
	slog.Info(initContext, "Wylis shutting down")
	c, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Stop(c)
}
