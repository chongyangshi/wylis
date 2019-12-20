package incoming

import (
	"github.com/monzo/typhon"
)

type healthCheckResponse struct{}

func serveHealthCheck(req typhon.Request) typhon.Response {
	// Returns a plain 200 success response to show that
	// the server is still alive.
	return req.Response(healthCheckResponse{})
}
