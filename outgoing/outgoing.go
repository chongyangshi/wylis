package outgoing

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/monzo/slog"
	"github.com/monzo/typhon"
	"golang.org/x/sync/errgroup"

	"github.com/icydoge/wylis/config"
	"github.com/icydoge/wylis/metrics"
)

var targetPort int

func initOutgoing(ctx context.Context) error {
	interval, err := time.ParseDuration(config.ConfigOutgoingInterval)
	if err != nil {
		slog.Error(ctx, "Failed to parse neighbour outgoing interval %s: %v", config.ConfigOutgoingInterval, err)
		return err
	}

	targetPortParsed, err := strconv.ParseInt(config.ConfigIncomingListenPort, 10, 32)
	if err != nil {
		slog.Error(ctx, "Failed to parse target port %s: %v", config.ConfigIncomingListenPort, err)
		return err
	}
	targetPort = int(targetPortParsed)

	// Main outgoing routine
	outgoingTicker := time.NewTicker(interval)
	outgoingQuit := make(chan struct{})
	go func() {
		for {
			select {
			case <-outgoingTicker.C:
				// One round of polling of all neighbours
				g, ctx := errgroup.WithContext(ctx)

				neighbours := getNeighbourPods()
				results := make([]bool, len(neighbours)) // Should be memory-safe

				for i, neighbour := range neighbours {
					i, neighbour := i, neighbour // Avoids shadowing
					g.Go(func() error {
						err := sendOutgoing(ctx, neighbour)
						if err == nil {
							results[i] = true
						}
						return err
					})
				}
				if err := g.Wait(); err != nil {
					slog.Error(ctx, "Error sending outgoing to at least one neighbour: %v", err)
				}

				success := 0
				for _, result := range results {
					if result == true {
						success++
					}
				}

				metrics.RegisterOutgoingStatus(success, len(results))

			case <-outgoingQuit:
				outgoingTicker.Stop()
				return
			}
		}
	}()

	return nil
}

func sendOutgoing(ctx context.Context, target neighbourPod) error {
	req := typhon.NewRequest(ctx, http.MethodGet, fmt.Sprintf("http://%s:%d/incoming", target.podIP, targetPort), nil)
	if req.Err() != nil {
		// If for any reason we fail to construct a valid HTTP request, return error
		return req.Err()
	}

	req.Header.Set(config.SourceNodeIPHeader, config.ConfigNodeIP)

	requestStart := time.Now()
	rsp := req.Send().Response()
	requestDuration := time.Now().Sub(requestStart)

	if rsp.Response == nil || rsp.StatusCode >= 400 {
		metrics.RegisterOutgoingRequest(target.nodeIP, false)
		return rsp.Error
	}

	// We do not time failed requests, as it could have been a timeout error
	metrics.RegisterOutgoingRequest(target.nodeIP, true)
	metrics.RegisterOutgoingTiming(target.nodeIP, requestDuration.Seconds())

	return nil
}
