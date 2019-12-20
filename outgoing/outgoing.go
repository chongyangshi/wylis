package outgoing

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/monzo/slog"
	"github.com/monzo/typhon"
	"golang.org/x/sync/errgroup"

	"github.com/icydoge/wylis/config"
	"github.com/icydoge/wylis/metrics"
)

func initOutgoing(ctx context.Context) error {
	interval, err := time.ParseDuration(config.ConfigOutgoingInterval)
	if err != nil {
		slog.Error(ctx, "Failed to parse neighbour outgoing interval %s: %v", config.ConfigOutgoingInterval, err)
		return err
	}

	outgoingTicker := time.NewTicker(interval)
	outgoingQuit := make(chan struct{})
	go func() {
		for {
			select {
			case <-outgoingTicker.C:
				g, ctx := errgroup.WithContext(ctx)
				for _, neighbourIP := range getNeighbourPods() {
					neighbourIP := neighbourIP // Avoids shadowing
					g.Go(func() error {
						err := sendOutgoing(ctx, neighbourIP)
						return err
					})
				}
				if err := g.Wait(); err != nil {
					slog.Error(ctx, "Error sending outgoing to at least one neighbour: %v", err)
				}

			case <-outgoingQuit:
				outgoingTicker.Stop()
				return
			}
		}
	}()

	return nil
}

func sendOutgoing(ctx context.Context, targetIP string) error {
	req := typhon.NewRequest(ctx, http.MethodGet, fmt.Sprintf("http://%s/incoming", targetIP), nil)

	requestStart := time.Now()
	rsp := req.Send().Response()
	requestDuration := time.Now().Sub(requestStart)

	if rsp.StatusCode >= 400 {
		metrics.RegisterOutgoingRequest(targetIP, false)
		return rsp.Error
	}

	// We do not time failed requests, as it could have been a timeout error
	metrics.RegisterOutgoingRequest(targetIP, true)
	metrics.RegisterOutgoingTiming(targetIP, requestDuration.Seconds())

	return nil
}
