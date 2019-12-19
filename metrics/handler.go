package metrics

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/monzo/slog"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/icydoge/wylis/config"
)

const (
	recentRestartsWindow = time.Minute * 10
	maxRecentRestarts    = 5
)

var (
	lastRestartTime time.Time
	recentRestarts  int
)

func Init() {
	ctx := context.Background()

	http.Handle("/metrics", promhttp.Handler())

	// A simple automatic recovery routine for the metrics server with limited recent retries
	go func() {
		for {
			err := http.ListenAndServe(fmt.Sprintf("%s:%s", config.ConfigListenAddr, config.ConfigMetricsListenPort), nil)
			if err != nil {
				slog.Error(ctx, "Local metrics server encountered error: %v", err)

				timeOfError := time.Now()

				if timeOfError.Sub(lastRestartTime) > recentRestartsWindow {
					recentRestarts = 0
				}

				if recentRestarts > maxRecentRestarts {
					slog.Critical(ctx, "Too many recent restarts (%d), exiting.", maxRecentRestarts)
					break
				}

				slog.Warn(ctx, "Restaring metrics server following recent error %v", err)
				recentRestarts++
			}
		}
	}()
}
