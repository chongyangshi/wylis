package outgoing

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/icydoge/wylis/config"

	"github.com/monzo/slog"
	"github.com/monzo/typhon"
)

var outgoingInterval time.Duration

func initTyphonClient(ctx context.Context) error {
	var err error
	outgoingInterval, err = time.ParseDuration(config.ConfigOutgoingInterval)
	if err != nil {
		slog.Error(ctx, "Error parsing outgoing interval from environmental config: %v", err)
		return err
	}

	timeOut, err := time.ParseDuration(config.ConfigOutgoingTimeout)
	if err != nil {
		slog.Error(ctx, "Error parsing timeout interval from environmental config: %v", err)
		return err
	}

	slog.Info(ctx, "Outgoing request interval: %v, outgoing request timeout: %v.", outgoingInterval, timeOut)

	// Do not reuse HTTP connections, this is less efficient but ensures that
	// middlebox kernel always observes new TCP connections going through,
	// keeping routing fresh.
	roundTripper := &http.Transport{
		DisableKeepAlives:  true,
		DisableCompression: false,
		DialContext: (&net.Dialer{
			Timeout:   timeOut,
			KeepAlive: -1 * time.Second, // Disabled
			DualStack: true,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		IdleConnTimeout:       60 * time.Second,
		TLSHandshakeTimeout:   timeOut,
		ExpectContinueTimeout: 1 * time.Second,
	}

	outgoingClient := func(req typhon.Request) typhon.Response {
		return typhon.HttpService(roundTripper)(req)
	}

	typhon.Client = outgoingClient

	return nil
}
