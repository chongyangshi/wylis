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

	"github.com/chongyangshi/wylis/config"
	"github.com/chongyangshi/wylis/incoming"
	"github.com/chongyangshi/wylis/metrics"
	"github.com/chongyangshi/wylis/outgoing"
)

func main() {
	initContext := context.Background()

	// Initialise server for incoming requests
	svc := incoming.Service()
	srv, err := typhon.Listen(svc, fmt.Sprintf("%s:%s", config.ConfigListenAddr, config.ConfigIncomingListenPort))
	if err != nil {
		panic(err)
	}
	slog.Info(initContext, "Wylis incoming listening on %v", srv.Listener().Addr())

	// Initialise metcirs server
	metrics.Init()
	slog.Info(initContext, "Wylis metrics listening on %s", fmt.Sprintf("%s:%s", config.ConfigListenAddr, config.ConfigMetricsListenPort))

	// Initialise client for outgoing requests
	err = outgoing.Init(initContext)
	if err != nil {
		panic(err)
	}

	// Log termination gracefully
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	<-done
	slog.Info(initContext, "Wylis shutting down")
	c, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Stop(c)
}
